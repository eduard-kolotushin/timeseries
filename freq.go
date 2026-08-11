package timeseries

import "time"

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
	out := make([]time.Time, 0, n+1)
	for t := start; t.Before(end); t = t.Add(step) {
		out = append(out, t)
	}
	return out, nil
}
