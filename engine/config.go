package engine

import "image"

type Config struct {
	Width     int
	Height    int
	Title     string
	TargetFPS int
	Pixelated bool
	GC        GCConfig
	Icon      []image.Image
}

type GCConfig struct {
	GOGC             int   // Go GC percent target
	GoMemLimit       int64 // Hard cap on Go runtime heap in MB
	PauseAnnotations bool  // Expose and annotate GC pause markers on perf graphs
}
