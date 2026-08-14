package timeseries

import (
	"math"
	"testing"
	"time"
)

func benchSeries(n int) Series[float64] {
	times := make([]time.Time, n)
	values := make([]float64, n)
	for i := range n {
		times[i] = time.Unix(int64(i), 0).UTC()
		values[i] = float64(i)
		if i%17 == 0 {
			values[i] = math.NaN()
		}
	}
	return MustNew(times, values)
}

func BenchmarkFillForward(b *testing.B) {
	s := benchSeries(10_000)
	b.ResetTimer()
	for b.Loop() {
		_ = Fill(s, FillForward)
	}
}

func BenchmarkLag(b *testing.B) {
	s := benchSeries(10_000)
	b.ResetTimer()
	for b.Loop() {
		_ = Lag(s, 5)
	}
}

func BenchmarkDiff(b *testing.B) {
	s := benchSeries(10_000)
	b.ResetTimer()
	for b.Loop() {
		_ = Diff(s, 1)
	}
}

func BenchmarkRollingMean(b *testing.B) {
	s := benchSeries(10_000)
	b.ResetTimer()
	for b.Loop() {
		_, _ = Rolling(s, 32, AggMean)
	}
}

func BenchmarkResampleSum(b *testing.B) {
	s := benchSeries(10_000)
	b.ResetTimer()
	for b.Loop() {
		_, _ = Resample(s, 10*time.Second, AggSum)
	}
}

func BenchmarkUpsampleLinear(b *testing.B) {
	s := benchSeries(1_000)
	start, end := s.times[0], s.times[s.Len()-1].Add(time.Second)
	b.ResetTimer()
	for b.Loop() {
		_, _ = Upsample(s, start, end, time.Second, InterpLinear)
	}
}

func BenchmarkAdd(b *testing.B) {
	a := benchSeries(10_000)
	c := Map(a, func(v float64) float64 { return v + 1 })
	b.ResetTimer()
	for b.Loop() {
		_ = Add(a, c)
	}
}
