package math

// Rect represents an axis-aligned bounding box using a bottom-left origin.
type Rect struct {
	X, Y, W, H float32
}

func NewRect(x, y, w, h float32) Rect {
	return Rect{X: x, Y: y, W: w, H: h}
}

func (r Rect) Min() Vec2 {
	return Vec2{X: r.X, Y: r.Y}
}

func (r Rect) Max() Vec2 {
	return Vec2{X: r.X + r.W, Y: r.Y + r.H}
}

func (r Rect) Contains(p Vec2) bool {
	return p.X >= r.X && p.X <= r.X+r.W && p.Y >= r.Y && p.Y <= r.Y+r.H
}

func (r Rect) Intersects(other Rect) bool {
	return r.X < other.X+other.W &&
		r.X+r.W > other.X &&
		r.Y < other.Y+other.H &&
		r.Y+r.H > other.Y
}
