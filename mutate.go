package timeseries

import "time"

// SliceIndex returns s[i:j] (half-open). Indices are clamped like Go slicing rules
// after bounds checks: i and j must satisfy 0 <= i <= j <= Len().
func (s Series[T]) SliceIndex(i, j int) (Series[T], error) {
	if i < 0 || j < i || j > len(s.times) {
		return Series[T]{}, ErrIndexOutOfRange
	}
	return Series[T]{
		times:  s.times[i:j:j],
		values: s.values[i:j:j],
	}, nil
}

// Slice returns points with start <= t < end.
func (s Series[T]) Slice(start, end time.Time) Series[T] {
	i := lowerBound(s.times, start)
	j := lowerBound(s.times, end)
	out, _ := s.SliceIndex(i, j)
	return out
}

// Head returns the first n points (or all if n >= Len).
func (s Series[T]) Head(n int) Series[T] {
	if n <= 0 {
		return Series[T]{}
	}
	if n > s.Len() {
		n = s.Len()
	}
	out, _ := s.SliceIndex(0, n)
	return out
}

// Tail returns the last n points (or all if n >= Len).
func (s Series[T]) Tail(n int) Series[T] {
	if n <= 0 {
		return Series[T]{}
	}
	if n > s.Len() {
		n = s.Len()
	}
	out, _ := s.SliceIndex(s.Len()-n, s.Len())
	return out
}

// Filter returns points for which fn returns true.
func (s Series[T]) Filter(fn func(time.Time, T) bool) Series[T] {
	times := make([]time.Time, 0, len(s.times))
	values := make([]T, 0, len(s.values))
	for i := range s.times {
		if fn(s.times[i], s.values[i]) {
			times = append(times, s.times[i])
			values = append(values, s.values[i])
		}
	}
	return Series[T]{times: times, values: values}
}

// Append returns a new series with point (t, v) added.
// t must be strictly after the last timestamp if the series is non-empty.
func (s Series[T]) Append(t time.Time, v T) (Series[T], error) {
	t = t.UTC()
	if s.Len() > 0 {
		last := s.times[s.Len()-1]
		if !t.After(last) {
			if t.Equal(last) {
				return Series[T]{}, ErrDuplicateTime
			}
			return Series[T]{}, ErrUnsorted
		}
	}
	times := make([]time.Time, len(s.times)+1)
	values := make([]T, len(s.values)+1)
	copy(times, s.times)
	copy(values, s.values)
	times[len(s.times)] = t
	values[len(s.values)] = v
	return Series[T]{times: times, values: values}, nil
}

// Upsert inserts or replaces the value at time t.
func (s Series[T]) Upsert(t time.Time, v T) Series[T] {
	t = t.UTC()
	i := lowerBound(s.times, t)
	if i < len(s.times) && s.times[i].Equal(t) {
		values := cloneSlice(s.values)
		values[i] = v
		return s.withValues(values)
	}
	times := make([]time.Time, 0, len(s.times)+1)
	values := make([]T, 0, len(s.values)+1)
	times = append(times, s.times[:i]...)
	values = append(values, s.values[:i]...)
	times = append(times, t)
	values = append(values, v)
	times = append(times, s.times[i:]...)
	values = append(values, s.values[i:]...)
	return Series[T]{times: times, values: values}
}

// DeleteAt removes the point at index i.
func (s Series[T]) DeleteAt(i int) (Series[T], error) {
	if i < 0 || i >= len(s.times) {
		return Series[T]{}, ErrIndexOutOfRange
	}
	n := len(s.times) - 1
	times := make([]time.Time, n)
	values := make([]T, n)
	copy(times, s.times[:i])
	copy(times[i:], s.times[i+1:])
	copy(values, s.values[:i])
	copy(values[i:], s.values[i+1:])
	return Series[T]{times: times, values: values}, nil
}
