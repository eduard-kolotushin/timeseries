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
	copy(outT, times)
	copy(outV, values)
	if err := normalizeIndex(outT); err != nil {
		return nil, nil, err
	}
	return outT, outV, nil
}

// normalizeIndex rewrites times in place to UTC and rejects duplicate or
// descending timestamps. The caller must own times: New copies first, and
// FromPoints passes the slice it just built.
func normalizeIndex(times []time.Time) error {
	for i := range times {
		t := times[i].UTC()
		times[i] = t
		if i == 0 {
			continue
		}
		switch prev := times[i-1]; {
		case t.Equal(prev):
			return ErrDuplicateTime
		case t.Before(prev):
			return ErrUnsorted
		}
	}
	return nil
}
