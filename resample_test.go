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

func TestResampleBucketsFromFirstTimestamp(t *testing.T) {
	t.Parallel()
	// Bucket starts are measured from the first timestamp, not from the epoch:
	// 2s buckets over points at 5..9 are [5,7), [7,9), [9,11).
	s := MustNew(
		[]time.Time{tAt(5), tAt(6), tAt(7), tAt(8), tAt(9)},
		[]float64{1, 2, 3, 4, 5},
	)
	r, err := Resample(s, 2*time.Second, AggSum)
	if err != nil {
		t.Fatal(err)
	}
	if r.Len() != 3 || r.values[0] != 3 || r.values[1] != 7 || r.values[2] != 5 {
		t.Fatalf("values = %v, want [3 7 5]", r.Values())
	}
	if !r.times[0].Equal(tAt(5)) || !r.times[1].Equal(tAt(7)) || !r.times[2].Equal(tAt(9)) {
		t.Fatalf("times = %v, want 5s, 7s, 9s", r.Times())
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

	// Step interpolation is NaN outside the series range, like the linear branch.
	after, err := Interpolate(s, []time.Time{tAt(3)}, InterpStep)
	if err != nil || !math.IsNaN(after.values[0]) {
		t.Fatalf("step past the last point = %v err=%v, want NaN", after.Values(), err)
	}
	before, err := Interpolate(s, []time.Time{tAt(-1)}, InterpStep)
	if err != nil || !math.IsNaN(before.values[0]) {
		t.Fatalf("step before the first point = %v err=%v, want NaN", before.Values(), err)
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

func TestRegularGridPreallocIsCapped(t *testing.T) {
	t.Parallel()
	start := tAt(0)

	// Exactly at the cap: the capacity hint is clamped to it, not to n+1.
	atCap, err := RegularGrid(start, start.Add(gridPreallocCap*time.Second), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if len(atCap) != gridPreallocCap {
		t.Fatalf("len=%d, want %d", len(atCap), gridPreallocCap)
	}
	if cap(atCap) > gridPreallocCap {
		t.Fatalf("cap=%d, want at most %d", cap(atCap), gridPreallocCap)
	}

	// Past the cap only the hint is clamped, never the grid itself.
	over := gridPreallocCap + 8
	big, err := RegularGrid(start, start.Add(time.Duration(over)*time.Second), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if len(big) != over || !big[over-1].Equal(start.Add(time.Duration(over-1)*time.Second)) {
		t.Fatalf("len=%d, want %d points through %v", len(big), over, start.Add(time.Duration(over-1)*time.Second))
	}
}

func TestInterpolateSortsAndDropsDuplicates(t *testing.T) {
	t.Parallel()
	s := MustNew([]time.Time{tAt(0), tAt(2)}, []float64{0, 10})
	out, err := Interpolate(s, []time.Time{tAt(2), tAt(0), tAt(2), tAt(1)}, InterpLinear)
	if err != nil {
		t.Fatal(err)
	}
	if out.Len() != 3 {
		t.Fatalf("len=%d, want 3 (sorted ascending, duplicates dropped)", out.Len())
	}
	for i, want := range []time.Time{tAt(0), tAt(1), tAt(2)} {
		if !out.Times()[i].Equal(want) {
			t.Fatalf("times[%d]=%v, want %v", i, out.Times()[i], want)
		}
	}
	if !equalFloats(out.Values(), []float64{0, 5, 10}) {
		t.Fatalf("values=%v", out.Values())
	}
}
