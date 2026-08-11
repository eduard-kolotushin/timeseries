# Architecture

## Layout

Single package `timeseries`. Files by concern:

| File | Responsibility |
| --- | --- |
| `series.go` | `Series[T]`, `Point[T]`, constructors, accessors |
| `index.go` | Sorted time index helpers, binary search |
| `mutate.go` | Slice, filter, clone, append, upsert, delete |
| `align.go` | Align, merge, join |
| `missing.go` | DropNA, fill |
| `freq.go` | Regular time grids |
| `resample.go` | Resample / downsample / upsample |
| `interpolate.go` | Linear and step interpolation |
| `rolling.go` | Rolling window aggregates |
| `math.go` | Arithmetic and scalar ops |
| `transform.go` | Lag, diff, pct change, cumulative |
| `errors.go` | Sentinel errors |

## Core types

```go
type Point[T any] struct {
    Time  time.Time
    Value T
}

type Series[T any] struct {
    // unexported times []time.Time, values []T
}
```

## Invariants

1. `len(times) == len(values)`
2. Times are stored as UTC (`t.UTC()`)
3. Times are strictly ascending; duplicates are rejected
4. Public ops return a new `Series`; they do not mutate the receiver’s backing store in a way visible to callers (clone as needed)

## Missing values

For `float64`, missing is `math.NaN()`. Fill, DropNA, interpolate, and numeric ops treat NaN accordingly.

## API style

- **Methods on `Series[T]`:** structural ops (Len, Slice, Filter, Clone, Append, …)
- **Package functions on `Series[float64]`:** numeric ops (`Add`, `Fill`, `Resample`, `Rolling`, …)

## Errors

- Validate at construction (`New`, `FromPoints`)
- Invalid arguments (zero duration, non-positive window) return errors
- `Div` uses `NaN` on divide-by-zero instead of returning an error

## Testing

Table-driven unit tests in `*_test.go` beside each concern. Cover:

- Construction invariant failures
- Align join modes
- Fill edge cases (leading/trailing NaN)
- Resample bucket boundaries `[t, t+d)`
