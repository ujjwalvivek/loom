package termrenderer

import (
	"fmt"
	"math"
	"os"

	loommath "github.com/ujjwalvivek/loom/math"
)

// Cell represents a single character cell in the terminal.
type Cell struct {
	Char   string
	Fg     loommath.Color
	Bg     loommath.Color
	IsText bool
}

// TermRenderer manages a double-buffered 2D terminal grid and renders it using ANSI escapes.
type TermRenderer struct {
	Width, Height int
	Buffer        []Cell
	PrevBuffer    []Cell
}

// NewTermRenderer initializes the terminal renderer and enters the alternate screen buffer.
func NewTermRenderer(width, height int) *TermRenderer {
	tr := &TermRenderer{
		Width:      width,
		Height:     height,
		Buffer:     make([]Cell, width*height),
		PrevBuffer: make([]Cell, width*height),
	}

	// Initialize terminal: Alternate screen buffer, Hide cursor
	fmt.Print("\033[?1049h\033[?25l")

	return tr
}

// Shutdown restores the terminal state before exiting.
func (tr *TermRenderer) Shutdown() {
	// Restore terminal: Exit alternate screen, Show cursor, Reset formatting
	fmt.Print("\033[?1049l\033[?25h\033[0m")
}

// Clear fills the entire buffer with a background color.
func (tr *TermRenderer) Clear(bgColor loommath.Color) {
	for i := range tr.Buffer {
		tr.Buffer[i] = Cell{
			Char:   " ",
			Fg:     bgColor,
			Bg:     bgColor,
			IsText: false,
		}
	}
}

// SetPixel writes a color to a specific virtual pixel coordinate (0 to Height*2).
func (tr *TermRenderer) SetPixel(px, py int, color loommath.Color) {
	cx := px
	cy := py / 2
	if cx < 0 || cx >= tr.Width || cy < 0 || cy >= tr.Height {
		return
	}
	idx := cy*tr.Width + cx
	tr.Buffer[idx].IsText = false
	if py%2 == 0 {
		// Top half pixel color stored in Fg
		tr.Buffer[idx].Fg = color
	} else {
		// Bottom half pixel color stored in Bg
		tr.Buffer[idx].Bg = color
	}
}

// SetCell safely writes a character and colors to a specific grid coordinate.
func (tr *TermRenderer) SetCell(x, y int, char string, fg, bg loommath.Color) {
	if x < 0 || x >= tr.Width || y < 0 || y >= tr.Height {
		return
	}
	tr.Buffer[y*tr.Width+x] = Cell{
		Char:   char,
		Fg:     fg,
		Bg:     bg,
		IsText: true,
	}
}

// Rect draws a filled rectangle. Coordinates are in virtual pixels.
func (tr *TermRenderer) Rect(pos, size loommath.Vec2, color loommath.Color) {
	startX := int(math.Round(float64(pos.X)))
	startY := int(math.Round(float64(pos.Y)))
	w := int(math.Round(float64(size.X)))
	h := int(math.Round(float64(size.Y)))

	for y := startY; y < startY+h; y++ {
		for x := startX; x < startX+w; x++ {
			tr.SetPixel(x, y, color)
		}
	}
}

// Circle draws a filled circle using integer distance checking.
func (tr *TermRenderer) Circle(center loommath.Vec2, radius float32, color loommath.Color) {
	cx := int(math.Round(float64(center.X)))
	cy := int(math.Round(float64(center.Y)))
	r := int(math.Round(float64(radius)))
	r2 := r * r

	for y := cy - r; y <= cy+r; y++ {
		for x := cx - r; x <= cx+r; x++ {
			dx := x - cx
			dy := y - cy
			if dx*dx+dy*dy <= r2 {
				tr.SetPixel(x, y, color)
			}
		}
	}
}

// Text draws a string at the specified character cell coordinates.
func (tr *TermRenderer) Text(text string, pos loommath.Vec2, fg, bg loommath.Color) {
	startX := int(math.Round(float64(pos.X)))
	startY := int(math.Round(float64(pos.Y)))

	for i, r := range text {
		x := startX + i
		y := startY
		if x < 0 || x >= tr.Width || y < 0 || y >= tr.Height {
			continue
		}
		idx := y*tr.Width + x
		cell := &tr.Buffer[idx]

		if bg.A == 0 {
			// Merge: use the underlying top half pixel color as the background
			cell.Bg = cell.Fg
		} else {
			cell.Bg = bg
		}
		cell.Fg = fg
		cell.Char = string(r)
		cell.IsText = true
	}
}

// Present compiles the 2D buffer into a single string of ANSI escapes and writes it to stdout.
// Optimized as a differential renderer: it only writes characters and ANSI escapes for cells that changed.
func (tr *TermRenderer) Present() {
	// Pre-allocate approximate capacity to avoid reallocation
	buf := make([]byte, 0, tr.Width*tr.Height*12)

	var lastFg, lastBg loommath.Color
	colorInit := false

	// Track the virtual cursor position
	cursorX := -1
	cursorY := -1

	// Fast allocation-free integer to ASCII append helper
	appendInt := func(val int) {
		if val == 0 {
			buf = append(buf, '0')
			return
		}
		var temp [10]byte
		i := len(temp)
		for val > 0 {
			i--
			temp[i] = byte('0' + val%10)
			val /= 10
		}
		buf = append(buf, temp[i:]...)
	}

	for y := 0; y < tr.Height; y++ {
		for x := 0; x < tr.Width; x++ {
			idx := y*tr.Width + x
			cell := tr.Buffer[idx]
			prev := tr.PrevBuffer[idx]

			// If cell is identical to previous frame, skip writing
			if cell.Char == prev.Char && cell.Fg == prev.Fg && cell.Bg == prev.Bg && cell.IsText == prev.IsText {
				continue
			}

			// We need to write this cell. Position cursor at (x, y) if not already there
			if cursorX != x || cursorY != y {
				// ANSI coordinates are 1-based: \033[row;colH
				buf = append(buf, "\033["...)
				appendInt(y + 1)
				buf = append(buf, ';')
				appendInt(x + 1)
				buf = append(buf, 'H')
				cursorX = x
				cursorY = y
			}

			var char string
			var fg, bg loommath.Color

			if cell.IsText {
				char = cell.Char
				fg = cell.Fg
				bg = cell.Bg
			} else {
				if cell.Fg == cell.Bg {
					char = " "
					fg = cell.Fg
					bg = cell.Bg
				} else {
					char = "▄"
					fg = cell.Bg // Foreground is bottom pixel color
					bg = cell.Fg // Background is top pixel color
				}
			}

			// Emit ANSI color codes only if they changed
			if !colorInit || fg != lastFg || bg != lastBg {
				fgR, fgG, fgB := int(fg.R*255), int(fg.G*255), int(fg.B*255)
				bgR, bgG, bgB := int(bg.R*255), int(bg.G*255), int(bg.B*255)

				// Clamp channels to 0-255
				if fgR < 0 { fgR = 0 } else if fgR > 255 { fgR = 255 }
				if fgG < 0 { fgG = 0 } else if fgG > 255 { fgG = 255 }
				if fgB < 0 { fgB = 0 } else if fgB > 255 { fgB = 255 }
				if bgR < 0 { bgR = 0 } else if bgR > 255 { bgR = 255 }
				if bgG < 0 { bgG = 0 } else if bgG > 255 { bgG = 255 }
				if bgB < 0 { bgB = 0 } else if bgB > 255 { bgB = 255 }

				// ANSI Truecolor format: \033[38;2;R;G;Bm\033[48;2;R;G;Bm
				buf = append(buf, "\033[38;2;"...)
				appendInt(fgR)
				buf = append(buf, ';')
				appendInt(fgG)
				buf = append(buf, ';')
				appendInt(fgB)
				buf = append(buf, 'm')

				buf = append(buf, "\033[48;2;"...)
				appendInt(bgR)
				buf = append(buf, ';')
				appendInt(bgG)
				buf = append(buf, ';')
				appendInt(bgB)
				buf = append(buf, 'm')
				
				lastFg = fg
				lastBg = bg
				colorInit = true
			}

			buf = append(buf, char...)
			cursorX++
		}
	}

	// Single flush of changed cells to stdout
	if len(buf) > 0 {
		os.Stdout.Write(buf)
	}

	// Copy current buffer to prev buffer for the next frame
	copy(tr.PrevBuffer, tr.Buffer)
}
