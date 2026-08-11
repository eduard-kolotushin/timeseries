package timeseries

import (
	"math"
	"time"
)

// Aggregator reduces a window of float64 values to a single value.
// Implementations should ignore NaNs unless documented otherwise.
type Aggregator func(values []float64) float64

// AggMean returns the mean of non-NaN values, or NaN if none.
func AggMean(values []float64) float64 {
	sum := 0.0
	n := 0
	for _, v := range values {
		if math.IsNaN(v) {
			continue
		}
		sum += v
		n++
	}
	if n == 0 {
		return math.NaN()
	}
	return sum / float64(n)
}

// AggSum returns the sum of non-NaN values, or NaN if none.
func AggSum(values []float64) float64 {
	sum := 0.0
	n := 0
	for _, v := range values {
		if math.IsNaN(v) {
			continue
		}
		sum += v
		n++
	}
	if n == 0 {
		return math.NaN()
	}
	return sum
}

// AggMin returns the minimum of non-NaN values, or NaN if none.
func AggMin(values []float64) float64 {
	min := math.Inf(1)
	n := 0
	for _, v := range values {
		if math.IsNaN(v) {
			continue
		}
		if v < min {
			min = v
		}
		n++
	}
	if n == 0 {
		return math.NaN()
	}
	return min
}

// AggMax returns the maximum of non-NaN values, or NaN if none.
func AggMax(values []float64) float64 {
	max := math.Inf(-1)
	n := 0
	for _, v := range values {
		if math.IsNaN(v) {
			continue
		}
		if v > max {
			max = v
		}
		n++
	}
	if n == 0 {
		return math.NaN()
	}
	return max
}

// AggFirst returns the first non-NaN value, or NaN if none.
func AggFirst(values []float64) float64 {
	for _, v := range values {
		if !math.IsNaN(v) {
			return v
		}
	}
	return math.NaN()
}

// AggLast returns the last non-NaN value, or NaN if none.
func AggLast(values []float64) float64 {
	for i := len(values) - 1; i >= 0; i-- {
		if !math.IsNaN(values[i]) {
			return values[i]
		}
	}
	return math.NaN()
}

// AggCount returns the count of non-NaN values.
func AggCount(values []float64) float64 {
	n := 0
	for _, v := range values {
		if !math.IsNaN(v) {
			n++
		}
	}
	return float64(n)
}

// AggStd returns the sample standard deviation of non-NaN values, or NaN if fewer than 2.
func AggStd(values []float64) float64 {
	sum := 0.0
	n := 0
	for _, v := range values {
		if math.IsNaN(v) {
			continue
		}
		sum += v
		n++
	}
	if n < 2 {
		return math.NaN()
	}
	mean := sum / float64(n)
	var ss float64
	for _, v := range values {
		if math.IsNaN(v) {
			continue
		}
		d := v - mean
		ss += d * d
	}
	return math.Sqrt(ss / float64(n-1))
}

// Resample aggregates s into buckets of width every.
// Bucket labels are bucket starts. Buckets are [t, t+every) relative to the
// first timestamp as origin. Empty buckets are omitted.
func Resample(s Series[float64], every time.Duration, agg Aggregator) (Series[float64], error) {
	if every <= 0 {
		return Series[float64]{}, ErrInvalidDuration
	}
	if s.Empty() {
		return Series[float64]{}, nil
	}
	if agg == nil {
		agg = AggMean
	}

	origin := s.times[0]
	type bucket struct {
		start  time.Time
		values []float64
	}
	var buckets []bucket

	for i := range s.times {
		t := s.times[i]
		offset := t.Sub(origin)
		n := offset / every
		if offset < 0 {
			// should not happen for sorted series starting at origin
			n = (offset - every + 1) / every
		}
		start := origin.Add(n * every)
		if len(buckets) == 0 || !buckets[len(buckets)-1].start.Equal(start) {
			buckets = append(buckets, bucket{start: start, values: []float64{s.values[i]}})
		} else {
			buckets[len(buckets)-1].values = append(buckets[len(buckets)-1].values, s.values[i])
		}
	}

	times := make([]time.Time, len(buckets))
	values := make([]float64, len(buckets))
	for i, b := range buckets {
		times[i] = b.start
		values[i] = agg(b.values)
	}
	return Series[float64]{times: times, values: values}, nil
}

// Downsample is an alias for Resample.
func Downsample(s Series[float64], every time.Duration, agg Aggregator) (Series[float64], error) {
	return Resample(s, every, agg)
}

// Upsample reindexes s onto a regular grid from start to end (end exclusive) with step,
// filling via interpolation method.
func Upsample(s Series[float64], start, end time.Time, step time.Duration, method InterpMethod) (Series[float64], error) {
	grid, err := RegularGrid(start, end, step)
	if err != nil {
		return Series[float64]{}, err
	}
	return Interpolate(s, grid, method)
}

// AsRegular reindexes s onto [start, end) with step using the given interpolation method.
func AsRegular(s Series[float64], start, end time.Time, step time.Duration, method InterpMethod) (Series[float64], error) {
	return Upsample(s, start, end, step, method)
}
