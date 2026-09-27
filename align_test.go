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

func TestJoinLeftKeepsTheLeftIndex(t *testing.T) {
	t.Parallel()
	a := MustNew([]time.Time{tAt(1), tAt(2)}, []float64{1, 2})
	b := MustNew([]time.Time{tAt(2), tAt(3)}, []float64{20, 30})
	ll, rl := AlignFloat(a, b, JoinLeft)

	// The left result is the left input itself (documented O(1) join that shares
	// the index and values because no op mutates a series in place); only the
	// observable half of that contract is asserted here, so a change to the
	// sharing that keeps the values would not fail this test.
	if !EqualFloat(ll, a) {
		t.Fatalf("left result = %v, want the left series unchanged", ll.Values())
	}
	// The right result is aligned onto the left index, not onto b's.
	if !rl.Times()[0].Equal(tAt(1)) || !rl.Times()[1].Equal(tAt(2)) {
		t.Fatalf("right times = %v, want the left index", rl.Times())
	}
	if !math.IsNaN(rl.values[0]) || rl.values[1] != 20 {
		t.Fatalf("right values = %v", rl.values)
	}
	mapped := ll.Values()
	mapped[0] = 99
	if ll.values[0] != 1 {
		t.Fatal("Values() must copy, so the sharing stays unobservable")
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
