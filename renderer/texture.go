package renderer

import (
	"math"
	"math/rand"

	"github.com/go-gl/gl/v3.3-core/gl"
	loommath "github.com/ujjwalvivek/loom/math"
)

type Texture struct {
	ID     uint32
	Width  int32
	Height int32
}

func NewTexture(width, height int32, data []uint8, filter uint32) *Texture {
	var id uint32
	gl.GenTextures(1, &id)
	gl.BindTexture(gl.TEXTURE_2D, id)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, int32(filter))
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, int32(filter))

	var ptr *uint8 = nil
	if len(data) > 0 {
		ptr = &data[0]
	}

	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		width,
		height,
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(ptr),
	)

	return &Texture{
		ID:     id,
		Width:  width,
		Height: height,
	}
}

func (t *Texture) Bind(unit uint32) {
	gl.ActiveTexture(gl.TEXTURE0 + unit)
	gl.BindTexture(gl.TEXTURE_2D, t.ID)
}

func (t *Texture) Delete() {
	gl.DeleteTextures(1, &t.ID)
}

// GenerateSolidTexture creates a single-color 1x1 texture, useful for drawing raw rectangles.
func GenerateSolidTexture(color loommath.Color) *Texture {
	data := []uint8{
		uint8(color.R * 255),
		uint8(color.G * 255),
		uint8(color.B * 255),
		uint8(color.A * 255),
	}
	return NewTexture(1, 1, data, gl.NEAREST)
}

// GenerateRadialGradient creates a circular gradient texture (radial light map)
func GenerateRadialGradient(size int32) *Texture {
	data := make([]uint8, size*size*4)
	center := float32(size) / 2.0
	radius := float32(size) / 2.0

	for y := int32(0); y < size; y++ {
		for x := int32(0); x < size; x++ {
			dx := float32(x) + 0.5 - center
			dy := float32(y) + 0.5 - center
			dist := float32(math.Sqrt(float64(dx*dx + dy*dy)))

			val := 1.0 - (dist / radius)
			if val < 0 {
				val = 0
			}

			// Square for smoother falloff
			val = val * val

			idx := (y*size + x) * 4
			data[idx] = 255
			data[idx+1] = 255
			data[idx+2] = 255
			data[idx+3] = uint8(val * 255.0)
		}
	}
	return NewTexture(size, size, data, gl.LINEAR)
}

// GenerateNoiseTexture creates a 2D Perlin noise texture
func GenerateNoiseTexture(size int32, seed int64) *Texture {
	data := make([]uint8, size*size*4)
	pn := loommath.NewPerlinNoise2D(seed)

	for y := int32(0); y < size; y++ {
		for x := int32(0); x < size; x++ {
			// Frequency scaling
			nx := float64(x) / 16.0
			ny := float64(y) / 16.0
			val := pn.Noise(nx, ny)

			idx := (y*size + x) * 4
			gray := uint8(val * 255.0)
			data[idx] = gray
			data[idx+1] = gray
			data[idx+2] = gray
			data[idx+3] = 255
		}
	}
	return NewTexture(size, size, data, gl.LINEAR)
}

// GenerateCheckeredTexture creates a checkered debug pattern texture
func GenerateCheckeredTexture(size int32, gridCount int32) *Texture {
	data := make([]uint8, size*size*4)
	cellSize := size / gridCount

	for y := int32(0); y < size; y++ {
		for x := int32(0); x < size; x++ {
			cellX := x / cellSize
			cellY := y / cellSize

			gray := uint8(240)
			if (cellX+cellY)%2 == 0 {
				gray = uint8(180)
			}

			idx := (y*size + x) * 4
			data[idx] = gray
			data[idx+1] = gray
			data[idx+2] = gray
			data[idx+3] = 255
		}
	}
	return NewTexture(size, size, data, gl.NEAREST)
}

// GenerateWhiteNoiseTexture creates raw random pixel noise
func GenerateWhiteNoiseTexture(width, height int32) *Texture {
	data := make([]uint8, width*height*4)
	for i := 0; i < len(data); i += 4 {
		gray := uint8(rand.Intn(256))
		data[i] = gray
		data[i+1] = gray
		data[i+2] = gray
		data[i+3] = 255
	}
	return NewTexture(width, height, data, gl.NEAREST)
}
