package math

type Color struct {
	R, G, B, A float32
}

func NewColor(r, g, b, a float32) Color {
	return Color{R: r, G: g, B: b, A: a}
}

// ColorFromRGBA8 converts 8-bit unsigned integers (0-255) into normalized float32 channels.
func ColorFromRGBA8(r, g, b, a uint8) Color {
	return Color{
		R: float32(r) / 255.0,
		G: float32(g) / 255.0,
		B: float32(b) / 255.0,
		A: float32(a) / 255.0,
	}
}

var (
	ColorWhite = Color{1, 1, 1, 1}
	ColorBlack = Color{0, 0, 0, 1}
	ColorRed   = Color{1, 0, 0, 1}
	ColorGreen = Color{0, 1, 0, 1}
	ColorBlue  = Color{0, 0, 1, 1}
)
