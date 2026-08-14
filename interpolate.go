package timeseries

import (
	"math"
	"slices"
	"time"
)

// InterpMethod selects how values are interpolated onto a target index.
type InterpMethod int

const (
	// InterpLinear linearly interpolates between neighboring points.
	InterpLinear InterpMethod = iota
	// InterpStep uses the previous observation (step/previous).
	InterpStep
)

// Interpolate evaluates s at the given times.
// Times outside the series range become NaN.
// Target times need not be sorted; the result is sorted ascending with duplicates dropped.
func Interpolate(s Series[float64], times []time.Time, method InterpMethod) (Series[float64], error) {
	if len(times) == 0 {
		return Series[float64]{}, nil
	}
	sorted := make([]time.Time, len(times))
	for i, t := range times {
		sorted[i] = t.UTC()
	}
	slices.SortFunc(sorted, time.Time.Compare)
	uniq := sorted[:0]
	for i, t := range sorted {
		if i == 0 || !t.Equal(sorted[i-1]) {
			uniq = append(uniq, t)
		}
	}
	return interpolateOnto(s, uniq, method), nil
}

// interpolateOnto evaluates s at times. times must be UTC, strictly ascending, and unique.
// The returned series takes ownership of times.
func interpolateOnto(s Series[float64], times []time.Time, method InterpMethod) Series[float64] {
	values := make([]float64, len(times))
	j := 0
	n := s.Len()
	for i, t := range times {
		for j < n && s.times[j].Before(t) {
			j++
		}
		if j < n && s.times[j].Equal(t) {
			values[i] = s.values[j]
			continue
		}
		switch method {
		case InterpStep:
			if j == 0 {
				values[i] = math.NaN()
			} else {
				values[i] = s.values[j-1]
			}
		default:
			if j == 0 || j >= n {
				values[i] = math.NaN()
				continue
			}
			t0, t1 := s.times[j-1], s.times[j]
			v0, v1 := s.values[j-1], s.values[j]
			if math.IsNaN(v0) || math.IsNaN(v1) {
				values[i] = math.NaN()
				continue
			}
			span := t1.Sub(t0)
			if span == 0 {
				values[i] = v0
				continue
			}
			w := float64(t.Sub(t0)) / float64(span)
			values[i] = v0*(1-w) + v1*w
		}
	}
	return Series[float64]{times: times, values: values}
}
