package renderer

import (
	"math"
	"math/rand"

	loommath "github.com/ujjwalvivek/loom/math"
)

type Camera struct {
	Pos    loommath.Vec2
	Zoom   float32
	Shake  float32
	shakeX float32
	shakeY float32
}

func NewCamera() *Camera {
	return &Camera{
		Zoom: 1.0,
	}
}

// Follow moves camera smoothly toward target
func (c *Camera) Follow(target loommath.Vec2, speed, dt float32) {
	c.Pos = c.Pos.Lerp(target, speed*dt)
}

// AddShake triggers screen shake intensity
func (c *Camera) AddShake(amount float32) {
	c.Shake += amount
	if c.Shake > 1.0 {
		c.Shake = 1.0
	}
}

// Update handles screen shake decay
func (c *Camera) Update(dt float32) {
	if c.Shake > 0 {
		angle := rand.Float32() * 2 * math.Pi
		offsetX := float32(math.Cos(float64(angle))) * c.Shake * 12.0
		offsetY := float32(math.Sin(float64(angle))) * c.Shake * 12.0
		c.shakeX = offsetX
		c.shakeY = offsetY

		c.Shake -= dt * 2.5 // decays completely in 0.4 seconds
		if c.Shake < 0 {
			c.Shake = 0
			c.shakeX = 0
			c.shakeY = 0
		}
	} else {
		c.shakeX = 0
		c.shakeY = 0
	}
}

// GetProjectionMatrix returns an ortho matrix centered around the camera view (Y-down)
func (c *Camera) GetProjectionMatrix(viewportWidth, viewportHeight float32) [16]float32 {
	w := viewportWidth / c.Zoom
	h := viewportHeight / c.Zoom

	left := c.Pos.X - w/2 + c.shakeX
	right := c.Pos.X + w/2 + c.shakeX
	bottom := c.Pos.Y + h/2 + c.shakeY
	top := c.Pos.Y - h/2 + c.shakeY

	return [16]float32{
		2.0 / (right - left), 0.0, 0.0, 0.0,
		0.0, 2.0 / (top - bottom), 0.0, 0.0,
		0.0, 0.0, -1.0, 0.0,
		-(right + left) / (right - left), -(top + bottom) / (top - bottom), 0.0, 1.0,
	}
}
