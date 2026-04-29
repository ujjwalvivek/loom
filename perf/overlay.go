package perf

import (
	loommath "github.com/ujjwalvivek/loom/math"
)

type OverlayRenderer interface {
	Rect(pos, size loommath.Vec2, color loommath.Color)
	Line(a, b loommath.Vec2, color loommath.Color, thickness float32)
	Circle(pos loommath.Vec2, radius float32, color loommath.Color)
}

func RenderOverlay(r OverlayRenderer, stats FrameStats, frameTimes []float32, windowWidth, windowHeight float32) {
	// 1. Draw main semi-transparent panel backing
	panelPos := loommath.Vec2{X: 15, Y: 15}
	panelSize := loommath.Vec2{X: 280, Y: 260}
	r.Rect(panelPos, panelSize, loommath.Color{R: 0.05, G: 0.05, B: 0.08, A: 0.82})

	// Draw panel border
	r.Line(panelPos, loommath.Vec2{X: panelPos.X + panelSize.X, Y: panelPos.Y}, loommath.Color{R: 0.2, G: 0.2, B: 0.3, A: 1.0}, 1.5)
	r.Line(loommath.Vec2{X: panelPos.X + panelSize.X, Y: panelPos.Y}, loommath.Vec2{X: panelPos.X + panelSize.X, Y: panelPos.Y + panelSize.Y}, loommath.Color{R: 0.2, G: 0.2, B: 0.3, A: 1.0}, 1.5)
	r.Line(loommath.Vec2{X: panelPos.X + panelSize.X, Y: panelPos.Y + panelSize.Y}, loommath.Vec2{X: panelPos.X, Y: panelPos.Y + panelSize.Y}, loommath.Color{R: 0.2, G: 0.2, B: 0.3, A: 1.0}, 1.5)
	r.Line(loommath.Vec2{X: panelPos.X, Y: panelPos.Y + panelSize.Y}, panelPos, loommath.Color{R: 0.2, G: 0.2, B: 0.3, A: 1.0}, 1.5)

	// 2. Draw Frame Time Graph (rolling window plot)
	graphPos := loommath.Vec2{X: 30, Y: 40}
	graphSize := loommath.Vec2{X: 250, Y: 80}
	
	// Graph background
	r.Rect(graphPos, graphSize, loommath.Color{R: 0.01, G: 0.01, B: 0.02, A: 0.9})
	
	// Target frame time markers (33.3ms for 30fps and 16.6ms for 60fps)
	y60 := graphPos.Y + graphSize.Y - (16.6 / 33.3)*graphSize.Y
	y30 := graphPos.Y
	
	r.Line(loommath.Vec2{X: graphPos.X, Y: y60}, loommath.Vec2{X: graphPos.X + graphSize.X, Y: y60}, loommath.Color{R: 0.1, G: 0.4, B: 0.1, A: 0.5}, 1.0)
	r.Line(loommath.Vec2{X: graphPos.X, Y: y30}, loommath.Vec2{X: graphPos.X + graphSize.X, Y: y30}, loommath.Color{R: 0.4, G: 0.1, B: 0.1, A: 0.5}, 1.0)

	// Plot frame times
	if len(frameTimes) > 1 {
		xStep := graphSize.X / float32(len(frameTimes))
		for i := 0; i < len(frameTimes)-1; i++ {
			x1 := graphPos.X + float32(i)*xStep
			x2 := graphPos.X + float32(i+1)*xStep

			// Clamp frame times to max graph range of 33.3ms for visual scaling
			val1 := frameTimes[i]
			if val1 > 33.3 {
				val1 = 33.3
			}
			val2 := frameTimes[i+1]
			if val2 > 33.3 {
				val2 = 33.3
			}

			y1 := graphPos.Y + graphSize.Y - (val1/33.3)*graphSize.Y
			y2 := graphPos.Y + graphSize.Y - (val2/33.3)*graphSize.Y

			lineColor := loommath.Color{R: 0.2, G: 0.8, B: 0.2, A: 1.0}
			if val1 > 16.6 {
				// Spike warnings
				lineColor = loommath.Color{R: 0.9, G: 0.6, B: 0.1, A: 1.0}
			}

			r.Line(loommath.Vec2{X: x1, Y: y1}, loommath.Vec2{X: x2, Y: y2}, lineColor, 1.2)

			// GC pause marker spikes (drawn as red vertical lines)
			if i > 0 && stats.LastPauseMs > 0 && float32(i) == float32(len(frameTimes)-2) {
				r.Line(
					loommath.Vec2{X: x2, Y: graphPos.Y + graphSize.Y},
					loommath.Vec2{X: x2, Y: graphPos.Y},
					loommath.Color{R: 1.0, G: 0.1, B: 0.1, A: 0.8},
					1.5,
				)
			}
		}
	}

	// 3. Draw Telemetry Text Values (represented as mini-bars/sliders)
	// Heap Memory indicator bar (Max limit 512MB)
	memLvl := stats.HeapAllocMB / 512.0
	if memLvl > 1.0 {
		memLvl = 1.0
	}
	drawBar(r, loommath.Vec2{X: 30, Y: 135}, memLvl, loommath.Color{R: 0.2, G: 0.6, B: 0.9, A: 1.0})

	// 4. Interactive dynamic tuning sliders
	// GOGC Slider (Range 10 to 300)
	gogcVal := float32(stats.GOGCPercent-10) / 290.0
	if gogcVal < 0 {
		gogcVal = 0
	} else if gogcVal > 1 {
		gogcVal = 1
	}
	drawSlider(r, loommath.Vec2{X: 30, Y: 175}, gogcVal, loommath.Color{R: 0.9, G: 0.7, B: 0.1, A: 1.0})

	// GOMEMLIMIT Slider (Range 16MB to 512MB)
	limitVal := (stats.GOMemLimitMB - 16.0) / 496.0
	if limitVal < 0 {
		limitVal = 0
	} else if limitVal > 1 {
		limitVal = 1
	}
	drawSlider(r, loommath.Vec2{X: 30, Y: 215}, limitVal, loommath.Color{R: 0.8, G: 0.2, B: 0.2, A: 1.0})
}

func drawBar(r OverlayRenderer, pos loommath.Vec2, fill float32, color loommath.Color) {
	// Draw border
	r.Rect(pos, loommath.Vec2{X: 250, Y: 14}, loommath.Color{R: 0.1, G: 0.1, B: 0.15, A: 1.0})
	// Draw fill
	if fill > 0 {
		r.Rect(pos, loommath.Vec2{X: 250 * fill, Y: 14}, color)
	}
	
	// Frame border outline
	r.Line(pos, loommath.Vec2{X: pos.X + 250, Y: pos.Y}, loommath.Color{R: 0.2, G: 0.2, B: 0.3, A: 0.5}, 1.0)
	r.Line(loommath.Vec2{X: pos.X + 250, Y: pos.Y}, loommath.Vec2{X: pos.X + 250, Y: pos.Y + 14}, loommath.Color{R: 0.2, G: 0.2, B: 0.3, A: 0.5}, 1.0)
	r.Line(loommath.Vec2{X: pos.X + 250, Y: pos.Y + 14}, loommath.Vec2{X: pos.X, Y: pos.Y + 14}, loommath.Color{R: 0.2, G: 0.2, B: 0.3, A: 0.5}, 1.0)
	r.Line(loommath.Vec2{X: pos.X, Y: pos.Y + 14}, pos, loommath.Color{R: 0.2, G: 0.2, B: 0.3, A: 0.5}, 1.0)
}

func drawSlider(r OverlayRenderer, pos loommath.Vec2, val float32, color loommath.Color) {
	// Draw slider guide track line
	trackY := pos.Y + 7
	r.Line(loommath.Vec2{X: pos.X, Y: trackY}, loommath.Vec2{X: pos.X + 250, Y: trackY}, loommath.Color{R: 0.2, G: 0.2, B: 0.3, A: 1.0}, 2.0)
	
	// Draw slider sliding knob circle
	knobX := pos.X + val*250
	r.Circle(loommath.Vec2{X: knobX, Y: trackY}, 6.0, color)
}
