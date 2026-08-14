package timeseries

import (
	"math"
	"time"
)

// Rolling applies agg over a trailing window of `window` points at each index.
// The first window-1 outputs are NaN. window must be > 0.
func Rolling(s Series[float64], window int, agg Aggregator) (Series[float64], error) {
	if window <= 0 {
		return Series[float64]{}, ErrInvalidWindow
	}
	if agg == nil {
		agg = AggMean
	}
	values := make([]float64, s.Len())
	for i := range s.values {
		if i+1 < window {
			values[i] = math.NaN()
			continue
		}
		values[i] = agg(s.values[i+1-window : i+1])
	}
	return s.withValues(values), nil
}

// RollingDuration applies agg over all points in (t-d, t] at each timestamp t.
func RollingDuration(s Series[float64], d time.Duration, agg Aggregator) (Series[float64], error) {
	if d <= 0 {
		return Series[float64]{}, ErrInvalidDuration
	}
	if agg == nil {
		agg = AggMean
	}
	values := make([]float64, s.Len())
	left := 0
	for i := range s.times {
		t := s.times[i]
		start := t.Add(-d)
		for left <= i && !s.times[left].After(start) {
			left++
		}
		// window is (start, t] => indices [left, i]
		if left > i {
			values[i] = math.NaN()
			continue
		}
		values[i] = agg(s.values[left : i+1])
	}
	return s.withValues(values), nil
}
