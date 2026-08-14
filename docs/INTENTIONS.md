# Project intentions

## Goal

Provide a reusable Go library for managing univariate timeseries data and applying a full set of core and math operations (fill missing data, sampling/resampling, align, rolling windows, arithmetic, transforms, interpolation).

Implement those operations in an **optimized** way: linear scans, minimal allocation, shared time indexes, and specialized float64 hot paths. Correctness stays first; performance is a standing design constraint, not a later rewrite.

## Locked choices

| Decision | Choice |
| --- | --- |
| Data model | Generic univariate `Series[T]` |
| Numeric ops | Specialized for `Series[float64]` |
| Module path | `github.com/eduard-kolotushin/timeseries` |
| Go version | 1.26+ |
| Package shape | Single root package `timeseries` |
| Performance | Optimize as we implement; do not ship naive extra copies or quadratic scans |

## v1 must-have operations

- Construction and accessors
- Slice / filter / clone / append / upsert / delete
- Align / merge / join (inner, outer, left)
- Missing data: DropNA, fill forward / backward / constant
- Sampling: resample, downsample, upsample, regular grid
- Rolling aggregates (mean, sum, min, max, std, count)
- Arithmetic (series–series and scalar), Map, Abs, Clip
- Transforms: Lag, Lead, Diff, PctChange, CumSum / CumMax / CumMin
- Interpolate: linear and step (previous)

## Performance aims

- Prefer **O(n)** (or O(n+m) merge) algorithms over nested scans
- Allocation should track **output size**, not hidden intermediates
- When only values change, **reuse the time index** (do not copy timestamps)
- Pre-size slices; avoid `append` growth in known-length loops
- Keep float64 numeric paths specialized (no interface{} / reflection)
- Add or update benchmarks when changing a hot path
- Do not mutate caller inputs; share backing storage only when it cannot be observed

## v1 non-goals

Do not add these without first updating this document:

- Multi-column / DataFrame APIs
- CSV / JSON persistence or I/O helpers
- Plotting or visualization hooks
- OHLC / candlestick bars
- Business-day or exchange calendars
- Concurrent mutation of a shared series
- SIMD/assembly kernels, GPU offload, or a separate “fast” API
- Forecasting models (those live in the sibling module `github.com/eduard-kolotushin/timeseries-forecast`)

## Quality bar

- Invariants enforced at construction (equal lengths, strictly ascending unique UTC times)
- Resample buckets are left-closed, right-open: `[t, t+d)`
- Division by zero yields `NaN` (documented, not an error)
- Public operations return new series; inputs are not mutated
- Table-driven tests cover fill, align, and resample edge cases
- Hot-path ops stay allocation-aware; public `Times()` / `Values()` still return copies
