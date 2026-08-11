# Project intentions

## Goal

Provide a reusable Go library for managing univariate timeseries data and applying a full set of core and math operations (fill missing data, sampling/resampling, align, rolling windows, arithmetic, transforms, interpolation).

## Locked choices

| Decision | Choice |
| --- | --- |
| Data model | Generic univariate `Series[T]` |
| Numeric ops | Specialized for `Series[float64]` |
| Module path | `github.com/eduard-kolotushin/timeseries` |
| Go version | 1.26+ |
| Package shape | Single root package `timeseries` |

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

## v1 non-goals

Do not add these without first updating this document:

- Multi-column / DataFrame APIs
- CSV / JSON persistence or I/O helpers
- Plotting or visualization hooks
- OHLC / candlestick bars
- Business-day or exchange calendars
- Concurrent mutation of a shared series

## Quality bar

- Invariants enforced at construction (equal lengths, strictly ascending unique UTC times)
- Resample buckets are left-closed, right-open: `[t, t+d)`
- Division by zero yields `NaN` (documented, not an error)
- Public operations return new series; inputs are not mutated
- Table-driven tests cover fill, align, and resample edge cases
