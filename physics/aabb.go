package physics

import (
	"math"

	loommath "github.com/ujjwalvivek/loom/math"
)

type BodyHandle uint32

type Body struct {
	Handle    BodyHandle
	Entity    uint32
	Bounds    loommath.Rect
	Velocity  loommath.Vec2
	LayerMask uint32
	Active    bool
}

type Collision struct {
	Hit      bool
	Time     float32 // TOI from 0.0 to 1.0
	Normal   loommath.Vec2
	PenDepth float32
}

// Sweep checks collision between a moving box A and a static box B.
// Returns time of collision (0.0 to 1.0) and normal vector.
func Sweep(a, b loommath.Rect, vel loommath.Vec2) (float32, loommath.Vec2) {
	var xInvEntry, yInvEntry float32
	var xInvExit, yInvExit float32

	if vel.X > 0.0 {
		xInvEntry = b.X - (a.X + a.W)
		xInvExit = (b.X + b.W) - a.X
	} else {
		xInvEntry = (b.X + b.W) - a.X
		xInvExit = b.X - (a.X + a.W)
	}

	if vel.Y > 0.0 {
		yInvEntry = b.Y - (a.Y + a.H)
		yInvExit = (b.Y + b.H) - a.Y
	} else {
		yInvEntry = (b.Y + b.H) - a.Y
		yInvExit = b.Y - (a.Y + a.H)
	}

	var xEntry, yEntry float32
	var xExit, yExit float32

	if vel.X == 0.0 {
		if a.X+a.W <= b.X || a.X >= b.X+b.W {
			return 1.0, loommath.Vec2{}
		}
		xEntry = -float32(math.Inf(1))
		xExit = float32(math.Inf(1))
	} else {
		xEntry = xInvEntry / vel.X
		xExit = xInvExit / vel.X
	}

	if vel.Y == 0.0 {
		if a.Y+a.H <= b.Y || a.Y >= b.Y+b.H {
			return 1.0, loommath.Vec2{}
		}
		yEntry = -float32(math.Inf(1))
		yExit = float32(math.Inf(1))
	} else {
		yEntry = yInvEntry / vel.Y
		yExit = yInvExit / vel.Y
	}

	entryTime := xEntry
	if yEntry > xEntry {
		entryTime = yEntry
	}

	exitTime := xExit
	if yExit < xExit {
		exitTime = yExit
	}

	if entryTime > exitTime || (xEntry < 0.0 && yEntry < 0.0) || xEntry > 1.0 || yEntry > 1.0 {
		return 1.0, loommath.Vec2{}
	}

	var normal loommath.Vec2
	if xEntry > yEntry {
		if vel.X < 0.0 {
			normal = loommath.Vec2{X: 1, Y: 0}
		} else {
			normal = loommath.Vec2{X: -1, Y: 0}
		}
		// Y-axis Seam/corner check (<= 16.0)
		overlapY := minF(a.Y+a.H, b.Y+b.H) - maxF(a.Y, b.Y)
		if overlapY > 0 && overlapY <= 16.0 {
			return 1.0, loommath.Vec2{}
		}
	} else {
		if vel.Y < 0.0 {
			normal = loommath.Vec2{X: 0, Y: 1}
		} else {
			normal = loommath.Vec2{X: 0, Y: -1}
		}
		// X-axis Seam/corner check (<= 16.0)
		overlapX := minF(a.X+a.W, b.X+b.W) - maxF(a.X, b.X)
		if overlapX > 0 && overlapX <= 16.0 {
			return 1.0, loommath.Vec2{}
		}
	}

	return entryTime, normal
}

// Overlap computes minimal depenetration normal & depth.
func Overlap(a, b loommath.Rect) (bool, loommath.Vec2, float32) {
	if !a.Intersects(b) {
		return false, loommath.Vec2{}, 0
	}

	overlapX := float32(0.0)
	var normalX loommath.Vec2
	if a.X < b.X {
		overlapX = (a.X + a.W) - b.X
		normalX = loommath.Vec2{X: -1, Y: 0}
	} else {
		overlapX = (b.X + b.W) - a.X
		normalX = loommath.Vec2{X: 1, Y: 0}
	}

	overlapY := float32(0.0)
	var normalY loommath.Vec2
	if a.Y < b.Y {
		overlapY = (a.Y + a.H) - b.Y
		normalY = loommath.Vec2{X: 0, Y: -1}
	} else {
		overlapY = (b.Y + b.H) - a.Y
		normalY = loommath.Vec2{X: 0, Y: 1}
	}

	if overlapX < overlapY {
		return true, normalX, overlapX
	}
	return true, normalY, overlapY
}
