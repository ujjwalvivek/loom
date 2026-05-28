package main

import (
	"math"

	loommath "github.com/ujjwalvivek/loom/math"
	"github.com/ujjwalvivek/loom/termrenderer"
)

func DrawTriangle(ctx *termrenderer.TermRenderer, x, y, width, height float32, color loommath.Color) {
	layers := int(height / 4)
	if layers <= 0 { return }
	for i := 0; i < layers; i++ {
		w := width * float32(i) / float32(layers)
		ctx.Rect(loommath.Vec2{X: x - w/2, Y: y + float32(i*4)}, loommath.Vec2{X: w, Y: 4}, color)
	}
}

func DrawBackground(ctx *termrenderer.TermRenderer, camX, scale float32) {
	// A beautiful premium retro sky gradient, darkened to look rich and not blinding
	for i := 0; i < 40; i++ {
		ratio := float32(i) / 40.0
		r := 0.12 + (0.28-0.12)*ratio
		g := 0.30 + (0.52-0.30)*ratio
		b := 0.65 + (0.85-0.65)*ratio
		ctx.Rect(loommath.Vec2{X: 0, Y: float32(i*2)}, loommath.Vec2{X: 120, Y: 2}, loommath.Color{R: r, G: g, B: b, A: 1})
	}

	// Soft semi-transparent clouds that blend smoothly with the sky
	cloudWhite := loommath.Color{R: 0.95, G: 0.98, B: 1.0, A: 0.7}
	p3 := camX * 0.3
	startCol3 := int(p3 / (300 / scale))
	for i := startCol3 - 1; i < startCol3+5; i++ {
		cx := float32(i)*(300.0/scale) - p3
		cy := (float32(180+math.Sin(float64(i*5))*80) / scale) * 2.0
		
		ctx.Rect(loommath.Vec2{X: cx + (20/scale), Y: cy}, loommath.Vec2{X: 80/scale, Y: 60/scale}, cloudWhite)
		ctx.Rect(loommath.Vec2{X: cx + (40/scale), Y: cy - (40/scale)}, loommath.Vec2{X: 40/scale, Y: 80/scale}, cloudWhite)
		ctx.Rect(loommath.Vec2{X: cx, Y: cy + (30/scale)}, loommath.Vec2{X: 120/scale, Y: 40/scale}, cloudWhite)
	}
}

func DrawGroundBlock(ctx *termrenderer.TermRenderer, pos loommath.Vec2) {
	// Warm Nes-styled terracotta ground colors
	base := loommath.Color{R: 0.80, G: 0.38, B: 0.22, A: 1.0}
	dark := loommath.Color{R: 0.48, G: 0.18, B: 0.08, A: 1.0}
	light := loommath.Color{R: 0.92, G: 0.58, B: 0.40, A: 1.0}

	baseX := int(math.Round(float64(pos.X)))
	baseY := int(math.Round(float64(pos.Y)))

	// Perfect 8x8 pixel grid for ground tile
	grid := [8]string{
		"LLLLLLLL",
		"LBBBBBBB",
		"LBDDBBBB",
		"LBDDBDDB",
		"LBBBBBDB",
		"LBBDDBBB",
		"LBBDDBBB",
		"DDDDDDDD",
	}

	for y := 0; y < 8; y++ {
		row := grid[y]
		for x := 0; x < 8; x++ {
			var c loommath.Color
			switch row[x] {
			case 'L':
				c = light
			case 'B':
				c = base
			case 'D':
				c = dark
			}
			ctx.SetPixel(baseX+x, baseY+y, c)
		}
	}
}

func DrawMario(ctx *termrenderer.TermRenderer, pos loommath.Vec2, velocity loommath.Vec2, facingLeft bool, walkCycle bool) {
	// Premium Vermilion Red, Cobalt Blue, Peach Skin, and Golden NES tones
	red := loommath.Color{R: 0.92, G: 0.16, B: 0.08, A: 1.0}
	blue := loommath.Color{R: 0.0, G: 0.35, B: 0.82, A: 1.0}
	skin := loommath.Color{R: 0.98, G: 0.76, B: 0.62, A: 1.0}
	brown := loommath.Color{R: 0.45, G: 0.20, B: 0.05, A: 1.0}
	yellow := loommath.Color{R: 0.98, G: 0.78, B: 0.05, A: 1.0}

	baseX := int(math.Round(float64(pos.X)))
	baseY := int(math.Round(float64(pos.Y)))

	// Perfect classic 6x8 NES Mario grid layout
	grid := [8]string{
		"..RRR.", // Hat top
		".RRRRR", // Hat brim (visor to the right)
		".OOSKS", // Hair, Face, Eye, Nose (visor above nose)
		".OSOOD", // Hair, Face, Mustache, Chin/Mustache end
		"..RBR.", // Red shirt, Blue overalls strap, Red shirt
		".RBBBR", // Red sleeve, Blue overalls chest, Red sleeve
		"..BBB.", // Blue overalls legs
		".OO.OO", // Brown boots (feet apart by default)
	}

	// Double check if mustache character is brown or not
	// We'll map 'D' to brown (mustache/hair/boots) as well.

	// Simple walking animation (feet alternate)
	isMoving := math.Abs(float64(velocity.X)) > 10.0

	for y := 0; y < 8; y++ {
		row := grid[y]
		if y == 7 && isMoving && walkCycle {
			row = "..OO.." // feet together
		}
		for x := 0; x < 6; x++ {
			charCol := row[x]
			if facingLeft {
				// Mirror horizontally
				charCol = row[5-x]
			}
			
			var c loommath.Color
			switch charCol {
			case 'R':
				c = red
			case 'B':
				c = blue
			case 'S':
				c = skin
			case 'O', 'D':
				c = brown
			case 'Y':
				c = yellow
			case 'K':
				c = loommath.ColorBlack
			default:
				continue // transparent sky
			}
			ctx.SetPixel(baseX+x, baseY+y, c)
		}
	}
}

func DrawGoomba(ctx *termrenderer.TermRenderer, pos loommath.Vec2, alive bool, walkCycle bool) {
	brown := loommath.Color{R: 0.65, G: 0.30, B: 0.10, A: 1.0}
	dark := loommath.Color{R: 0.35, G: 0.15, B: 0.05, A: 1.0}
	skin := loommath.Color{R: 0.96, G: 0.75, B: 0.58, A: 1.0}
	black := loommath.Color{R: 0, G: 0, B: 0, A: 1.0}

	baseX := int(math.Round(float64(pos.X)))
	baseY := int(math.Round(float64(pos.Y)))

	if !alive {
		// Squished Goomba
		grid := [6]string{
			"......",
			"......",
			"......",
			"......",
			".OOOO.",
			"DDDDDD",
		}
		drawGoombaGrid(ctx, baseX, baseY, grid, brown, dark, skin, black)
		return
	}

	// 6x6 pixel Goomba grid with stepping feet animation (steps outward vs inward)
	grid := [6]string{
		"..OO..", // Cap top
		".OOOO.", // Cap middle
		"OOOOOO", // Cap bottom
		"OKOOKO", // Eyes
		"OOSSOO", // Mouth/stem
		"DD..DD", // Feet apart
	}

	// Walking animation (feet step in and out)
	if walkCycle {
		grid[4] = ".OSSO."
		grid[5] = ".DDDD." // Feet in
	}

	drawGoombaGrid(ctx, baseX, baseY, grid, brown, dark, skin, black)
}

func drawGoombaGrid(ctx *termrenderer.TermRenderer, baseX, baseY int, grid [6]string, brown, dark, skin, black loommath.Color) {
	for y := 0; y < 6; y++ {
		row := grid[y]
		for x := 0; x < 6; x++ {
			charCol := row[x]
			var c loommath.Color
			switch charCol {
			case 'O':
				c = brown
			case 'D':
				c = dark
			case 'S':
				c = skin
			case 'K':
				c = black
			default:
				continue
			}
			ctx.SetPixel(baseX+x, baseY+y, c)
		}
	}
}

func DrawBrick(ctx *termrenderer.TermRenderer, pos loommath.Vec2) {
	orange := loommath.Color{R: 0.78, G: 0.32, B: 0.12, A: 1.0}
	dark := loommath.Color{R: 0.38, G: 0.12, B: 0.05, A: 1.0}
	light := loommath.Color{R: 0.90, G: 0.55, B: 0.35, A: 1.0}

	baseX := int(math.Round(float64(pos.X)))
	baseY := int(math.Round(float64(pos.Y)))

	// Perfect 8x8 pixel grid for brick tile
	grid := [8]string{
		"DDDDDDDD",
		"LLLLDLLL",
		"OOOODOOO",
		"DDDDDDDD",
		"LLDLLLLD",
		"OODOOOOD",
		"OODOOOOD",
		"DDDDDDDD",
	}

	for y := 0; y < 8; y++ {
		row := grid[y]
		for x := 0; x < 8; x++ {
			var c loommath.Color
			switch row[x] {
			case 'L':
				c = light
			case 'O':
				c = orange
			case 'D':
				c = dark
			}
			ctx.SetPixel(baseX+x, baseY+y, c)
		}
	}
}

func DrawQuestionBlock(ctx *termrenderer.TermRenderer, pos loommath.Vec2, empty bool, timer float32) {
	baseX := int(math.Round(float64(pos.X)))
	baseY := int(math.Round(float64(pos.Y)))

	if empty {
		base := loommath.Color{R: 0.6, G: 0.4, B: 0.3, A: 1.0}
		dark := loommath.Color{R: 0.3, G: 0.2, B: 0.1, A: 1.0}
		grid := [8]string{
			"DDDDDDDD",
			"DBBBBBBD",
			"DBKBBKBD",
			"DBBBBBBD",
			"DBBBBBBD",
			"DBKBBKBD",
			"DBBBBBBD",
			"DDDDDDDD",
		}
		for y := 0; y < 8; y++ {
			row := grid[y]
			for x := 0; x < 8; x++ {
				var c loommath.Color
				switch row[x] {
				case 'B':
					c = base
				case 'D':
					c = dark
				case 'K':
					c = dark // rivets
				}
				ctx.SetPixel(baseX+x, baseY+y, c)
			}
		}
		return
	}

	pulse := float32(math.Sin(float64(timer)*8.0))*0.15 + 0.85
	yellow := loommath.Color{R: 0.95 * pulse, G: 0.72 * pulse, B: 0.05 * pulse, A: 1.0}
	dark := loommath.Color{R: 0.60, G: 0.32, B: 0.02, A: 1.0}
	qColor := loommath.Color{R: 0.8, G: 0.4, B: 0.1, A: 1}

	// Perfect 8x8 pixel grid for question block with a beautiful center question mark
	grid := [8]string{
		"DDDDDDDD",
		"DYQQQYYD",
		"DYQYYQYD",
		"DYYYQYYD",
		"DYYQQYYD",
		"DYYYYYYD",
		"DYYQQYYD",
		"DDDDDDDD",
	}

	for y := 0; y < 8; y++ {
		row := grid[y]
		for x := 0; x < 8; x++ {
			var c loommath.Color
			switch row[x] {
			case 'Y':
				c = yellow
			case 'D':
				c = dark
			case 'Q':
				c = qColor
			}
			ctx.SetPixel(baseX+x, baseY+y, c)
		}
	}
}

func DrawPipe(ctx *termrenderer.TermRenderer, pos loommath.Vec2, width, height float32) {
	pipeColor := loommath.Color{R: 0.0, G: 0.68, B: 0.08, A: 1.0}
	highlight1 := loommath.Color{R: 0.35, G: 0.85, B: 0.22, A: 1.0}
	highlight2 := loommath.Color{R: 0.65, G: 0.95, B: 0.45, A: 1.0}
	dark := loommath.Color{R: 0.0, G: 0.42, B: 0.02, A: 1.0}
	black := loommath.Color{R: 0, G: 0, B: 0, A: 1}

	baseX := int(math.Round(float64(pos.X)))
	baseY := int(math.Round(float64(pos.Y)))

	ow := int(math.Round(float64(width)))
	oh := int(math.Round(float64(height)))

	drawRect := func(rx, ry, rw, rh float32, c loommath.Color) {
		ctx.Rect(loommath.Vec2{X: float32(baseX) + rx, Y: float32(baseY) + ry}, loommath.Vec2{X: rw, Y: rh}, c)
	}

	// Outline
	drawRect(-1, -1, float32(ow) + 2, float32(oh) + 1, black)

	// Body
	drawRect(1, 4, float32(ow) - 2, float32(oh) - 4, pipeColor)
	drawRect(2, 4, 2, float32(oh) - 4, highlight1)
	drawRect(3, 4, 1, float32(oh) - 4, highlight2)
	drawRect(float32(ow) - 3, 4, 2, float32(oh) - 4, dark)

	// Rim
	drawRect(0, 0, float32(ow), 4, pipeColor)
	drawRect(2, 0, 2, 4, highlight1)
	drawRect(3, 0, 1, 4, highlight2)
	drawRect(float32(ow) - 3, 0, 2, 4, dark)
	
	// Inner shadow
	drawRect(1, 0, float32(ow) - 2, 1, black)
}

func DrawCoin(ctx *termrenderer.TermRenderer, pos loommath.Vec2, timer float32) {
	spin := float32(math.Abs(math.Sin(float64(timer)*4.0)))
	width := 24.0 * spin
	if width < 4 { width = 4 }

	baseX := int(math.Round(float64(pos.X)))
	baseY := int(math.Round(float64(pos.Y)))

	// Coin size is 40x40. Width is scaled by 5.33.
	cw := int(math.Round(float64(width / 5.33)))
	if cw < 1 { cw = 1 }

	cx := baseX + (int(math.Round(40.0/5.33)) - cw)/2
	cy := baseY + int(math.Round(6.0/5.33))
	
	ctx.Rect(loommath.Vec2{X: float32(cx), Y: float32(cy)}, loommath.Vec2{X: float32(cw), Y: float32(int(math.Round(28.0/5.33)))}, loommath.Color{R: 0.98, G: 0.78, B: 0.05, A: 1.0})
	
	if cw > 1 {
		ctx.Rect(loommath.Vec2{X: float32(cx + cw/2), Y: float32(cy + 2)}, loommath.Vec2{X: 1, Y: 2}, loommath.Color{R: 0.60, G: 0.32, B: 0.02, A: 1.0})
	}
}

func DrawFlag(ctx *termrenderer.TermRenderer, pos loommath.Vec2) {
	baseX := int(math.Round(float64(pos.X)))
	baseY := int(math.Round(float64(pos.Y)))

	// Beautiful retro 8x8 triangular flag banner (NES styling)
	grid := [8]string{
		"WWWWWW..",
		"WWGGWW..",
		"WGGGGW..",
		"WWGGWW..",
		"WWWW....",
		"WWW.....",
		"WW......",
		"W.......",
	}

	white := loommath.ColorWhite
	green := loommath.Color{R: 0.0, G: 0.68, B: 0.08, A: 1.0}

	for y := 0; y < 8; y++ {
		row := grid[y]
		for x := 0; x < 8; x++ {
			var c loommath.Color
			switch row[x] {
			case 'W':
				c = white
			case 'G':
				c = green
			default:
				continue
			}
			ctx.SetPixel(baseX+x, baseY+y, c)
		}
	}
}

func DrawFlagpole(ctx *termrenderer.TermRenderer, pos loommath.Vec2) {
	baseX := int(math.Round(float64(pos.X)))
	baseY := int(math.Round(float64(pos.Y)))

	// A light grey vertical pole line (width 1 pixel, height 8 pixels)
	// with a dark shadow on the right for cylinder effect
	poleColor := loommath.Color{R: 0.75, G: 0.75, B: 0.75, A: 1.0}
	shadowColor := loommath.Color{R: 0.35, G: 0.35, B: 0.35, A: 1.0}

	// Centered in the 8-pixel grid block
	for y := 0; y < 8; y++ {
		ctx.SetPixel(baseX+3, baseY+y, poleColor)
		ctx.SetPixel(baseX+4, baseY+y, shadowColor)
	}
}

func DrawPixelText(ctx *termrenderer.TermRenderer, text string, baseX, baseY int, color loommath.Color) {
	DrawPixelTextScale(ctx, text, baseX, baseY, 1, color)
}

func DrawPixelTextScale(ctx *termrenderer.TermRenderer, text string, baseX, baseY int, scale int, color loommath.Color) {
	for idx, char := range text {
		glyph, ok := Font[char]
		if !ok {
			glyph = Font[' ']
		}
		
		bx := baseX + idx*6*scale
		for gy := 0; gy < 7; gy++ {
			row := glyph[gy]
			for gx := 0; gx < 5; gx++ {
				if gx < len(row) && row[gx] == '*' {
					if scale == 1 {
						ctx.SetPixel(bx+gx, baseY+gy, color)
					} else {
						for sy := 0; sy < scale; sy++ {
							for sx := 0; sx < scale; sx++ {
								ctx.SetPixel(bx+gx*scale+sx, baseY+gy*scale+sy, color)
							}
						}
					}
				}
			}
		}
	}
}

var Font = map[rune][]string{
	'0': {` *** `, `*   *`, `*  **`, `* * *`, `**  *`, `*   *`, ` *** `},
	'1': {`  *  `, ` **  `, `  *  `, `  *  `, `  *  `, `  *  `, ` *** `},
	'2': {` *** `, `*   *`, `    *`, `   * `, `  *  `, ` *   `, `*****`},
	'3': {`*****`, `    *`, `   * `, `  ** `, `    *`, `*   *`, ` *** `},
	'4': {`*   *`, `*   *`, `*   *`, `*****`, `    *`, `    *`, `    *`},
	'5': {`*****`, `*    `, `**** `, `    *`, `    *`, `*   *`, ` *** `},
	'6': {` *** `, `*    `, `**** `, `*   *`, `*   *`, `*   *`, ` *** `},
	'7': {`*****`, `    *`, `   * `, `  *  `, ` *   `, ` *   `, ` *   `},
	'8': {` *** `, `*   *`, `*   *`, ` *** `, `*   *`, `*   *`, ` *** `},
	'9': {` *** `, `*   *`, `*   *`, ` ****`, `    *`, `    *`, ` *** `},
	'A': {` *** `, `*   *`, `*   *`, `*****`, `*   *`, `*   *`, `*   *`},
	'B': {`**** `, `*   *`, `*   *`, `**** `, `*   *`, `*   *`, `**** `},
	'C': {` ****`, `*    `, `*    `, `*    `, `*    `, `*    `, ` ****`},
	'D': {`**** `, `*   *`, `*   *`, `*   *`, `*   *`, `*   *`, `**** `},
	'E': {`*****`, `*    `, `**** `, `*    `, `*    `, `*    `, `*****`},
	'F': {`*****`, `*    `, `**** `, `*    `, `*    `, `*    `, `*    `},
	'G': {` *** `, `*   *`, `*    `, `* ***`, `*   *`, `*   *`, ` *** `},
	'H': {`*   *`, `*   *`, `*   *`, `*****`, `*   *`, `*   *`, `*   *`},
	'I': {`*****`, `  *  `, `  *  `, `  *  `, `  *  `, `  *  `, `*****`},
	'J': {`  ***`, `   * `, `   * `, `   * `, `   * `, `*  * `, ` **  `},
	'K': {`*  * `, `* *  `, `**   `, `***  `, `** * `, `*  * `, `*   *`},
	'L': {`*    `, `*    `, `*    `, `*    `, `*    `, `*    `, `*****`},
	'M': {`*   *`, `** **`, `* * *`, `*   *`, `*   *`, `*   *`, `*   *`},
	'N': {`*   *`, `**  *`, `* * *`, `*  **`, `*   *`, `*   *`, `*   *`},
	'O': {` *** `, `*   *`, `*   *`, `*   *`, `*   *`, `*   *`, ` *** `},
	'P': {`**** `, `*   *`, `*   *`, `**** `, `*    `, `*    `, `*    `},
	'Q': {` *** `, `*   *`, `*   *`, `*   *`, `* * *`, `*  * `, ` **  *`},
	'R': {`**** `, `*   *`, `*   *`, `**** `, `* *  `, `*  * `, `*   *`},
	'S': {` ****`, `*    `, `*    `, ` *** `, `    *`, `    *`, `**** `},
	'T': {`*****`, `  *  `, `  *  `, `  *  `, `  *  `, `  *  `, `  *  `},
	'U': {`*   *`, `*   *`, `*   *`, `*   *`, `*   *`, `*   *`, ` *** `},
	'V': {`*   *`, `*   *`, `*   *`, `*   *`, `*   *`, ` * * `, `  *  `},
	'W': {`*   *`, `*   *`, `*   *`, `* * *`, `** **`, `*   *`, `*   *`},
	'X': {`*   *`, `*   *`, ` * * `, `  *  `, ` * * `, `*   *`, `*   *`},
	'Y': {`*   *`, `*   *`, ` * * `, `  *  `, `  *  `, `  *  `, `  *  `},
	'Z': {`*****`, `    *`, `   * `, `  *  `, ` *   `, `*    `, `*****`},
	'-': {`     `, `     `, `     `, `*****`, `     `, `     `, `     `},
	'/': {`    *`, `   * `, `  *  `, ` *   `, ` *   `, `*    `, `*    `},
	':': {`     `, ` **  `, ` **  `, `     `, ` **  `, ` **  `, `     `},
	'.': {`     `, `     `, `     `, `     `, `     `, ` **  `, ` **  `},
	'+': {`     `, `  *  `, `  *  `, `*****`, `  *  `, `  *  `, `     `},
	' ': {`     `, `     `, `     `, `     `, `     `, `     `, `     `},
}

func DrawSmallPixelText(ctx *termrenderer.TermRenderer, text string, baseX, baseY int, color loommath.Color) {
	for idx, char := range text {
		glyph, ok := SmallFont[char]
		if !ok {
			glyph = SmallFont[' ']
		}
		
		bx := baseX + idx*4 // 3 pixels wide + 1 pixel spacing
		for gy := 0; gy < 5; gy++ {
			row := glyph[gy]
			for gx := 0; gx < 3; gx++ {
				if gx < len(row) && row[gx] == '*' {
					ctx.SetPixel(bx+gx, baseY+gy, color)
				}
			}
		}
	}
}

var SmallFont = map[rune][]string{
	'0': {`***`, `* *`, `* *`, `* *`, `***`},
	'1': {` * `, `** `, ` * `, ` * `, `***`},
	'2': {`***`, `  *`, `***`, `*  `, `***`},
	'3': {`***`, `  *`, `***`, `  *`, `***`},
	'4': {`* *`, `* *`, `***`, `  *`, `  *`},
	'5': {`***`, `*  `, `***`, `  *`, `***`},
	'6': {`***`, `*  `, `***`, `* *`, `***`},
	'7': {`***`, `  *`, `  *`, `  *`, `  *`},
	'8': {`***`, `* *`, `***`, `* *`, `***`},
	'9': {`***`, `* *`, `***`, `  *`, `***`},
	'M': {`* *`, `***`, `* *`, `* *`, `* *`},
	'C': {`***`, `*  `, `*  `, `*  `, `***`},
	'x': {`   `, `* *`, ` * `, `* *`, `   `},
	'-': {`   `, `   `, `***`, `   `, `   `},
	' ': {`   `, `   `, `   `, `   `, `   `},
}
