package timeseries

import (
	"math"
	"time"
)

// JoinHow selects how two series are aligned on time.
type JoinHow int

const (
	// JoinInner keeps timestamps present in both series.
	JoinInner JoinHow = iota
	// JoinOuter keeps the union of timestamps; gaps get the zero value (NaN for float helpers).
	JoinOuter
	// JoinLeft keeps timestamps from the left series.
	JoinLeft
)

// Align aligns a and b onto a common time index according to how.
// Missing positions are filled with zero T. For float64 use AlignFloat.
func Align[T any](a, b Series[T], how JoinHow) (left, right Series[T]) {
	var missing T
	return align(a, b, how, missing)
}

// AlignFloat is like Align but fills gaps with NaN.
func AlignFloat(a, b Series[float64], how JoinHow) (left, right Series[float64]) {
	return align(a, b, how, math.NaN())
}

func align[T any](a, b Series[T], how JoinHow, missing T) (left, right Series[T]) {
	switch how {
	case JoinInner:
		return alignInner(a, b)
	case JoinLeft:
		return alignLeft(a, b, missing)
	default:
		return alignOuter(a, b, missing)
	}
}

func alignInner[T any](a, b Series[T]) (Series[T], Series[T]) {
	n := min(a.Len(), b.Len())
	times := make([]time.Time, 0, n)
	lv := make([]T, 0, n)
	rv := make([]T, 0, n)
	i, j := 0, 0
	for i < a.Len() && j < b.Len() {
		switch a.times[i].Compare(b.times[j]) {
		case 0:
			times = append(times, a.times[i])
			lv = append(lv, a.values[i])
			rv = append(rv, b.values[j])
			i++
			j++
		case -1:
			i++
		default:
			j++
		}
	}
	return Series[T]{times: times, values: lv}, Series[T]{times: times, values: rv}
}

func alignLeft[T any](a, b Series[T], missing T) (Series[T], Series[T]) {
	rv := make([]T, a.Len())
	j := 0
	for i := 0; i < a.Len(); i++ {
		rv[i] = missing
		for j < b.Len() && b.times[j].Before(a.times[i]) {
			j++
		}
		if j < b.Len() && b.times[j].Equal(a.times[i]) {
			rv[i] = b.values[j]
		}
	}
	return a, Series[T]{times: a.times, values: rv}
}

func alignOuter[T any](a, b Series[T], missing T) (Series[T], Series[T]) {
	capn := a.Len() + b.Len()
	times := make([]time.Time, 0, capn)
	lv := make([]T, 0, capn)
	rv := make([]T, 0, capn)
	i, j := 0, 0
	for i < a.Len() || j < b.Len() {
		switch {
		case j >= b.Len() || (i < a.Len() && a.times[i].Before(b.times[j])):
			times = append(times, a.times[i])
			lv = append(lv, a.values[i])
			rv = append(rv, missing)
			i++
		case i >= a.Len() || (j < b.Len() && b.times[j].Before(a.times[i])):
			times = append(times, b.times[j])
			lv = append(lv, missing)
			rv = append(rv, b.values[j])
			j++
		default:
			times = append(times, a.times[i])
			lv = append(lv, a.values[i])
			rv = append(rv, b.values[j])
			i++
			j++
		}
	}
	return Series[T]{times: times, values: lv}, Series[T]{times: times, values: rv}
}

// ConflictFunc resolves two values at the same timestamp during Merge.
// If nil, Merge returns ErrConflict when both sides have a value at the same time.
type ConflictFunc[T any] func(t time.Time, a, b T) (T, error)

// Merge combines a and b into a sorted union of timestamps.
// When both have the same time, conflict is called; if conflict is nil, ErrConflict is returned.
func Merge[T any](a, b Series[T], conflict ConflictFunc[T]) (Series[T], error) {
	times := make([]time.Time, 0, a.Len()+b.Len())
	values := make([]T, 0, a.Len()+b.Len())
	i, j := 0, 0
	for i < a.Len() || j < b.Len() {
		switch {
		case j >= b.Len() || (i < a.Len() && a.times[i].Before(b.times[j])):
			times = append(times, a.times[i])
			values = append(values, a.values[i])
			i++
		case i >= a.Len() || (j < b.Len() && b.times[j].Before(a.times[i])):
			times = append(times, b.times[j])
			values = append(values, b.values[j])
			j++
		default:
			t := a.times[i]
			if conflict == nil {
				return Series[T]{}, ErrConflict
			}
			v, err := conflict(t, a.values[i], b.values[j])
			if err != nil {
				return Series[T]{}, err
			}
			times = append(times, t)
			values = append(values, v)
			i++
			j++
		}
	}
	return Series[T]{times: times, values: values}, nil
}

// Concat appends b after a. The first timestamp of b must be strictly after the last of a
// (or either series may be empty).
func Concat[T any](a, b Series[T]) (Series[T], error) {
	if a.Empty() {
		return b.Clone(), nil
	}
	if b.Empty() {
		return a.Clone(), nil
	}
	if !b.times[0].After(a.times[a.Len()-1]) {
		if b.times[0].Equal(a.times[a.Len()-1]) {
			return Series[T]{}, ErrDuplicateTime
		}
		return Series[T]{}, ErrUnsorted
	}
	n := a.Len() + b.Len()
	times := make([]time.Time, n)
	values := make([]T, n)
	copy(times, a.times)
	copy(times[a.Len():], b.times)
	copy(values, a.values)
	copy(values[a.Len():], b.values)
	return Series[T]{times: times, values: values}, nil
}
