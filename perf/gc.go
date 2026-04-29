package perf

import (
	"runtime/debug"
	"runtime/metrics"
	"time"
)

type GCTelemetry struct {
	LastPauseDuration time.Duration
	PauseHistory      []time.Duration
	LastCycleCount    uint64

	// Pre-allocated GCStats struct to prevent runtime allocations
	gcStats debug.GCStats
}

func NewGCTelemetry() *GCTelemetry {
	gt := &GCTelemetry{
		PauseHistory: make([]time.Duration, 0, 100),
	}
	gt.gcStats.Pause = make([]time.Duration, 20)
	return gt
}

// Update polls Go runtime metrics. Returns true if a new GC cycle completed.
func (gt *GCTelemetry) Update() bool {
	const cycleMetric = "/gc/cycles:completed:count"
	samples := make([]metrics.Sample, 1)
	samples[0].Name = cycleMetric

	metrics.Read(samples)
	if samples[0].Value.Kind() == metrics.KindUint64 {
		count := samples[0].Value.Uint64()
		if count > gt.LastCycleCount {
			gt.LastCycleCount = count

			// A GC cycle completed, fetch pause details directly from internal circular log (no STW!)
			debug.ReadGCStats(&gt.gcStats)
			if len(gt.gcStats.Pause) > 0 {
				gt.LastPauseDuration = gt.gcStats.Pause[0]
				gt.PauseHistory = append(gt.PauseHistory, gt.LastPauseDuration)
				if len(gt.PauseHistory) > 100 {
					gt.PauseHistory = gt.PauseHistory[1:]
				}
			}
			return true
		}
	}
	return false
}
