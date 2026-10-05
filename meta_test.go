package timeseries

import (
	"math"
	"testing"
	"time"
)

// gridSeries builds a series at tAt(0), tAt(1), ... one second apart.
func gridSeries(vs ...float64) Series[float64] {
	times := make([]time.Time, len(vs))
	for i := range vs {
		times[i] = tAt(int64(i))
	}
	return MustNew(times, vs)
}

func TestInterpolateStepOutsideRangeIsNaN(t *testing.T) {
	t.Parallel()
	s := MustNew([]time.Time{tAt(1), tAt(2), tAt(3), tAt(4)}, []float64{10, 20, 30, 40})
	targets := []time.Time{tAt(0), tAt(2), tAt(3).Add(500 * time.Millisecond), tAt(5)}
	got, err := Interpolate(s, targets, InterpStep)
	if err != nil {
		t.Fatal(err)
	}
	// Before the first observation and after the last one are outside the series
	// range, so they are NaN; in between, step/previous holds the last observation.
	want := []float64{math.NaN(), 20, 30, math.NaN()}
	if !equalTimes(got.times, targets) {
		t.Fatalf("times = %v, want %v", got.Times(), targets)
	}
	if !equalFloats(got.values, want) {
		t.Fatalf("values = %v, want %v", got.Values(), want)
	}
}

func TestLead(t *testing.T) {
	t.Parallel()
	s := gridSeries(1, 2, 3)
	if got := Lead(s, 1); !equalFloats(got.values, []float64{2, 3, math.NaN()}) {
		t.Fatalf("Lead(1) = %v, want [2 3 NaN]", got.Values())
	}
	if got := Lead(s, -1); !equalFloats(got.values, []float64{math.NaN(), 1, 2}) {
		t.Fatalf("Lead(-1) = %v, want [NaN 1 2]", got.Values())
	}
}

func TestAddScalarAndMulScalar(t *testing.T) {
	t.Parallel()
	s := gridSeries(1, math.NaN(), 3)
	if got := AddScalar(s, 10); !equalFloats(got.values, []float64{11, math.NaN(), 13}) {
		t.Fatalf("AddScalar = %v, want [11 NaN 13]", got.Values())
	}
	if got := MulScalar(gridSeries(1, 2, 3), 3); !equalFloats(got.values, []float64{3, 6, 9}) {
		t.Fatalf("MulScalar = %v, want [3 6 9]", got.Values())
	}
}
