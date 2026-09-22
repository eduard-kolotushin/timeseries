package timeseries

import (
	"math"
	"testing"
	"time"
)

func TestAlignJoins(t *testing.T) {
	t.Parallel()
	a := MustNew([]time.Time{tAt(1), tAt(2), tAt(3)}, []float64{1, 2, 3})
	b := MustNew([]time.Time{tAt(2), tAt(3), tAt(4)}, []float64{20, 30, 40})

	li, ri := AlignFloat(a, b, JoinInner)
	if li.Len() != 2 || li.values[0] != 2 || ri.values[0] != 20 {
		t.Fatalf("inner: L=%v R=%v", li.Values(), ri.Values())
	}

	ll, rl := AlignFloat(a, b, JoinLeft)
	if ll.Len() != 3 || !math.IsNaN(rl.values[0]) || rl.values[1] != 20 {
		t.Fatalf("left: L=%v R=%v", ll.Values(), rl.Values())
	}

	lo, ro := AlignFloat(a, b, JoinOuter)
	if lo.Len() != 4 || !math.IsNaN(lo.values[3]) || !math.IsNaN(ro.values[0]) {
		t.Fatalf("outer: L=%v R=%v", lo.Values(), ro.Values())
	}
}

func TestJoinLeftSharesTheLeftSeries(t *testing.T) {
	t.Parallel()
	a := MustNew([]time.Time{tAt(1), tAt(2)}, []float64{1, 2})
	b := MustNew([]time.Time{tAt(2), tAt(3)}, []float64{20, 30})
	ll, rl := AlignFloat(a, b, JoinLeft)

	// Documented contract: the left result is the left input itself, sharing its
	// time index and values (O(1), no copy) because no op mutates a series in
	// place and Times/Values/Points copy. Pinned so a future change to that
	// ownership cannot slip through unnoticed.
	if ll.Len() != a.Len() || &ll.times[0] != &a.times[0] || &ll.values[0] != &a.values[0] {
		t.Fatal("left result must be the left series itself")
	}
	if !equalFloats(ll.values, []float64{1, 2}) {
		t.Fatalf("left values = %v", ll.values)
	}
	mapped := ll.Values()
	mapped[0] = 99
	if ll.values[0] != 1 {
		t.Fatal("Values() must copy, so the sharing stays unobservable")
	}
	// The right result is aligned onto the left index, not onto b's.
	if len(rl.times) != 2 || &rl.times[0] != &a.times[0] {
		t.Fatal("right result must use the left index")
	}
	if !math.IsNaN(rl.values[0]) || rl.values[1] != 20 {
		t.Fatalf("right values = %v", rl.values)
	}
}

func TestMergeConcat(t *testing.T) {
	t.Parallel()
	a := MustNew([]time.Time{tAt(1), tAt(3)}, []float64{1, 3})
	b := MustNew([]time.Time{tAt(2), tAt(3)}, []float64{2, 30})

	if _, err := Merge(a, b, nil); err != ErrConflict {
		t.Fatalf("want conflict, got %v", err)
	}
	m, err := Merge(a, b, func(_ time.Time, x, y float64) (float64, error) {
		return x + y, nil
	})
	if err != nil || m.Len() != 3 || m.values[2] != 33 {
		t.Fatalf("merge: %v err=%v", m.Values(), err)
	}

	c1 := MustNew([]time.Time{tAt(1)}, []float64{1})
	c2 := MustNew([]time.Time{tAt(2)}, []float64{2})
	c, err := Concat(c1, c2)
	if err != nil || c.Len() != 2 {
		t.Fatalf("concat: %v err=%v", c.Values(), err)
	}
}
