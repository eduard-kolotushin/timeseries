package timeseries

import (
	"math"
	"time"
)

// Point is a single timestamped observation.
type Point[T any] struct {
	Time  time.Time
	Value T
}

// Series is a univariate timeseries with a strictly ascending unique UTC index.
type Series[T any] struct {
	times  []time.Time
	values []T
}

// New builds a Series from parallel times and values slices.
// Times are normalized to UTC. Duplicates and unsorted times are rejected.
func New[T any](times []time.Time, values []T) (Series[T], error) {
	ts, vs, err := validateIndex(times, values)
	if err != nil {
		return Series[T]{}, err
	}
	return Series[T]{times: ts, values: vs}, nil
}

// FromPoints builds a Series from points in time order.
func FromPoints[T any](points []Point[T]) (Series[T], error) {
	times := make([]time.Time, len(points))
	values := make([]T, len(points))
	for i, p := range points {
		times[i] = p.Time
		values[i] = p.Value
	}
	return New(times, values)
}

// MustNew is like New but panics on error. Intended for tests and fixtures.
func MustNew[T any](times []time.Time, values []T) Series[T] {
	s, err := New(times, values)
	if err != nil {
		panic(err)
	}
	return s
}

// MustFromPoints is like FromPoints but panics on error.
func MustFromPoints[T any](points []Point[T]) Series[T] {
	s, err := FromPoints(points)
	if err != nil {
		panic(err)
	}
	return s
}

// Len returns the number of observations.
func (s Series[T]) Len() int { return len(s.times) }

// Empty reports whether the series has no points.
func (s Series[T]) Empty() bool { return len(s.times) == 0 }

// At returns the point at index i.
func (s Series[T]) At(i int) (Point[T], error) {
	if i < 0 || i >= len(s.times) {
		return Point[T]{}, ErrIndexOutOfRange
	}
	return Point[T]{Time: s.times[i], Value: s.values[i]}, nil
}

// Time returns the timestamp at index i.
func (s Series[T]) Time(i int) (time.Time, error) {
	if i < 0 || i >= len(s.times) {
		return time.Time{}, ErrIndexOutOfRange
	}
	return s.times[i], nil
}

// Value returns the value at index i.
func (s Series[T]) Value(i int) (T, error) {
	if i < 0 || i >= len(s.values) {
		var zero T
		return zero, ErrIndexOutOfRange
	}
	return s.values[i], nil
}

// Times returns a copy of the time index.
func (s Series[T]) Times() []time.Time {
	return cloneSlice(s.times)
}

// Values returns a copy of the values.
func (s Series[T]) Values() []T {
	return cloneSlice(s.values)
}

// Points returns a copy of all points.
func (s Series[T]) Points() []Point[T] {
	out := make([]Point[T], len(s.times))
	for i := range s.times {
		out[i] = Point[T]{Time: s.times[i], Value: s.values[i]}
	}
	return out
}

// Clone returns a deep copy of the series (copied times and values slices).
func (s Series[T]) Clone() Series[T] {
	return Series[T]{
		times:  cloneSlice(s.times),
		values: cloneSlice(s.values),
	}
}

// withValues returns a series that shares s's time index and uses values.
// values must not alias a backing array that will be mutated later in place.
func (s Series[T]) withValues(values []T) Series[T] {
	return Series[T]{times: s.times, values: values}
}

func cloneSlice[T any](s []T) []T {
	if len(s) == 0 {
		return nil
	}
	out := make([]T, len(s))
	copy(out, s)
	return out
}

// Equal reports whether a and b have the same times and values.
// For float64, NaNs compare equal to NaNs.
func Equal[T comparable](a, b Series[T]) bool {
	if a.Len() != b.Len() {
		return false
	}
	for i := range a.times {
		if !a.times[i].Equal(b.times[i]) || a.values[i] != b.values[i] {
			return false
		}
	}
	return true
}

// EqualFloat reports whether two float64 series are equal, treating NaN as equal to NaN.
func EqualFloat(a, b Series[float64]) bool {
	if a.Len() != b.Len() {
		return false
	}
	for i := range a.times {
		if !a.times[i].Equal(b.times[i]) {
			return false
		}
		av, bv := a.values[i], b.values[i]
		if math.IsNaN(av) && math.IsNaN(bv) {
			continue
		}
		if av != bv {
			return false
		}
	}
	return true
}

// IndexOf returns the index of t, or -1 if absent.
func (s Series[T]) IndexOf(t time.Time) int {
	return searchTime(s.times, t)
}
