# Architecture

## Layout

Single package `timeseries`. Files by concern:

| File | Responsibility |
| --- | --- |
| `series.go` | `Series[T]`, `Point[T]`, constructors, accessors, clone helpers |
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

## Performance

Optimize by default. Public behavior stays immutable; internals may share storage.

- **Share the time index** when an op only rewrites values (`withValues`).
- **Never mutate** `times` or `values` in place on an existing series. Allocate a new values slice, then attach it. Structural ops that grow/shrink must copy (or use a 3-index subslice so `append` cannot clobber a parent).
- **Pre-size** output slices (`make(..., 0, n)` or `make(..., n)`).
- **Merge/search** with two pointers or binary search; no per-point hidden sorts.
- **Skip a second `New`** when the op already produced a valid UTC, unique, sorted index.
- **Special-case same-index math** (lag/diff/pct_change) instead of going through generic align.
- Public `Times()`, `Values()`, and `Clone()` copy so callers cannot alias internals.
- Benchmarks live in `bench_test.go` for fill, lag, rolling, resample, interpolate, and arithmetic.

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
- Input isolation (ops must not mutate the caller’s series)
