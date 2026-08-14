package timeseries

import (
	"slices"
	"time"
)

// searchTime returns the index of t in times, or -1 if not found.
// times must be sorted ascending (UTC).
func searchTime(times []time.Time, t time.Time) int {
	t = t.UTC()
	i, found := slices.BinarySearchFunc(times, t, time.Time.Compare)
	if !found {
		return -1
	}
	return i
}

// lowerBound returns the first index i in times such that times[i] >= t.
func lowerBound(times []time.Time, t time.Time) int {
	t = t.UTC()
	i, _ := slices.BinarySearchFunc(times, t, time.Time.Compare)
	return i
}

// upperBound returns the first index i in times such that times[i] > t.
func upperBound(times []time.Time, t time.Time) int {
	t = t.UTC()
	i, found := slices.BinarySearchFunc(times, t, time.Time.Compare)
	if found {
		return i + 1
	}
	return i
}

// validateIndex checks equal length and strictly ascending unique UTC times.
// It returns UTC-normalized copies of times (and a copy of values).
func validateIndex[T any](times []time.Time, values []T) ([]time.Time, []T, error) {
	if len(times) != len(values) {
		return nil, nil, ErrLengthMismatch
	}
	outT := make([]time.Time, len(times))
	outV := make([]T, len(values))
	copy(outV, values)
	for i, t := range times {
		outT[i] = t.UTC()
		if i > 0 {
			prev := outT[i-1]
			cur := outT[i]
			switch {
			case cur.Equal(prev):
				return nil, nil, ErrDuplicateTime
			case cur.Before(prev):
				return nil, nil, ErrUnsorted
			}
		}
	}
	return outT, outV, nil
}
