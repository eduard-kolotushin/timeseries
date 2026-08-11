package timeseries

import (
	"math"
	"testing"
	"time"
)

func TestDropNAAndFill(t *testing.T) {
	t.Parallel()
	s := MustNew(
		[]time.Time{tAt(1), tAt(2), tAt(3), tAt(4), tAt(5)},
		[]float64{math.NaN(), 1, math.NaN(), math.NaN(), 5},
	)

	d := DropNA(s)
	if d.Len() != 2 || d.values[0] != 1 || d.values[1] != 5 {
		t.Fatalf("DropNA: %v", d.Values())
	}

	ff := Fill(s, FillForward)
	if !math.IsNaN(ff.values[0]) || ff.values[2] != 1 || ff.values[3] != 1 || ff.values[4] != 5 {
		t.Fatalf("FillForward: %v", ff.Values())
	}

	fb := Fill(s, FillBackward)
	if fb.values[0] != 1 || fb.values[2] != 5 || fb.values[3] != 5 {
		t.Fatalf("FillBackward: %v", fb.Values())
	}

	fv := FillWith(s, 0)
	if fv.values[0] != 0 || fv.values[2] != 0 {
		t.Fatalf("FillWith: %v", fv.Values())
	}

	fl := FillLimit(s, FillForward, 1)
	if fl.values[2] != 1 || !math.IsNaN(fl.values[3]) {
		t.Fatalf("FillLimit: %v", fl.Values())
	}
}
