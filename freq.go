package timeseries

import "time"

// gridPreallocCap bounds RegularGrid's up-front allocation. A longer grid is still
// returned in full; only the capacity hint is clamped, so a mistyped step cannot ask
// for gigabytes before producing anything.
const gridPreallocCap = 1 << 20

// RegularGrid returns timestamps from start inclusive to end exclusive, stepping by step.
func RegularGrid(start, end time.Time, step time.Duration) ([]time.Time, error) {
	if step <= 0 {
		return nil, ErrInvalidDuration
	}
	start = start.UTC()
	end = end.UTC()
	if !start.Before(end) {
		return []time.Time{}, nil
	}
	n := int(end.Sub(start) / step)
	if n < 0 {
		n = 0
	}
	if n+1 > gridPreallocCap {
		n = gridPreallocCap - 1
	}
	out := make([]time.Time, 0, n+1)
	for t := start; t.Before(end); t = t.Add(step) {
		out = append(out, t)
	}
	return out, nil
}
