package timeseries

import "math"

// Lag shifts values forward by n positions; the first n values become NaN.
func Lag(s Series[float64], n int) Series[float64] {
	if n < 0 {
		return Lead(s, -n)
	}
	values := make([]float64, s.Len())
	for i := range values {
		if i < n {
			values[i] = math.NaN()
		} else {
			values[i] = s.values[i-n]
		}
	}
	return s.withValues(values)
}

// Lead shifts values backward by n positions; the last n values become NaN.
func Lead(s Series[float64], n int) Series[float64] {
	if n < 0 {
		return Lag(s, -n)
	}
	values := make([]float64, s.Len())
	for i := range values {
		if i+n >= s.Len() {
			values[i] = math.NaN()
		} else {
			values[i] = s.values[i+n]
		}
	}
	return s.withValues(values)
}

// Diff returns s - Lag(s, n).
func Diff(s Series[float64], n int) Series[float64] {
	if n <= 0 {
		n = 1
	}
	values := make([]float64, s.Len())
	for i := range values {
		if i < n {
			values[i] = math.NaN()
			continue
		}
		values[i] = s.values[i] - s.values[i-n]
	}
	return s.withValues(values)
}

// PctChange returns (s - Lag(s, n)) / Lag(s, n). Division by zero yields NaN.
func PctChange(s Series[float64], n int) Series[float64] {
	if n <= 0 {
		n = 1
	}
	values := make([]float64, s.Len())
	for i := range values {
		if i < n {
			values[i] = math.NaN()
			continue
		}
		prev := s.values[i-n]
		if prev == 0 {
			values[i] = math.NaN()
			continue
		}
		values[i] = (s.values[i] - prev) / prev
	}
	return s.withValues(values)
}

// CumSum returns the cumulative sum, skipping NaNs (NaN positions stay NaN and do not reset the sum).
func CumSum(s Series[float64]) Series[float64] {
	values := make([]float64, s.Len())
	sum := 0.0
	for i, v := range s.values {
		if math.IsNaN(v) {
			values[i] = math.NaN()
			continue
		}
		sum += v
		values[i] = sum
	}
	return s.withValues(values)
}

// CumMax returns the cumulative maximum; NaN positions stay NaN.
func CumMax(s Series[float64]) Series[float64] {
	values := make([]float64, s.Len())
	max := math.Inf(-1)
	have := false
	for i, v := range s.values {
		if math.IsNaN(v) {
			values[i] = math.NaN()
			continue
		}
		if !have || v > max {
			max = v
			have = true
		}
		values[i] = max
	}
	return s.withValues(values)
}

// CumMin returns the cumulative minimum; NaN positions stay NaN.
func CumMin(s Series[float64]) Series[float64] {
	values := make([]float64, s.Len())
	min := math.Inf(1)
	have := false
	for i, v := range s.values {
		if math.IsNaN(v) {
			values[i] = math.NaN()
			continue
		}
		if !have || v < min {
			min = v
			have = true
		}
		values[i] = min
	}
	return s.withValues(values)
}
