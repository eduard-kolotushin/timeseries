package timeseries

import (
	"math"
	"testing"
	"time"
)

func TestResampleBuckets(t *testing.T) {
	t.Parallel()
	// origin t=0: points at 0,1,2,3,4 with values 1..5; every=2s => buckets [0,2), [2,4), [4,6)
	s := MustNew(
		[]time.Time{tAt(0), tAt(1), tAt(2), tAt(3), tAt(4)},
		[]float64{1, 2, 3, 4, 5},
	)
	r, err := Resample(s, 2*time.Second, AggSum)
	if err != nil {
		t.Fatal(err)
	}
	if r.Len() != 3 {
		t.Fatalf("len=%d want 3 vals=%v", r.Len(), r.Values())
	}
	if r.values[0] != 3 || r.values[1] != 7 || r.values[2] != 5 {
		t.Fatalf("sums=%v", r.Values())
	}
	if !r.times[0].Equal(tAt(0)) || !r.times[1].Equal(tAt(2)) || !r.times[2].Equal(tAt(4)) {
		t.Fatalf("times=%v", r.Times())
	}
}

func TestInterpolateAndUpsample(t *testing.T) {
	t.Parallel()
	s := MustNew([]time.Time{tAt(0), tAt(2)}, []float64{0, 10})
	out, err := Interpolate(s, []time.Time{tAt(0), tAt(1), tAt(2)}, InterpLinear)
	if err != nil {
		t.Fatal(err)
	}
	if out.values[1] != 5 {
		t.Fatalf("linear mid=%v", out.Values())
	}

	step, err := Interpolate(s, []time.Time{tAt(1)}, InterpStep)
	if err != nil || step.values[0] != 0 {
		t.Fatalf("step=%v err=%v", step.Values(), err)
	}

	up, err := Upsample(s, tAt(0), tAt(3), time.Second, InterpLinear)
	if err != nil || up.Len() != 3 || up.values[1] != 5 {
		t.Fatalf("upsample=%v err=%v", up.Values(), err)
	}

	outside, _ := Interpolate(s, []time.Time{tAt(3)}, InterpLinear)
	if !math.IsNaN(outside.values[0]) {
		t.Fatal("outside range should be NaN")
	}
}

func TestRegularGrid(t *testing.T) {
	t.Parallel()
	g, err := RegularGrid(tAt(0), tAt(3), time.Second)
	if err != nil || len(g) != 3 {
		t.Fatalf("grid=%v err=%v", g, err)
	}
	if _, err := RegularGrid(tAt(0), tAt(3), 0); err != ErrInvalidDuration {
		t.Fatalf("want ErrInvalidDuration, got %v", err)
	}
}
