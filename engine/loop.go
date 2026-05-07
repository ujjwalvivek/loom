package engine

import (
	"time"
)

type Loop struct {
	TargetFPS     int
	FixedTimestep float32
	Accumulator   float32
	LastTime      time.Time
	Running       bool
}

func NewLoop(targetFPS int) *Loop {
	if targetFPS <= 0 {
		targetFPS = 60
	}
	return &Loop{
		TargetFPS:     targetFPS,
		FixedTimestep: 1.0 / float32(targetFPS),
		LastTime:      time.Now(),
		Running:       false,
	}
}

// Reset resets the timing baseline of the loop, preventing massive delta jumps.
func (l *Loop) Reset() {
	l.LastTime = time.Now()
	l.Accumulator = 0
}

// Tick updates the timer, returns the number of fixed updates to run, the frame's total delta time, and the alpha interpolation factor.
func (l *Loop) Tick() (int, float32, float32) {
	now := time.Now()
	dt := float32(now.Sub(l.LastTime).Seconds())
	l.LastTime = now

	// Prevent "spiral of death" (large spikes running too many ticks)
	if dt > 0.25 {
		dt = 0.25
	}

	l.Accumulator += dt

	ticks := 0
	for l.Accumulator >= l.FixedTimestep {
		ticks++
		l.Accumulator -= l.FixedTimestep
	}

	alpha := l.Accumulator / l.FixedTimestep
	return ticks, dt, alpha
}
