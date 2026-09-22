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

func TestFromPointsBuildsAndValidatesInOnePass(t *testing.T) {
	t.Parallel()
	points := []Point[float64]{{Time: tAt(1), Value: 10}, {Time: tAt(2), Value: 20}}
	s, err := FromPoints(points)
	if err != nil {
		t.Fatal(err)
	}
	// The series owns its slices: rewriting the caller's points changes nothing.
	points[0] = Point[float64]{Time: tAt(9), Value: 99}
	if !equalFloats(s.Values(), []float64{10, 20}) || !s.Times()[0].Equal(tAt(1)) {
		t.Fatalf("series follows the caller's slice: %v %v", s.Times(), s.Values())
	}
	// The same index validation as New.
	if _, err := FromPoints([]Point[float64]{{Time: tAt(2), Value: 1}, {Time: tAt(1), Value: 2}}); err != ErrUnsorted {
		t.Fatalf("got %v, want ErrUnsorted", err)
	}
	if _, err := FromPoints([]Point[float64]{{Time: tAt(1), Value: 1}, {Time: tAt(1), Value: 2}}); err != ErrDuplicateTime {
		t.Fatalf("got %v, want ErrDuplicateTime", err)
	}
	if got, err := FromPoints([]Point[float64]{}); err != nil || !got.Empty() {
		t.Fatalf("empty series: len=%d err=%v", got.Len(), err)
	}
	// Times are normalized to UTC, as in New.
	aware := time.Date(2026, 1, 1, 12, 0, 0, 0, time.FixedZone("x", 3600))
	z, err := FromPoints([]Point[float64]{{Time: aware, Value: 1}})
	if err != nil {
		t.Fatal(err)
	}
	if got := z.Times()[0]; got.Location() != time.UTC || !got.Equal(aware) {
		t.Fatalf("time = %v (%v), want the same instant in UTC", got, got.Location())
	}
}

func TestFromPointsAllocation(t *testing.T) {
	// No t.Parallel: AllocsPerRun panics inside a parallel test.
	// FromPoints builds and validates one pair of slices; it must not also copy
	// them through New (docs/ARCHITECTURE.md: "Skip a second New when the op
	// already produced a valid UTC, unique, sorted index").
	points := []Point[float64]{{Time: tAt(1), Value: 10}, {Time: tAt(2), Value: 20}}
	allocs := testing.AllocsPerRun(50, func() {
		if _, err := FromPoints(points); err != nil {
			panic(err)
		}
	})
	if allocs != 2 {
		t.Fatalf("FromPoints allocates %v times, want 2 (times + values)", allocs)
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
