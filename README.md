# timeseries

Univariate timeseries library for Go. Operations are implemented for correctness and for low allocation on sorted indexes.

**Module:** `github.com/eduard-kolotushin/timeseries`  
**Go:** 1.26+

See [docs/INTENTIONS.md](docs/INTENTIONS.md) for scope and [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for design.

## Install

```bash
go get github.com/eduard-kolotushin/timeseries
```

## Quick start

```go
package main

import (
	"fmt"
	"math"
	"time"

	"github.com/eduard-kolotushin/timeseries"
)

func main() {
	times := []time.Time{
		time.Unix(0, 0).UTC(),
		time.Unix(1, 0).UTC(),
		time.Unix(2, 0).UTC(),
		time.Unix(3, 0).UTC(),
	}
	values := []float64{1, math.NaN(), 3, 4}

	s, err := timeseries.New(times, values)
	if err != nil {
		panic(err)
	}

	filled := timeseries.Fill(s, timeseries.FillForward)
	resampled, err := timeseries.Resample(filled, 2*time.Second, timeseries.AggMean)
	if err != nil {
		panic(err)
	}
	rolling, err := timeseries.Rolling(resampled, 1, timeseries.AggMean)
	if err != nil {
		panic(err)
	}

	fmt.Println(rolling.Values())
}
```

## Invariants

- Times stored as UTC
- Strictly ascending, unique timestamps
- `math.NaN()` marks missing `float64` values
- Resample buckets are left-closed, right-open: `[t, t+d)`
- Public operations return new series (inputs are not mutated)
- `Div` yields `NaN` on division by zero
- Ops that only rewrite values reuse the time index internally; `Times()` / `Values()` still return copies

## Agents

Contributors and coding agents: start with [AGENTS.md](AGENTS.md).
