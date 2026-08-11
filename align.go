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
	switch how {
	case JoinInner:
		return alignInner(a, b)
	case JoinLeft:
		return alignLeft(a, b)
	default:
		return alignOuter(a, b)
	}
}

func alignInner[T any](a, b Series[T]) (Series[T], Series[T]) {
	times := make([]time.Time, 0)
	lv := make([]T, 0)
	rv := make([]T, 0)
	i, j := 0, 0
	for i < a.Len() && j < b.Len() {
		ta, tb := a.times[i], b.times[j]
		switch {
		case ta.Equal(tb):
			times = append(times, ta)
			lv = append(lv, a.values[i])
			rv = append(rv, b.values[j])
			i++
			j++
		case ta.Before(tb):
			i++
		default:
			j++
		}
	}
	return Series[T]{times: times, values: lv}, Series[T]{times: append([]time.Time(nil), times...), values: rv}
}

func alignLeft[T any](a, b Series[T]) (Series[T], Series[T]) {
	times := append([]time.Time(nil), a.times...)
	lv := append([]T(nil), a.values...)
	rv := make([]T, a.Len())
	j := 0
	for i := 0; i < a.Len(); i++ {
		for j < b.Len() && b.times[j].Before(a.times[i]) {
			j++
		}
		if j < b.Len() && b.times[j].Equal(a.times[i]) {
			rv[i] = b.values[j]
		}
	}
	return Series[T]{times: times, values: lv}, Series[T]{times: append([]time.Time(nil), times...), values: rv}
}

func alignOuter[T any](a, b Series[T]) (Series[T], Series[T]) {
	times := make([]time.Time, 0, a.Len()+b.Len())
	lv := make([]T, 0, a.Len()+b.Len())
	rv := make([]T, 0, a.Len()+b.Len())
	i, j := 0, 0
	var zero T
	for i < a.Len() || j < b.Len() {
		switch {
		case j >= b.Len() || (i < a.Len() && a.times[i].Before(b.times[j])):
			times = append(times, a.times[i])
			lv = append(lv, a.values[i])
			rv = append(rv, zero)
			i++
		case i >= a.Len() || (j < b.Len() && b.times[j].Before(a.times[i])):
			times = append(times, b.times[j])
			lv = append(lv, zero)
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
	return Series[T]{times: times, values: lv}, Series[T]{times: append([]time.Time(nil), times...), values: rv}
}

// AlignFloat is like Align but fills gaps with NaN.
func AlignFloat(a, b Series[float64], how JoinHow) (left, right Series[float64]) {
	left, right = Align(a, b, how)
	if how == JoinInner {
		return left, right
	}
	// For JoinLeft / JoinOuter, zero-filled gaps should be NaN.
	// Re-run with NaN awareness by rebuilding from indices.
	switch how {
	case JoinLeft:
		return alignLeftFloat(a, b)
	default:
		return alignOuterFloat(a, b)
	}
}

func alignLeftFloat(a, b Series[float64]) (Series[float64], Series[float64]) {
	times := append([]time.Time(nil), a.times...)
	lv := append([]float64(nil), a.values...)
	rv := make([]float64, a.Len())
	j := 0
	for i := 0; i < a.Len(); i++ {
		rv[i] = math.NaN()
		for j < b.Len() && b.times[j].Before(a.times[i]) {
			j++
		}
		if j < b.Len() && b.times[j].Equal(a.times[i]) {
			rv[i] = b.values[j]
		}
	}
	return Series[float64]{times: times, values: lv}, Series[float64]{times: append([]time.Time(nil), times...), values: rv}
}

func alignOuterFloat(a, b Series[float64]) (Series[float64], Series[float64]) {
	times := make([]time.Time, 0, a.Len()+b.Len())
	lv := make([]float64, 0, a.Len()+b.Len())
	rv := make([]float64, 0, a.Len()+b.Len())
	i, j := 0, 0
	nan := math.NaN()
	for i < a.Len() || j < b.Len() {
		switch {
		case j >= b.Len() || (i < a.Len() && a.times[i].Before(b.times[j])):
			times = append(times, a.times[i])
			lv = append(lv, a.values[i])
			rv = append(rv, nan)
			i++
		case i >= a.Len() || (j < b.Len() && b.times[j].Before(a.times[i])):
			times = append(times, b.times[j])
			lv = append(lv, nan)
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
	return Series[float64]{times: times, values: lv}, Series[float64]{times: append([]time.Time(nil), times...), values: rv}
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
	times := append(append([]time.Time(nil), a.times...), b.times...)
	values := append(append([]T(nil), a.values...), b.values...)
	return Series[T]{times: times, values: values}, nil
}
