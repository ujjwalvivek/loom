package math

import (
	"math"
	"math/rand"
)

// LFSR represents a 16-bit Galois Linear-Feedback Shift Register generator
type LFSR struct {
	state uint16
}

// NewLFSR creates a new LFSR generator with a seed
func NewLFSR(seed uint16) *LFSR {
	if seed == 0 {
		seed = 0xACE1 // must not be 0
	}
	return &LFSR{state: seed}
}

// Next returns a pseudo-random float32 value in [0.0, 1.0)
func (l *LFSR) Next() float32 {
	bit := l.state & 1
	l.state >>= 1
	if bit != 0 {
		l.state ^= 0xB400 // Taps: 16, 14, 13, 11
	}
	return float32(l.state) / 65535.0
}

// PerlinNoise2D generates a standard 2D gradient noise
type PerlinNoise2D struct {
	permutation [512]int
}

// NewPerlinNoise2D creates a Perlin noise generator
func NewPerlinNoise2D(seed int64) *PerlinNoise2D {
	p := &PerlinNoise2D{}
	r := rand.New(rand.NewSource(seed))
	
	// Create a permutation array
	perm := r.Perm(256)
	for i := 0; i < 256; i++ {
		p.permutation[i] = perm[i]
		p.permutation[256+i] = perm[i]
	}
	return p
}

func fade(t float64) float64 {
	return t * t * t * (t*(t*6-15) + 10)
}

func lerp(t, a, b float64) float64 {
	return a + t*(b-a)
}

func grad2D(hash int, x, y float64) float64 {
	// hash & 7 yields 8 gradient vectors: (1,1), (-1,1), (1,-1), (-1,-1), (1,0), (-1,0), (0,1), (0,-1)
	switch hash & 7 {
	case 0:
		return x + y
	case 1:
		return -x + y
	case 2:
		return x - y
	case 3:
		return -x - y
	case 4:
		return x
	case 5:
		return -x
	case 6:
		return y
	case 7:
		return -y
	}
	return 0
}

// Noise returns Perlin noise value at (x, y), scaled to range [0.0, 1.0]
func (p *PerlinNoise2D) Noise(x, y float64) float64 {
	// Find unit grid cell containing point
	X := int(math.Floor(x)) & 255
	Y := int(math.Floor(y)) & 255

	// Get relative coordinates of point in cell
	x -= math.Floor(x)
	y -= math.Floor(y)

	// Compute fade curves
	u := fade(x)
	v := fade(y)

	// Hash coordinates of the 4 cell corners
	aa := p.permutation[p.permutation[X]+Y]
	ab := p.permutation[p.permutation[X]+Y+1]
	ba := p.permutation[p.permutation[X+1]+Y]
	bb := p.permutation[p.permutation[X+1]+Y+1]

	// Add blended results from 4 corners
	res := lerp(v, lerp(u, grad2D(aa, x, y),
		grad2D(ba, x-1, y)),
		lerp(u, grad2D(ab, x, y-1),
			grad2D(bb, x-1, y-1)))

	// Scale and shift from [-1, 1] to [0, 1]
	return (res + 1.0) / 2.0
}
