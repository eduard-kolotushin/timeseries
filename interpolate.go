package timeseries

import (
	"math"
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
// Target times need not be sorted but the result is sorted ascending;
// duplicates in times are rejected via New.
func Interpolate(s Series[float64], times []time.Time, method InterpMethod) (Series[float64], error) {
	if len(times) == 0 {
		return Series[float64]{}, nil
	}
	// Sort/unique via constructing after computing values on sorted unique times.
	// Normalize and sort a copy.
	sorted := append([]time.Time(nil), times...)
	for i := range sorted {
		sorted[i] = sorted[i].UTC()
	}
	// insertion sort is fine for typical grids; use simple sort
	for i := 1; i < len(sorted); i++ {
		j := i
		for j > 0 && sorted[j].Before(sorted[j-1]) {
			sorted[j], sorted[j-1] = sorted[j-1], sorted[j]
			j--
		}
	}
	// dedupe
	uniq := sorted[:0]
	for i, t := range sorted {
		if i == 0 || !t.Equal(sorted[i-1]) {
			uniq = append(uniq, t)
		}
	}
	sorted = uniq

	values := make([]float64, len(sorted))
	for i, t := range sorted {
		values[i] = interpolateAt(s, t, method)
	}
	return New(sorted, values)
}

func interpolateAt(s Series[float64], t time.Time, method InterpMethod) float64 {
	if s.Empty() {
		return math.NaN()
	}
	t = t.UTC()
	if i := searchTime(s.times, t); i >= 0 {
		return s.values[i]
	}
	i := lowerBound(s.times, t)
	switch method {
	case InterpStep:
		if i == 0 {
			return math.NaN()
		}
		return s.values[i-1]
	default: // InterpLinear
		if i == 0 || i >= s.Len() {
			return math.NaN()
		}
		t0, t1 := s.times[i-1], s.times[i]
		v0, v1 := s.values[i-1], s.values[i]
		if math.IsNaN(v0) || math.IsNaN(v1) {
			return math.NaN()
		}
		span := t1.Sub(t0).Seconds()
		if span == 0 {
			return v0
		}
		w := t.Sub(t0).Seconds() / span
		return v0*(1-w) + v1*w
	}
}
