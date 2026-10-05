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

func TestEqualTreatsNaNAsUnequal(t *testing.T) {
	t.Parallel()
	a := MustNew([]time.Time{tAt(1), tAt(2)}, []float64{math.NaN(), 2})
	b := MustNew([]time.Time{tAt(1), tAt(2)}, []float64{math.NaN(), 2})
	if Equal(a, b) {
		t.Fatal("Equal must report a NaN position as unequal, as Go's != does")
	}
	if !EqualFloat(a, b) {
		t.Fatal("EqualFloat must treat NaN as equal to NaN")
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
	if allocs > 2 {
		t.Fatalf("FromPoints allocates %v times, want at most 2 (times + values)", allocs)
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

	tm, err := s.Time(1)
	if err != nil || !tm.Equal(tAt(2)) {
		t.Fatalf("Time: %v err=%v", tm, err)
	}
	v, err := s.Value(1)
	if err != nil || v != 20 {
		t.Fatalf("Value: %v err=%v", v, err)
	}
	if got := s.IndexOf(tAt(2)); got != 1 {
		t.Fatalf("IndexOf(hit) = %d, want 1", got)
	}
	if got := s.IndexOf(tAt(9)); got != -1 {
		t.Fatalf("IndexOf(miss) = %d, want -1", got)
	}
	for _, i := range []int{-1, 2} {
		if _, err := s.Time(i); err != ErrIndexOutOfRange {
			t.Fatalf("Time(%d) err=%v, want ErrIndexOutOfRange", i, err)
		}
		if _, err := s.Value(i); err != ErrIndexOutOfRange {
			t.Fatalf("Value(%d) err=%v, want ErrIndexOutOfRange", i, err)
		}
	}
	if _, err := s.At(-1); err != ErrIndexOutOfRange {
		t.Fatalf("At(-1) err=%v, want ErrIndexOutOfRange", err)
	}
}

func TestNewNormalizesNonUTCTimes(t *testing.T) {
	t.Parallel()
	aware := time.Date(2026, 1, 1, 12, 0, 0, 0, time.FixedZone("x", 3600))
	s, err := New([]time.Time{aware}, []float64{1})
	if err != nil {
		t.Fatal(err)
	}
	if got := s.Times()[0]; got.Location() != time.UTC || !got.Equal(aware) {
		t.Fatalf("time = %v (%v), want the same instant in UTC", got, got.Location())
	}
}

func TestMustConstructorPanics(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		call func()
	}{
		{"MustNew duplicate time", func() { MustNew([]time.Time{tAt(1), tAt(1)}, []float64{1, 2}) }},
		{"MustNew length mismatch", func() { MustNew([]time.Time{tAt(1)}, []float64{1, 2}) }},
		{"MustFromPoints duplicate time", func() {
			MustFromPoints([]Point[float64]{{Time: tAt(1), Value: 1}, {Time: tAt(1), Value: 2}})
		}},
	}
	for _, tc := range cases {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatalf("%s: expected panic", tc.name)
				}
			}()
			tc.call()
		}()
	}
}

func TestEqualAndEqualFloatBranches(t *testing.T) {
	t.Parallel()
	base := MustNew([]time.Time{tAt(1), tAt(2)}, []float64{1, 2})
	same := MustNew([]time.Time{tAt(1), tAt(2)}, []float64{1, 2})
	if !Equal(base, same) || !EqualFloat(base, same) {
		t.Fatal("identical NaN-free series must compare equal")
	}
	cases := []struct {
		name string
		b    Series[float64]
	}{
		{"value mismatch", MustNew([]time.Time{tAt(1), tAt(2)}, []float64{1, 3})},
		{"time mismatch", MustNew([]time.Time{tAt(1), tAt(3)}, []float64{1, 2})},
		{"length mismatch", MustNew([]time.Time{tAt(1)}, []float64{1})},
	}
	for _, tc := range cases {
		if Equal(base, tc.b) {
			t.Fatalf("Equal must be false on %s", tc.name)
		}
		if EqualFloat(base, tc.b) {
			t.Fatalf("EqualFloat must be false on %s", tc.name)
		}
	}
}

func TestOpsDoNotMutateInput(t *testing.T) {
	t.Parallel()
	idx := []time.Time{tAt(1), tAt(2), tAt(3), tAt(4)}
	s := MustNew(idx, []float64{1, math.NaN(), 3, 4})
	other := MustNew(idx, []float64{5, 6, 7, 8})
	beforeTimes := cloneSlice(s.times)
	beforeVals := cloneSlice(s.values)
	otherBefore := cloneSlice(other.values)

	cases := []struct {
		name string
		run  func()
	}{
		{"Fill", func() { _ = Fill(s, FillForward) }},
		{"FillWith", func() { _ = FillWith(s, 0) }},
		{"Map", func() { _ = Map(s, func(v float64) float64 { return v + 1 }) }},
		{"AddScalar", func() { _ = AddScalar(s, 10) }},
		{"MulScalar", func() { _ = MulScalar(s, 2) }},
		{"Abs", func() { _ = Abs(s) }},
		{"Clip", func() { _ = Clip(s, 0, 2) }},
		{"Lag", func() { _ = Lag(s, 1) }},
		{"Lead", func() { _ = Lead(s, 1) }},
		{"Diff", func() { _ = Diff(s, 1) }},
		{"PctChange", func() { _ = PctChange(s, 1) }},
		{"CumSum", func() { _ = CumSum(s) }},
		{"CumMax", func() { _ = CumMax(s) }},
		{"CumMin", func() { _ = CumMin(s) }},
		{"DropNA", func() { _ = DropNA(s) }},
		{"Trim", func() { _ = Trim(s) }},
		{"RollingDuration", func() { _, _ = RollingDuration(s, 2*time.Second, AggMean) }},
		{"Resample", func() { _, _ = Resample(s, 2*time.Second, AggSum) }},
		{"Downsample", func() { _, _ = Downsample(s, 2*time.Second, AggSum) }},
		{"Upsample", func() { _, _ = Upsample(s, tAt(1), tAt(4), time.Second, InterpLinear) }},
		{"Interpolate", func() { _, _ = Interpolate(s, []time.Time{tAt(2), tAt(5)}, InterpStep) }},
		{"Slice", func() { _ = s.Slice(tAt(2), tAt(4)) }},
		{"SliceInPlace", func() { _, _ = s.SliceIndex(0, 2) }},
		{"Align", func() { _, _ = Align(s, other, JoinOuter) }},
		{"Merge", func() { _, _ = Merge(s, other, nil) }},
		{"Add", func() { _ = Add(s, other) }},
		{"Div", func() { _ = Div(s, other) }},
		{"Times", func() { _ = s.Times() }},
		{"Values", func() { _ = s.Values() }},
		{"Points", func() { _ = s.Points() }},
		{"Clone", func() { _ = s.Clone() }},
		{"At", func() { _, _ = s.At(1) }},
		{"Filter", func() { _ = s.Filter(func(time.Time, float64) bool { return true }) }},
	}
	for _, tc := range cases {
		tc.run()
		if !equalTimes(s.times, beforeTimes) {
			t.Fatalf("%s mutated input times: %v", tc.name, s.times)
		}
		if !equalFloats(s.values, beforeVals) {
			t.Fatalf("%s mutated input values: %v", tc.name, s.values)
		}
		if !equalFloats(other.values, otherBefore) {
			t.Fatalf("%s mutated the second operand: %v", tc.name, other.values)
		}
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

func equalTimes(a, b []time.Time) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !a[i].Equal(b[i]) {
			return false
		}
	}
	return true
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
