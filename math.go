package timeseries

import "math"

// Map applies fn to each value.
func Map(s Series[float64], fn func(float64) float64) Series[float64] {
	values := make([]float64, s.Len())
	for i, v := range s.values {
		values[i] = fn(v)
	}
	return s.withValues(values)
}

// AddScalar adds c to each value.
func AddScalar(s Series[float64], c float64) Series[float64] {
	return Map(s, func(v float64) float64 { return v + c })
}

// MulScalar multiplies each value by c.
func MulScalar(s Series[float64], c float64) Series[float64] {
	return Map(s, func(v float64) float64 { return v * c })
}

// Abs returns the absolute value of each element.
func Abs(s Series[float64]) Series[float64] {
	return Map(s, math.Abs)
}

// Clip clamps each value to [lo, hi]. NaN stays NaN.
func Clip(s Series[float64], lo, hi float64) Series[float64] {
	return Map(s, func(v float64) float64 {
		if math.IsNaN(v) {
			return v
		}
		if v < lo {
			return lo
		}
		if v > hi {
			return hi
		}
		return v
	})
}

func binaryOp(a, b Series[float64], op func(float64, float64) float64) Series[float64] {
	left, right := AlignFloat(a, b, JoinInner)
	values := make([]float64, left.Len())
	for i := range left.values {
		values[i] = op(left.values[i], right.values[i])
	}
	return left.withValues(values)
}

// Add returns the element-wise sum of a and b on their inner time join.
func Add(a, b Series[float64]) Series[float64] {
	return binaryOp(a, b, func(x, y float64) float64 { return x + y })
}

// Sub returns the element-wise difference a - b on their inner time join.
func Sub(a, b Series[float64]) Series[float64] {
	return binaryOp(a, b, func(x, y float64) float64 { return x - y })
}

// Mul returns the element-wise product of a and b on their inner time join.
func Mul(a, b Series[float64]) Series[float64] {
	return binaryOp(a, b, func(x, y float64) float64 { return x * y })
}

// Div returns the element-wise quotient a / b on their inner time join.
// Division by zero yields NaN.
func Div(a, b Series[float64]) Series[float64] {
	return binaryOp(a, b, func(x, y float64) float64 {
		if y == 0 {
			return math.NaN()
		}
		return x / y
	})
}
