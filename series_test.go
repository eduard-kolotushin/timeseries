package timeseries

import (
	"math"
	"testing"
	"time"
)

func tAt(sec int64) time.Time {
	return time.Unix(sec, 0).UTC()
}

func TestNewInvariants(t *testing.T) {
	t.Parallel()
	_, err := New([]time.Time{tAt(1)}, []float64{1, 2})
	if err != ErrLengthMismatch {
		t.Fatalf("got %v, want ErrLengthMismatch", err)
	}
	_, err = New([]time.Time{tAt(2), tAt(1)}, []float64{1, 2})
	if err != ErrUnsorted {
		t.Fatalf("got %v, want ErrUnsorted", err)
	}
	_, err = New([]time.Time{tAt(1), tAt(1)}, []float64{1, 2})
	if err != ErrDuplicateTime {
		t.Fatalf("got %v, want ErrDuplicateTime", err)
	}
}

func TestNewAndAccessors(t *testing.T) {
	t.Parallel()
	s, err := New([]time.Time{tAt(1), tAt(2)}, []float64{10, 20})
	if err != nil {
		t.Fatal(err)
	}
	if s.Len() != 2 || s.Empty() {
		t.Fatalf("len=%d empty=%v", s.Len(), s.Empty())
	}
	p, err := s.At(1)
	if err != nil || p.Value != 20 || !p.Time.Equal(tAt(2)) {
		t.Fatalf("At: %+v err=%v", p, err)
	}
	if !equalFloats(s.Values(), []float64{10, 20}) {
		t.Fatalf("Values: %v", s.Values())
	}
	c := s.Clone()
	c.values[0] = 99
	if s.values[0] != 10 {
		t.Fatal("Clone must copy values")
	}
}

func TestOpsDoNotMutateInput(t *testing.T) {
	t.Parallel()
	s := MustNew([]time.Time{tAt(1), tAt(2), tAt(3), tAt(4)}, []float64{1, math.NaN(), 3, 4})
	orig := cloneSlice(s.values)

	_ = Fill(s, FillForward)
	_ = Map(s, func(v float64) float64 { return v + 1 })
	_ = Lag(s, 1)
	_ = Diff(s, 1)
	sl := s.Slice(tAt(2), tAt(4))
	_ = Fill(sl, FillForward)
	if !equalFloats(s.values, orig) {
		t.Fatalf("input mutated: %v", s.values)
	}
}

func TestEqualFloat(t *testing.T) {
	t.Parallel()
	a := MustNew([]time.Time{tAt(1)}, []float64{math.NaN()})
	b := MustNew([]time.Time{tAt(1)}, []float64{math.NaN()})
	if !EqualFloat(a, b) {
		t.Fatal("NaNs should compare equal")
	}
}

func equalFloats(a, b []float64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if math.IsNaN(a[i]) && math.IsNaN(b[i]) {
			continue
		}
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
