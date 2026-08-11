package timeseries

import (
	"math"
	"time"
)

// FillMethod selects how missing (NaN) values are replaced.
type FillMethod int

const (
	// FillForward carries the last valid observation forward.
	FillForward FillMethod = iota
	// FillBackward carries the next valid observation backward.
	FillBackward
	// FillValue replaces NaNs with a constant (use FillWith).
	FillValue
)

// DropNA removes observations whose value is NaN.
func DropNA(s Series[float64]) Series[float64] {
	return s.Filter(func(_ time.Time, v float64) bool {
		return !math.IsNaN(v)
	})
}

// Trim drops leading and trailing NaN values.
func Trim(s Series[float64]) Series[float64] {
	start, end := 0, s.Len()
	for start < end && math.IsNaN(s.values[start]) {
		start++
	}
	for end > start && math.IsNaN(s.values[end-1]) {
		end--
	}
	out, _ := s.SliceIndex(start, end)
	return out
}

// Fill replaces NaN values according to method.
// For FillValue, use FillWith instead.
func Fill(s Series[float64], method FillMethod) Series[float64] {
	switch method {
	case FillBackward:
		return fillBackward(s, 0)
	case FillValue:
		return FillWith(s, 0)
	default:
		return fillForward(s, 0)
	}
}

// FillWith replaces all NaN values with v.
func FillWith(s Series[float64], v float64) Series[float64] {
	values := append([]float64(nil), s.values...)
	for i, x := range values {
		if math.IsNaN(x) {
			values[i] = v
		}
	}
	return Series[float64]{times: append([]time.Time(nil), s.times...), values: values}
}

// FillLimit is like Fill but stops after limit consecutive fills (0 = unlimited).
func FillLimit(s Series[float64], method FillMethod, limit int) Series[float64] {
	switch method {
	case FillBackward:
		return fillBackward(s, limit)
	case FillValue:
		return FillWith(s, 0)
	default:
		return fillForward(s, limit)
	}
}

func fillForward(s Series[float64], limit int) Series[float64] {
	values := append([]float64(nil), s.values...)
	var last float64
	have := false
	run := 0
	for i, v := range values {
		if !math.IsNaN(v) {
			last = v
			have = true
			run = 0
			continue
		}
		if !have {
			continue
		}
		run++
		if limit > 0 && run > limit {
			continue
		}
		values[i] = last
	}
	return Series[float64]{times: append([]time.Time(nil), s.times...), values: values}
}

func fillBackward(s Series[float64], limit int) Series[float64] {
	values := append([]float64(nil), s.values...)
	var next float64
	have := false
	run := 0
	for i := len(values) - 1; i >= 0; i-- {
		v := values[i]
		if !math.IsNaN(v) {
			next = v
			have = true
			run = 0
			continue
		}
		if !have {
			continue
		}
		run++
		if limit > 0 && run > limit {
			continue
		}
		values[i] = next
	}
	return Series[float64]{times: append([]time.Time(nil), s.times...), values: values}
}
