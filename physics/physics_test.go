package physics

import (
	"testing"

	loommath "github.com/ujjwalvivek/loom/math"
)

func TestPhysicsSweptAABB(t *testing.T) {
	// Static body B
	b := loommath.Rect{X: 10, Y: 10, W: 5, H: 10}

	// Moving body A, starting 5 units away from B's left edge
	a := loommath.Rect{X: 0, Y: 10, W: 5, H: 10}
	
	// Velocity designed to hit B at exactly half duration (TOI = 0.5)
	vel := loommath.Vec2{X: 10, Y: 0}

	toi, normal := Sweep(a, b, vel)
	if toi != 0.5 {
		t.Errorf("Expected TOI 0.5, got %f", toi)
	}
	if normal.X != -1.0 || normal.Y != 0.0 {
		t.Errorf("Expected normal (-1, 0), got (%f, %f)", normal.X, normal.Y)
	}
}
