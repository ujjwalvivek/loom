package physics

import (
	"math"

	loommath "github.com/ujjwalvivek/loom/math"
)

type SpatialGrid struct {
	cellSize float32
	buckets  [2048][]BodyHandle
}

func NewSpatialGrid(cellSize float32) *SpatialGrid {
	g := &SpatialGrid{
		cellSize: cellSize,
	}
	for i := 0; i < len(g.buckets); i++ {
		g.buckets[i] = make([]BodyHandle, 0, 16)
	}
	return g
}

func (g *SpatialGrid) Clear() {
	for i := 0; i < len(g.buckets); i++ {
		g.buckets[i] = g.buckets[i][:0]
	}
}

func (g *SpatialGrid) hash(cellX, cellY int32) int {
	h := (cellX * 73856093) ^ (cellY * 19349663)
	if h < 0 {
		h = -h
	}
	return int(h % 2048)
}

func (g *SpatialGrid) Insert(body *Body) {
	if !body.Active {
		return
	}
	minX := int32(math.Floor(float64(body.Bounds.X / g.cellSize)))
	maxX := int32(math.Floor(float64((body.Bounds.X + body.Bounds.W) / g.cellSize)))
	minY := int32(math.Floor(float64(body.Bounds.Y / g.cellSize)))
	maxY := int32(math.Floor(float64((body.Bounds.Y + body.Bounds.H) / g.cellSize)))

	for x := minX; x <= maxX; x++ {
		for y := minY; y <= maxY; y++ {
			idx := g.hash(x, y)
			
			// Simple duplicate check within same cell bucket
			duplicate := false
			for _, h := range g.buckets[idx] {
				if h == body.Handle {
					duplicate = true
					break
				}
			}
			if !duplicate {
				g.buckets[idx] = append(g.buckets[idx], body.Handle)
			}
		}
	}
}

// Query collects unique BodyHandles in cells overlapping the query bounds.
// Appends to 'out' to prevent heap allocations.
func (g *SpatialGrid) Query(bounds loommath.Rect, out []BodyHandle) []BodyHandle {
	minX := int32(math.Floor(float64(bounds.X / g.cellSize)))
	maxX := int32(math.Floor(float64((bounds.X + bounds.W) / g.cellSize)))
	minY := int32(math.Floor(float64(bounds.Y / g.cellSize)))
	maxY := int32(math.Floor(float64((bounds.Y + bounds.H) / g.cellSize)))

	out = out[:0]

	for x := minX; x <= maxX; x++ {
		for y := minY; y <= maxY; y++ {
			idx := g.hash(x, y)
			for _, handle := range g.buckets[idx] {
				duplicate := false
				for _, h := range out {
					if h == handle {
						duplicate = true
						break
					}
				}
				if !duplicate {
					out = append(out, handle)
				}
			}
		}
	}
	return out
}
