package timeseries

import (
	"math"
	"testing"
	"time"
)

func TestSliceFilterAppendUpsertDelete(t *testing.T) {
	t.Parallel()
	s := MustNew([]time.Time{tAt(1), tAt(2), tAt(3), tAt(4)}, []float64{1, 2, 3, 4})

	sl := s.Slice(tAt(2), tAt(4))
	if sl.Len() != 2 || sl.values[0] != 2 || sl.values[1] != 3 {
		t.Fatalf("Slice: %+v", sl.Points())
	}

	idx, err := s.SliceIndex(1, 3)
	if err != nil || idx.Len() != 2 {
		t.Fatalf("SliceIndex: %v err=%v", idx.Len(), err)
	}

	f := s.Filter(func(_ time.Time, v float64) bool { return int(v)%2 == 0 })
	if f.Len() != 2 || f.values[0] != 2 {
		t.Fatalf("Filter: %v", f.Values())
	}

	a, err := s.Append(tAt(5), 5)
	if err != nil || a.Len() != 5 {
		t.Fatalf("Append: len=%d err=%v", a.Len(), err)
	}
	if _, err := s.Append(tAt(4), 9); err != ErrDuplicateTime {
		t.Fatalf("Append dup: %v", err)
	}

	u := s.Upsert(tAt(2), 99)
	if u.values[1] != 99 || s.values[1] != 2 {
		t.Fatalf("Upsert replace failed: %v vs %v", u.Values(), s.Values())
	}
	u2 := s.Upsert(tAt(0), 0)
	if u2.Len() != 5 || u2.values[0] != 0 {
		t.Fatalf("Upsert insert: %v", u2.Values())
	}

	d, err := s.DeleteAt(1)
	if err != nil || d.Len() != 3 || d.values[1] != 3 {
		t.Fatalf("DeleteAt: %v err=%v", d.Values(), err)
	}

	if s.Head(2).Len() != 2 || s.Tail(1).values[0] != 4 {
		t.Fatal("Head/Tail failed")
	}
}

func TestTrim(t *testing.T) {
	t.Parallel()
	s := MustNew(
		[]time.Time{tAt(1), tAt(2), tAt(3), tAt(4)},
		[]float64{math.NaN(), 1, 2, math.NaN()},
	)
	tr := Trim(s)
	if tr.Len() != 2 || tr.values[0] != 1 || tr.values[1] != 2 {
		t.Fatalf("Trim: %v", tr.Values())
	}
}
