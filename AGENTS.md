# AGENTS.md

Operating manual for agents working in this repository.

## Project

Reusable Go timeseries library: univariate `Series[T]` with float64 numeric helpers, implemented for correctness **and** performance.

- **Module:** `github.com/eduard-kolotushin/timeseries`
- **Go:** 1.26+

## Read first

1. [docs/INTENTIONS.md](docs/INTENTIONS.md) — product scope, non-goals, performance aims
2. [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) — layout, invariants, API and performance conventions

Cursor rules in [`.cursor/rules/`](.cursor/rules/) summarize the same constraints for every session.

## Hard constraints

- Single root package `timeseries` (no subpackages for core API)
- Generic `Series[T]` for structure; float64 ops as **package functions**
- Public ops return a **new** series (do not mutate caller inputs)
- Times are UTC, strictly ascending, unique (duplicates rejected)
- Missing float64 values use `math.NaN()`
- Stay within v1 scope unless `docs/INTENTIONS.md` is updated first
- Implement ops efficiently: share time indexes, pre-size, linear scans

## v1 in scope

Construction, slice/filter, merge/align, DropNA/fill, resample/upsample/downsample, rolling aggregates, arithmetic, lag/diff/pct_change, interpolate.

## v1 out of scope

Multi-column DataFrame, CSV/JSON I/O, plotting, OHLC bars, business calendars, concurrency-safe shared mutation, SIMD/GPU kernels, forecasting (see sibling `timeseries-forecast`).

## Workflow

- Prefer table-driven tests next to the code under test
- Document public API and bucket semantics (`[t, t+d)` for resample)
- Do not expand scope (DataFrame/IO/plotting) without updating INTENTIONS
- Keep changes focused; avoid unrelated refactors
- When touching a hot path, keep or add a benchmark in `bench_test.go`
