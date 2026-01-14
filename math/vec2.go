package math

import "math"

type Vec2 struct {
	X, Y float32
}

// NewVec2 initializes and returns a new Vec2.
func NewVec2(x, y float32) Vec2 {
	return Vec2{X: x, Y: y}
}

func (v Vec2) Add(other Vec2) Vec2 {
	return Vec2{X: v.X + other.X, Y: v.Y + other.Y}
}

func (v Vec2) Sub(other Vec2) Vec2 {
	return Vec2{X: v.X - other.X, Y: v.Y - other.Y}
}

func (v Vec2) MulScalar(s float32) Vec2 {
	return Vec2{X: v.X * s, Y: v.Y * s}
}

// DivScalar divides the vector components by a scalar value.
// Returns a zero vector if the scalar is zero.
func (v Vec2) DivScalar(s float32) Vec2 {
	if s == 0 {
		return Vec2{}
	}
	return Vec2{X: v.X / s, Y: v.Y / s}
}

// Len calculates the Euclidean length (magnitude) of the vector.
func (v Vec2) Len() float32 {
	return float32(math.Sqrt(float64(v.X*v.X + v.Y*v.Y)))
}

// Normalize returns a unit vector with the same direction.
// Returns a zero vector if the original length is zero.
func (v Vec2) Normalize() Vec2 {
	l := v.Len()
	if l == 0 {
		return Vec2{}
	}
	return v.DivScalar(l)
}

func (v Vec2) Dot(other Vec2) float32 {
	return v.X*other.X + v.Y*other.Y
}

// Distance returns the Euclidean distance between v and other.
func (v Vec2) Distance(other Vec2) float32 {
	return v.Sub(other).Len()
}

func (v Vec2) Lerp(other Vec2, t float32) Vec2 {
	return Vec2{
		X: v.X + (other.X-v.X)*t,
		Y: v.Y + (other.Y-v.Y)*t,
	}
}
