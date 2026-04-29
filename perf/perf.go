package perf

import (
	"runtime/debug"
	"runtime/metrics"
	"sort"
)

type FrameStats struct {
	FPS           float32
	FrameTimeMs   float32
	P50Ms         float32
	P99Ms         float32
	LastPauseMs   float32
	HeapAllocMB   float32
	GOGCPercent   int
	GOMemLimitMB  float32
}

type PerfSystem struct {
	FrameTimes []float32
	idx        int
	count      int
	GCTel      *GCTelemetry
}

func NewPerfSystem() *PerfSystem {
	return &PerfSystem{
		FrameTimes: make([]float32, 120), // Rolling window of 120 frames
		GCTel:      NewGCTelemetry(),
	}
}

func (ps *PerfSystem) AddFrame(duration float32) {
	ps.FrameTimes[ps.idx] = duration * 1000.0 // convert to milliseconds
	ps.idx = (ps.idx + 1) % len(ps.FrameTimes)
	if ps.count < len(ps.FrameTimes) {
		ps.count++
	}

	ps.GCTel.Update()
}

func (ps *PerfSystem) GetStats() FrameStats {
	if ps.count == 0 {
		return FrameStats{}
	}

	// 1. Calculate P50 and P99
	activeFrameTimes := make([]float32, ps.count)
	copy(activeFrameTimes, ps.FrameTimes[:ps.count])
	sort.Slice(activeFrameTimes, func(i, j int) bool {
		return activeFrameTimes[i] < activeFrameTimes[j]
	})

	p50 := activeFrameTimes[len(activeFrameTimes)/2]
	p99 := activeFrameTimes[int(float32(len(activeFrameTimes))*0.99)]

	var total float32
	for _, t := range activeFrameTimes {
		total += t
	}
	avg := total / float32(ps.count)
	fps := 1000.0 / avg
	if avg == 0 {
		fps = 0
	}

	// 2. Fetch runtime parameters using non-STW metrics API
	const (
		heapObjectsMetric = "/memory/classes/heap/objects:bytes"
		gogcMetric        = "/gc/gogc:percent"
		memlimitMetric    = "/gc/memlimit:bytes"
	)
	samples := make([]metrics.Sample, 3)
	samples[0].Name = heapObjectsMetric
	samples[1].Name = gogcMetric
	samples[2].Name = memlimitMetric

	metrics.Read(samples)

	var heapAlloc float32
	if samples[0].Value.Kind() == metrics.KindUint64 {
		heapAlloc = float32(samples[0].Value.Uint64()) / (1024 * 1024) // to MB
	}

	var gogc int
	if samples[1].Value.Kind() == metrics.KindUint64 {
		gogc = int(samples[1].Value.Uint64())
	}

	var memlimit float32
	if samples[2].Value.Kind() == metrics.KindUint64 {
		memlimit = float32(samples[2].Value.Uint64()) / (1024 * 1024) // to MB
	}

	return FrameStats{
		FPS:          fps,
		FrameTimeMs:  avg,
		P50Ms:        p50,
		P99Ms:        p99,
		LastPauseMs:  float32(ps.GCTel.LastPauseDuration.Seconds()) * 1000.0,
		HeapAllocMB:  heapAlloc,
		GOGCPercent:  gogc,
		GOMemLimitMB: memlimit,
	}
}

func (ps *PerfSystem) SetGCTuning(gogc int, memLimitBytes int64) {
	if gogc > 0 {
		debug.SetGCPercent(gogc)
	}
	if memLimitBytes > 0 {
		debug.SetMemoryLimit(memLimitBytes)
	}
}
