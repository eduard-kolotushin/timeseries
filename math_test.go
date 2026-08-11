package timeseries

import (
	"math"
	"testing"
	"time"
)

func TestRolling(t *testing.T) {
	t.Parallel()
	s := MustNew([]time.Time{tAt(1), tAt(2), tAt(3), tAt(4)}, []float64{1, 2, 3, 4})
	r, err := Rolling(s, 2, AggSum)
	if err != nil {
		t.Fatal(err)
	}
	if !math.IsNaN(r.values[0]) || r.values[1] != 3 || r.values[2] != 5 || r.values[3] != 7 {
		t.Fatalf("rolling=%v", r.Values())
	}

	rd, err := RollingDuration(s, 2*time.Second, AggCount)
	if err != nil {
		t.Fatal(err)
	}
	// at t=3: points in (1,3] => t=2,3 => count 2; at t=4: (2,4] => 3,4 => 2
	if rd.values[0] != 1 || rd.values[2] != 2 {
		t.Fatalf("rolling duration=%v", rd.Values())
	}
}

func TestMathAndTransforms(t *testing.T) {
	t.Parallel()
	a := MustNew([]time.Time{tAt(1), tAt(2)}, []float64{10, 20})
	b := MustNew([]time.Time{tAt(1), tAt(2)}, []float64{2, 4})

	if !equalFloats(Add(a, b).Values(), []float64{12, 24}) {
		t.Fatal("Add")
	}
	if !equalFloats(Sub(a, b).Values(), []float64{8, 16}) {
		t.Fatal("Sub")
	}
	if !equalFloats(Mul(a, b).Values(), []float64{20, 80}) {
		t.Fatal("Mul")
	}
	if !equalFloats(Div(a, b).Values(), []float64{5, 5}) {
		t.Fatal("Div")
	}
	z := MustNew([]time.Time{tAt(1)}, []float64{0})
	if !math.IsNaN(Div(a, z).values[0]) {
		t.Fatal("Div by zero should be NaN")
	}

	if Abs(MustNew([]time.Time{tAt(1)}, []float64{-3})).values[0] != 3 {
		t.Fatal("Abs")
	}
	if Clip(a, 12, 15).values[0] != 12 || Clip(a, 12, 15).values[1] != 15 {
		t.Fatal("Clip")
	}

	lag := Lag(a, 1)
	if !math.IsNaN(lag.values[0]) || lag.values[1] != 10 {
		t.Fatalf("Lag=%v", lag.Values())
	}
	diff := Diff(a, 1)
	if !math.IsNaN(diff.values[0]) || diff.values[1] != 10 {
		t.Fatalf("Diff=%v", diff.Values())
	}
	pct := PctChange(a, 1)
	if !math.IsNaN(pct.values[0]) || pct.values[1] != 1 {
		t.Fatalf("PctChange=%v", pct.Values())
	}

	cs := CumSum(a)
	if cs.values[0] != 10 || cs.values[1] != 30 {
		t.Fatalf("CumSum=%v", cs.Values())
	}
	cm := CumMax(MustNew([]time.Time{tAt(1), tAt(2), tAt(3)}, []float64{1, 3, 2}))
	if cm.values[2] != 3 {
		t.Fatalf("CumMax=%v", cm.Values())
	}
	cn := CumMin(MustNew([]time.Time{tAt(1), tAt(2), tAt(3)}, []float64{3, 1, 2}))
	if cn.values[2] != 1 {
		t.Fatalf("CumMin=%v", cn.Values())
	}
}
