package repository

import (
	"context"
	"strconv"
	"testing"
)

// Benchmarks
func BenchmarkMemStorage_UpdateGauge(b *testing.B) {
	s := NewMemStorage()
	ctx := context.Background()

	b.ResetTimer()
	for i := range b.N {
		s.UpdateGauge(ctx, "test_metric", float64(i))
	}
}

func BenchmarkMemStorage_UpdateCounter(b *testing.B) {
	s := NewMemStorage()
	ctx := context.Background()

	b.ResetTimer()
	for i := range b.N {
		s.UpdateCounter(ctx, "test_counter", int64(i))
	}
}

func BenchmarkMemStorage_GetGauges_Small(b *testing.B) {
	s := NewMemStorage()
	ctx := context.Background()

	// Pre-populate with 10 gauges
	for i := range 10 {
		s.UpdateGauge(ctx, "metric_"+strconv.Itoa(i), float64(i))
	}

	b.ResetTimer()
	for b.Loop() {
		s.GetGauges(ctx)
	}
}

func BenchmarkMemStorage_GetGauges_Medium(b *testing.B) {
	s := NewMemStorage()
	ctx := context.Background()

	// Pre-populate with 100 gauges
	for i := range 100 {
		s.UpdateGauge(ctx, "metric_"+strconv.Itoa(i), float64(i))
	}

	b.ResetTimer()
	for b.Loop() {
		s.GetGauges(ctx)
	}
}

func BenchmarkMemStorage_GetGauges_Large(b *testing.B) {
	s := NewMemStorage()
	ctx := context.Background()

	// Pre-populate with 1000 gauges
	for i := range 1000 {
		s.UpdateGauge(ctx, "metric_"+strconv.Itoa(i), float64(i))
	}

	b.ResetTimer()
	for b.Loop() {
		s.GetGauges(ctx)
	}
}

func BenchmarkMemStorage_GetCounters_Small(b *testing.B) {
	s := NewMemStorage()
	ctx := context.Background()

	// Pre-populate with 10 counters
	for i := range 10 {
		s.UpdateCounter(ctx, "counter_"+strconv.Itoa(i), int64(i))
	}

	b.ResetTimer()
	for b.Loop() {
		s.GetCounters(ctx)
	}
}

func BenchmarkMemStorage_GetCounters_Medium(b *testing.B) {
	s := NewMemStorage()
	ctx := context.Background()

	// Pre-populate with 100 counters
	for i := range 100 {
		s.UpdateCounter(ctx, "counter_"+strconv.Itoa(i), int64(i))
	}

	b.ResetTimer()
	for b.Loop() {
		s.GetCounters(ctx)
	}
}

func BenchmarkMemStorage_GetCounters_Large(b *testing.B) {
	s := NewMemStorage()
	ctx := context.Background()

	// Pre-populate with 1000 counters
	for i := range 1000 {
		s.UpdateCounter(ctx, "counter_"+strconv.Itoa(i), int64(i))
	}

	b.ResetTimer()
	for b.Loop() {
		s.GetCounters(ctx)
	}
}

func BenchmarkMemStorage_ConcurrentUpdates(b *testing.B) {
	s := NewMemStorage()
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%2 == 0 {
				s.UpdateGauge(ctx, "concurrent_metric", float64(i))
			} else {
				s.UpdateCounter(ctx, "concurrent_counter", int64(i))
			}
			i++
		}
	})
}

func BenchmarkMemStorage_ConcurrentReads(b *testing.B) {
	s := NewMemStorage()
	ctx := context.Background()

	// Pre-populate storage
	for i := range 100 {
		s.UpdateGauge(ctx, "metric_"+strconv.Itoa(i), float64(i))
		s.UpdateCounter(ctx, "counter_"+strconv.Itoa(i), int64(i))
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			s.GetGauges(ctx)
			s.GetCounters(ctx)
		}
	})
}
