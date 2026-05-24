package main

import (
	"math"

	"github.com/go-gl/glfw/v3.3/glfw"
	loommath "github.com/ujjwalvivek/loom/math"
	"github.com/ujjwalvivek/loom/renderer"
)

// Deterministic 2D hash function for procedural star placement
func hash2D(x, y int) float64 {
	val := math.Sin(float64(x)*12.9898 + float64(y)*78.233) * 43758.5453123
	return val - math.Floor(val)
}

// DrawTriangle draws a stylized triangle (useful for pine trees/mountains)
func DrawTriangle(ctx *renderer.RenderSystem, x, y, width, height float32, color loommath.Color) {
	layers := int(height / 4)
	if layers <= 0 { return }
	for i := 0; i < layers; i++ {
		w := width * float32(i) / float32(layers)
		ctx.Rect(loommath.Vec2{X: x - w/2, Y: y + float32(i*4)}, loommath.Vec2{X: w, Y: 4}, color)
	}
}

// DrawFacetedTriangle draws a triangle split down the middle with shadow and highlight colors, supporting optional detailing textures
func DrawFacetedTriangle(ctx *renderer.RenderSystem, x, y, width, height float32, leftColor, rightColor loommath.Color, detailingStyle int) {
	layers := int(height / 4)
	if layers <= 0 { return }
	for i := 0; i < layers; i++ {
		w := width * float32(i) / float32(layers)
		halfW := w / 2
		if halfW > 0 {
			lCol := leftColor
			rCol := rightColor
			yOffset := float32(i * 4)

			if detailingStyle == 1 {
				// Mountain rock strata detailing (horizontal bands and crevices)
				strata := float32(math.Sin(float64(yOffset)*0.15)*0.08 + 0.95)
				if int(yOffset)%24 < 4 {
					strata *= 0.85
				}
				lCol.R *= strata; lCol.G *= strata; lCol.B *= strata
				rCol.R *= strata; rCol.G *= strata; rCol.B *= strata
			} else if detailingStyle == 2 {
				// Snow drifts detailing (fine ice/snow density variations)
				strata := float32(math.Sin(float64(yOffset)*0.25)*0.06 + 0.97)
				if int(yOffset)%16 < 4 {
					strata *= 0.90
				}
				lCol.R *= strata; lCol.G *= strata; lCol.B *= strata
				rCol.R *= strata; rCol.G *= strata; rCol.B *= strata
			} else if detailingStyle == 3 {
				// Pine foliage tiered branch layers detailing (layered tips/shadows)
				if int(yOffset)%16 < 4 {
					// Base shadow of branch tier
					lCol.R *= 0.75; lCol.G *= 0.75; lCol.B *= 0.75
					rCol.R *= 0.75; rCol.G *= 0.75; rCol.B *= 0.75
				} else if int(yOffset)%16 >= 12 {
					// Bright tip of branch tier
					lCol.R *= 1.25; lCol.G *= 1.25; lCol.B *= 1.25
					rCol.R *= 1.25; rCol.G *= 1.25; rCol.B *= 1.25
				}
			}
			
			// Left half (shadow)
			ctx.Rect(loommath.Vec2{X: x - halfW, Y: y + yOffset}, loommath.Vec2{X: halfW, Y: 4}, lCol)
			// Right half (highlight)
			ctx.Rect(loommath.Vec2{X: x, Y: y + yOffset}, loommath.Vec2{X: halfW, Y: 4}, rCol)
		}
	}
}

// DrawBackground renders the parallax environment, including sky gradients, clouds, and topography.
func DrawBackground(ctx *renderer.RenderSystem, camX float32) {
	// Render a stunning dusk/sunset sky gradient
	for i := 0; i < 36; i++ {
		t := float32(i) / 36.0
		var r, g, b float32
		if t < 0.5 {
			factor := t * 2.0
			r = 0.08 + factor*(0.18-0.08)
			g = 0.06 + factor*(0.12-0.06)
			b = 0.16 + factor*(0.24-0.16)
		} else {
			factor := (t - 0.5) * 2.0
			r = 0.18 + factor*(0.35-0.18)
			g = 0.12 + factor*(0.18-0.12)
			b = 0.24 + factor*(0.28-0.24)
		}
		y := float32(i) * 20.0
		ctx.Rect(loommath.Vec2{X: camX, Y: y}, loommath.Vec2{X: 1280, Y: 20}, loommath.Color{R: r, G: g, B: b, A: 1.0})
	}

	// Render procedurally generated sparkling stars in the sky region
	// Stars scroll extremely slowly (e.g., parallax factor of 0.02)
	pStars := camX * 0.02
	cellSize := 45
	startCellX := int(pStars / float32(cellSize))
	endCellX := startCellX + (640 / cellSize) + 2
	
	// We want stars in the sky region: Y world coordinates from 360 to 520
	startCellY := 360 / cellSize // 360 / 45 = 8
	endCellY := 520 / cellSize   // 520 / 45 = 11

	for cx := startCellX - 1; cx < endCellX; cx++ {
		for cy := startCellY; cy <= endCellY; cy++ {
			h := hash2D(cx, cy)
			if h < 0.22 { // Star density
				offsetX := float32(hash2D(cx+100, cy) * float64(cellSize))
				offsetY := float32(hash2D(cx, cy+100) * float64(cellSize))
				
				sx := float32(cx)*float32(cellSize) + offsetX + (camX - pStars)
				sy := float32(cy)*float32(cellSize) + offsetY
				
				// Rapid sparkling twinkle effect using glfw.GetTime()
				tTime := float32(glfw.GetTime())
				twinkleTimer := float64(tTime)*8.0 + h*100.0
				twinkle := float32(math.Sin(twinkleTimer)*0.5 + 0.5) // 0.0 to 1.0
				
				var starColor loommath.Color
				if h < 0.07 {
					starColor = loommath.Color{R: 1.0, G: 0.95, B: 0.7, A: twinkle} // Warm gold star
				} else if h < 0.14 {
					starColor = loommath.Color{R: 0.75, G: 0.9, B: 1.0, A: twinkle} // Cool cyan star
				} else {
					starColor = loommath.Color{R: 1.0, G: 1.0, B: 1.0, A: twinkle} // White star
				}
				
				// Draw tiny sparkle cross arms if the star is currently bright/flashing
				if twinkle > 0.6 {
					flareColor := starColor
					flareColor.A *= 0.4
					ctx.Rect(loommath.Vec2{X: sx - 1, Y: sy}, loommath.Vec2{X: 1, Y: 1}, flareColor)
					ctx.Rect(loommath.Vec2{X: sx + 1, Y: sy}, loommath.Vec2{X: 1, Y: 1}, flareColor)
					ctx.Rect(loommath.Vec2{X: sx, Y: sy - 1}, loommath.Vec2{X: 1, Y: 1}, flareColor)
					ctx.Rect(loommath.Vec2{X: sx, Y: sy + 1}, loommath.Vec2{X: 1, Y: 1}, flareColor)
				}
				
				// Core pixel (always 1x1)
				ctx.Rect(loommath.Vec2{X: sx, Y: sy}, loommath.Vec2{X: 1, Y: 1}, starColor)
			}
		}
	}

	// Render parallax clouds (lowered to Y around 400 so they are visible in the sky)
	cloudWhite := loommath.Color{R: 1, G: 1, B: 1, A: 0.9}
	cloudShadow := loommath.Color{R: 0.8, G: 0.85, B: 0.9, A: 0.9}

	p3 := camX * 0.08
	startCol3 := int(p3 / 300)
	for i := startCol3 - 1; i < startCol3+8; i++ {
		cx := float32(i)*300.0 + (camX - p3)
		cy := float32(400 + math.Sin(float64(i*5))*40)
		
		// Shadow
		ctx.Rect(loommath.Vec2{X: cx + 20, Y: cy + 4}, loommath.Vec2{X: 80, Y: 30}, cloudShadow)
		ctx.Rect(loommath.Vec2{X: cx + 40, Y: cy - 20 + 4}, loommath.Vec2{X: 40, Y: 40}, cloudShadow)
		ctx.Rect(loommath.Vec2{X: cx, Y: cy + 15 + 4}, loommath.Vec2{X: 120, Y: 20}, cloudShadow)
		
		// Highlight
		ctx.Rect(loommath.Vec2{X: cx + 20, Y: cy}, loommath.Vec2{X: 80, Y: 30}, cloudWhite)
		ctx.Rect(loommath.Vec2{X: cx + 40, Y: cy - 20}, loommath.Vec2{X: 40, Y: 40}, cloudWhite)
		ctx.Rect(loommath.Vec2{X: cx, Y: cy + 15}, loommath.Vec2{X: 120, Y: 20}, cloudWhite)
	}

	// Render distant topography (mountains) with low-poly faceted shading and snow caps
	mount1Left := loommath.Color{R: 0.15, G: 0.10, B: 0.25, A: 1.0}
	mount1Right := loommath.Color{R: 0.22, G: 0.15, B: 0.32, A: 1.0}
	mount2Left := loommath.Color{R: 0.10, G: 0.08, B: 0.18, A: 1.0}
	mount2Right := loommath.Color{R: 0.16, G: 0.12, B: 0.24, A: 1.0}

	snow1Left := loommath.Color{R: 0.70, G: 0.65, B: 0.75, A: 1.0}
	snow1Right := loommath.Color{R: 0.90, G: 0.85, B: 0.92, A: 1.0}
	snow2Left := loommath.Color{R: 0.55, G: 0.50, B: 0.60, A: 1.0}
	snow2Right := loommath.Color{R: 0.75, G: 0.70, B: 0.80, A: 1.0}

	pMountains := camX * 0.20
	startMountain := int(pMountains / 400)
	for i := startMountain - 1; i < startMountain+8; i++ {
		mx := float32(i)*400.0 + (camX - pMountains)
		
		// Mountain 1 (back layer, larger)
		DrawFacetedTriangle(ctx, mx+200, 300, 400, 380, mount1Left, mount1Right, 1)
		DrawFacetedTriangle(ctx, mx+200, 300, 95, 90, snow1Left, snow1Right, 2) // Snow peak
		// Snow cap tongues/drippings for glacier-cut bottom edge
		DrawFacetedTriangle(ctx, mx+170, 390, 25, 30, snow1Left, snow1Left, 2)
		DrawFacetedTriangle(ctx, mx+200, 390, 35, 45, snow1Left, snow1Right, 2)
		DrawFacetedTriangle(ctx, mx+220, 390, 20, 25, snow1Right, snow1Right, 2)
		
		// Handcrafted snow drift textures on Mountain 1
		ctx.Rect(loommath.Vec2{X: mx+175, Y: 350}, loommath.Vec2{X: 12, Y: 4}, snow1Right)
		ctx.Rect(loommath.Vec2{X: mx+190, Y: 330}, loommath.Vec2{X: 8, Y: 4}, snow1Right)
		ctx.Rect(loommath.Vec2{X: mx+165, Y: 370}, loommath.Vec2{X: 15, Y: 4}, snow1Right)
		ctx.Rect(loommath.Vec2{X: mx+210, Y: 340}, loommath.Vec2{X: 16, Y: 4}, snow1Left)
		ctx.Rect(loommath.Vec2{X: mx+230, Y: 360}, loommath.Vec2{X: 12, Y: 4}, snow1Left)
		ctx.Rect(loommath.Vec2{X: mx+215, Y: 375}, loommath.Vec2{X: 20, Y: 4}, snow1Left)

		// Mountain 2 (front layer, smaller)
		DrawFacetedTriangle(ctx, mx+400, 400, 300, 280, mount2Left, mount2Right, 1)
		DrawFacetedTriangle(ctx, mx+400, 400, 75, 70, snow2Left, snow2Right, 2) // Snow peak
		// Snow cap tongues/drippings
		DrawFacetedTriangle(ctx, mx+380, 470, 20, 20, snow2Left, snow2Left, 2)
		DrawFacetedTriangle(ctx, mx+400, 470, 25, 35, snow2Left, snow2Right, 2)
		DrawFacetedTriangle(ctx, mx+415, 470, 15, 18, snow2Right, snow2Right, 2)

		// Handcrafted snow drift textures on Mountain 2
		ctx.Rect(loommath.Vec2{X: mx+385, Y: 435}, loommath.Vec2{X: 8, Y: 4}, snow2Right)
		ctx.Rect(loommath.Vec2{X: mx+375, Y: 450}, loommath.Vec2{X: 10, Y: 4}, snow2Right)
		ctx.Rect(loommath.Vec2{X: mx+410, Y: 430}, loommath.Vec2{X: 10, Y: 4}, snow2Left)
		ctx.Rect(loommath.Vec2{X: mx+420, Y: 450}, loommath.Vec2{X: 12, Y: 4}, snow2Left)
	}

	// Render foreground parallax flora (pine trees) with shaded trunks and detailed tiered foliage
	trunkShadow := loommath.Color{R: 0.18, G: 0.10, B: 0.08, A: 1.0}
	trunkHighlight := loommath.Color{R: 0.26, G: 0.16, B: 0.12, A: 1.0}
	
	pineGreenLeft := loommath.Color{R: 0.05, G: 0.20, B: 0.12, A: 1.0}
	pineGreenRight := loommath.Color{R: 0.08, G: 0.30, B: 0.18, A: 1.0}
	pineLightLeft := loommath.Color{R: 0.08, G: 0.28, B: 0.15, A: 1.0}
	pineLightRight := loommath.Color{R: 0.12, G: 0.38, B: 0.22, A: 1.0}

	pForest := camX * 0.38
	startForest := int(pForest / 150)
	for i := startForest - 1; i < startForest+15; i++ {
		fx := float32(i)*150.0 + (camX - pForest)
		fy := float32(500 + math.Sin(float64(i*9))*30)
		
		// Shaded trunk
		ctx.Rect(loommath.Vec2{X: fx - 8, Y: fy + 100}, loommath.Vec2{X: 8, Y: 80}, trunkShadow)
		ctx.Rect(loommath.Vec2{X: fx, Y: fy + 100}, loommath.Vec2{X: 8, Y: 80}, trunkHighlight)
		
		// Faceted and tiered pine foliage (Style 3)
		DrawFacetedTriangle(ctx, fx, fy+60, 140, 80, pineGreenLeft, pineGreenRight, 3)
		DrawFacetedTriangle(ctx, fx, fy+20, 110, 70, pineLightLeft, pineLightRight, 3)
		DrawFacetedTriangle(ctx, fx, fy-20, 80, 60, pineGreenLeft, pineGreenRight, 3)
	}
}

// DrawGroundBlock draws the classic orange-brown dirt blocks
func DrawGroundBlock(ctx *renderer.RenderSystem, pos loommath.Vec2) {
	base := loommath.Color{R: 0.85, G: 0.45, B: 0.25, A: 1.0}
	dark := loommath.Color{R: 0.55, G: 0.25, B: 0.1, A: 1.0}
	light := loommath.Color{R: 0.95, G: 0.65, B: 0.45, A: 1.0}

	// Base block
	ctx.Rect(pos, loommath.Vec2{X: float32(TileWidth), Y: float32(TileHeight)}, base)
	
	// Top light edge
	ctx.Rect(pos, loommath.Vec2{X: float32(TileWidth), Y: 4}, light)
	
	// Diagonal dark pattern mimicking the cracked ground
	ctx.Rect(loommath.Vec2{X: pos.X+4, Y: pos.Y+8}, loommath.Vec2{X: 8, Y: 8}, dark)
	ctx.Rect(loommath.Vec2{X: pos.X+24, Y: pos.Y+24}, loommath.Vec2{X: 8, Y: 8}, dark)
	ctx.Rect(loommath.Vec2{X: pos.X+24, Y: pos.Y+8}, loommath.Vec2{X: 4, Y: 4}, dark)
	ctx.Rect(loommath.Vec2{X: pos.X+4, Y: pos.Y+24}, loommath.Vec2{X: 4, Y: 4}, dark)
	
	// Dark bottom/right edge
	ctx.Rect(loommath.Vec2{X: pos.X, Y: pos.Y + float32(TileHeight) - 4}, loommath.Vec2{X: float32(TileWidth), Y: 4}, dark)
	ctx.Rect(loommath.Vec2{X: pos.X + float32(TileWidth) - 4, Y: pos.Y}, loommath.Vec2{X: 4, Y: float32(TileHeight)}, dark)
}

// DrawMario renders a stylized vector version of Mario built from block primitives
func DrawMario(ctx *renderer.RenderSystem, pos loommath.Vec2, velocity loommath.Vec2, facingLeft bool) {
	red := loommath.Color{R: 0.95, G: 0.1, B: 0.05, A: 1.0}
	blue := loommath.Color{R: 0.15, G: 0.25, B: 0.95, A: 1.0}
	skin := loommath.Color{R: 0.98, G: 0.8, B: 0.6, A: 1.0}
	brown := loommath.Color{R: 0.5, G: 0.25, B: 0.0, A: 1.0}
	yellow := loommath.Color{R: 0.95, G: 0.85, B: 0.1, A: 1.0}

	x := pos.X
	y := pos.Y
	
	// Simple animation bobbing
	walkBob := float32(0)
	if math.Abs(float64(velocity.X)) > 10 {
		walkBob = float32(math.Sin(float64(pos.X)*0.1)) * 2.0
	}

	drawBlock := func(bx, by, bw, bh float32, c loommath.Color) {
		if facingLeft {
			// Mirror horizontally around center 16
			bx = 32 - bx - bw
		}
		ctx.Rect(loommath.Vec2{X: x + bx, Y: y + by}, loommath.Vec2{X: bw, Y: bh}, c)
	}

	// Hat
	drawBlock(6, 0+walkBob, 18, 6, red)
	drawBlock(4, 6+walkBob, 26, 4, red) // Brim extends forward

	// Face & Hair
	drawBlock(4, 10+walkBob, 8, 12, brown) // Hair/ear
	drawBlock(12, 10+walkBob, 16, 12, skin) // Face
	drawBlock(28, 14+walkBob, 6, 6, skin) // Nose

	// Mustache
	drawBlock(20, 18+walkBob, 10, 4, brown)

	// Eye
	drawBlock(20, 10+walkBob, 4, 8, loommath.Color{R: 0, G: 0, B: 0, A: 1})

	// Shirt & Overalls
	drawBlock(8, 22, 16, 10, red) // Shirt
	drawBlock(12, 24, 8, 10, blue) // Overalls center
	drawBlock(8, 30, 16, 8, blue) // Overalls base

	// Buttons
	drawBlock(10, 26, 4, 4, yellow)
	drawBlock(18, 26, 4, 4, yellow)

	// Arms (swing while walking)
	armSwing := float32(0)
	if math.Abs(float64(velocity.X)) > 10 {
		armSwing = float32(math.Sin(float64(pos.X)*0.1)) * 4.0
	}
	drawBlock(2, 22-armSwing, 6, 10, red) // Back arm
	drawBlock(24, 22+armSwing, 6, 10, red) // Front arm

	// Hands
	drawBlock(0, 32-armSwing, 8, 6, skin)
	drawBlock(24, 32+armSwing, 8, 6, skin)

	// Boots (swing opposite to arms)
	drawBlock(4, 34+armSwing, 10, 6, brown)
	drawBlock(18, 34-armSwing, 10, 6, brown)
}

// DrawGoomba renders a classic lively Goomba using vector blocks
func DrawGoomba(ctx *renderer.RenderSystem, pos loommath.Vec2, alive bool) {
	brown := loommath.Color{R: 0.7, G: 0.35, B: 0.15, A: 1.0}
	skin := loommath.Color{R: 0.98, G: 0.8, B: 0.6, A: 1.0}
	black := loommath.Color{R: 0, G: 0, B: 0, A: 1.0}
	
	if !alive {
		// Squished
		ctx.Rect(loommath.Vec2{X: pos.X+4, Y: pos.Y + 24}, loommath.Vec2{X: 24, Y: 8}, brown)
		return
	}

	// Walk animation
	walkBob := float32(math.Sin(float64(pos.X)*0.2)) * 2.0
	
	// Head dome
	ctx.Rect(loommath.Vec2{X: pos.X + 8, Y: pos.Y + 4 + walkBob}, loommath.Vec2{X: 16, Y: 4}, brown)
	ctx.Rect(loommath.Vec2{X: pos.X + 4, Y: pos.Y + 8 + walkBob}, loommath.Vec2{X: 24, Y: 14}, brown)
	
	// Body
	ctx.Rect(loommath.Vec2{X: pos.X + 10, Y: pos.Y + 22}, loommath.Vec2{X: 12, Y: 6}, skin)

	// Eyes
	ctx.Rect(loommath.Vec2{X: pos.X + 6, Y: pos.Y + 12 + walkBob}, loommath.Vec2{X: 6, Y: 8}, black)
	ctx.Rect(loommath.Vec2{X: pos.X + 8, Y: pos.Y + 12 + walkBob}, loommath.Vec2{X: 2, Y: 4}, loommath.Color{R: 1, G: 1, B: 1, A: 1}) // Highlight
	ctx.Rect(loommath.Vec2{X: pos.X + 20, Y: pos.Y + 12 + walkBob}, loommath.Vec2{X: 6, Y: 8}, black)
	ctx.Rect(loommath.Vec2{X: pos.X + 22, Y: pos.Y + 12 + walkBob}, loommath.Vec2{X: 2, Y: 4}, loommath.Color{R: 1, G: 1, B: 1, A: 1}) // Highlight

	// Feet walking animation
	ctx.Rect(loommath.Vec2{X: pos.X + 2, Y: pos.Y + 28 - walkBob}, loommath.Vec2{X: 10, Y: 6}, black)
	ctx.Rect(loommath.Vec2{X: pos.X + 20, Y: pos.Y + 28 + walkBob}, loommath.Vec2{X: 10, Y: 6}, black)
}

// DrawBrick renders a lively classic Mario brick
func DrawBrick(ctx *renderer.RenderSystem, pos loommath.Vec2) {
	orange := loommath.Color{R: 0.85, G: 0.35, B: 0.1, A: 1.0}
	dark := loommath.Color{R: 0.4, G: 0.15, B: 0.05, A: 1.0}
	light := loommath.Color{R: 0.95, G: 0.6, B: 0.4, A: 1.0}

	// Base
	ctx.Rect(pos, loommath.Vec2{X: float32(TileWidth), Y: float32(TileHeight)}, dark)

	// 4 Sub-bricks
	bw := float32(18)
	bh := float32(18)
	
	drawSubBrick := func(bx, by float32) {
		ctx.Rect(loommath.Vec2{X: pos.X+bx, Y: pos.Y+by}, loommath.Vec2{X: bw, Y: bh}, orange)
		ctx.Rect(loommath.Vec2{X: pos.X+bx, Y: pos.Y+by}, loommath.Vec2{X: bw, Y: 2}, light)
		ctx.Rect(loommath.Vec2{X: pos.X+bx, Y: pos.Y+by}, loommath.Vec2{X: 2, Y: bh}, light)
	}

	drawSubBrick(2, 2)
	drawSubBrick(22, 2)
	// Staggered bottom row
	drawSubBrick(0, 22)
	drawSubBrick(20, 22)
}

// DrawQuestionBlock renders a flashing classic block
func DrawQuestionBlock(ctx *renderer.RenderSystem, pos loommath.Vec2, empty bool, timer float32) {
	if empty {
		base := loommath.Color{R: 0.6, G: 0.4, B: 0.3, A: 1.0}
		dark := loommath.Color{R: 0.3, G: 0.2, B: 0.1, A: 1.0}
		ctx.Rect(pos, loommath.Vec2{X: float32(TileWidth), Y: float32(TileHeight)}, dark)
		ctx.Rect(loommath.Vec2{X: pos.X+2, Y: pos.Y+2}, loommath.Vec2{X: float32(TileWidth)-4, Y: float32(TileHeight)-4}, base)
		
		// Rivets
		ctx.Rect(loommath.Vec2{X: pos.X+4, Y: pos.Y+4}, loommath.Vec2{X: 4, Y: 4}, dark)
		ctx.Rect(loommath.Vec2{X: pos.X+32, Y: pos.Y+4}, loommath.Vec2{X: 4, Y: 4}, dark)
		ctx.Rect(loommath.Vec2{X: pos.X+4, Y: pos.Y+32}, loommath.Vec2{X: 4, Y: 4}, dark)
		ctx.Rect(loommath.Vec2{X: pos.X+32, Y: pos.Y+32}, loommath.Vec2{X: 4, Y: 4}, dark)
		return
	}

	pulse := float32(math.Sin(float64(timer)*8.0))*0.15 + 0.85
	yellow := loommath.Color{R: 0.9 * pulse, G: 0.8 * pulse, B: 0.1 * pulse, A: 1.0}
	dark := loommath.Color{R: 0.7, G: 0.4, B: 0.05, A: 1.0}
	
	ctx.Rect(pos, loommath.Vec2{X: float32(TileWidth), Y: float32(TileHeight)}, dark)
	ctx.Rect(loommath.Vec2{X: pos.X+2, Y: pos.Y+2}, loommath.Vec2{X: float32(TileWidth)-4, Y: float32(TileHeight)-4}, yellow)
	
	// Rivets
	ctx.Rect(loommath.Vec2{X: pos.X+4, Y: pos.Y+4}, loommath.Vec2{X: 4, Y: 4}, dark)
	ctx.Rect(loommath.Vec2{X: pos.X+32, Y: pos.Y+4}, loommath.Vec2{X: 4, Y: 4}, dark)
	ctx.Rect(loommath.Vec2{X: pos.X+4, Y: pos.Y+32}, loommath.Vec2{X: 4, Y: 4}, dark)
	ctx.Rect(loommath.Vec2{X: pos.X+32, Y: pos.Y+32}, loommath.Vec2{X: 4, Y: 4}, dark)
	
	// Question mark
	qColor := loommath.Color{R: 0.8, G: 0.4, B: 0.1, A: 1}
	ctx.Rect(loommath.Vec2{X: pos.X + 14, Y: pos.Y + 8}, loommath.Vec2{X: 12, Y: 6}, qColor)
	ctx.Rect(loommath.Vec2{X: pos.X + 20, Y: pos.Y + 14}, loommath.Vec2{X: 6, Y: 6}, qColor)
	ctx.Rect(loommath.Vec2{X: pos.X + 16, Y: pos.Y + 20}, loommath.Vec2{X: 6, Y: 6}, qColor)
	ctx.Rect(loommath.Vec2{X: pos.X + 16, Y: pos.Y + 28}, loommath.Vec2{X: 6, Y: 6}, qColor)
}

// DrawPipe renders a lively classic pipe
func DrawPipe(ctx *renderer.RenderSystem, pos loommath.Vec2, width, height float32) {
	pipeColor := loommath.Color{R: 0.1, G: 0.7, B: 0.1, A: 1.0}
	highlight1 := loommath.Color{R: 0.4, G: 0.9, B: 0.3, A: 1.0}
	highlight2 := loommath.Color{R: 0.6, G: 1.0, B: 0.5, A: 1.0}
	dark := loommath.Color{R: 0.05, G: 0.4, B: 0.05, A: 1.0}
	black := loommath.Color{R: 0, G: 0, B: 0, A: 1}

	// Outline
	ctx.Rect(loommath.Vec2{X: pos.X-2, Y: pos.Y-2}, loommath.Vec2{X: width+4, Y: height+2}, black)

	// Body
	ctx.Rect(loommath.Vec2{X: pos.X + 2, Y: pos.Y + 20}, loommath.Vec2{X: width - 4, Y: height - 20}, pipeColor)
	ctx.Rect(loommath.Vec2{X: pos.X + 6, Y: pos.Y + 20}, loommath.Vec2{X: 12, Y: height - 20}, highlight1)
	ctx.Rect(loommath.Vec2{X: pos.X + 10, Y: pos.Y + 20}, loommath.Vec2{X: 4, Y: height - 20}, highlight2)
	ctx.Rect(loommath.Vec2{X: pos.X + width - 12, Y: pos.Y + 20}, loommath.Vec2{X: 8, Y: height - 20}, dark)

	// Rim
	ctx.Rect(pos, loommath.Vec2{X: width, Y: 20}, pipeColor)
	ctx.Rect(loommath.Vec2{X: pos.X + 6, Y: pos.Y}, loommath.Vec2{X: 12, Y: 20}, highlight1)
	ctx.Rect(loommath.Vec2{X: pos.X + 10, Y: pos.Y}, loommath.Vec2{X: 4, Y: 20}, highlight2)
	ctx.Rect(loommath.Vec2{X: pos.X + width - 12, Y: pos.Y}, loommath.Vec2{X: 8, Y: 20}, dark)
	
	// Inner hole shadow
	ctx.Rect(loommath.Vec2{X: pos.X + 4, Y: pos.Y}, loommath.Vec2{X: width - 8, Y: 4}, black)
}

// DrawCoin renders a spinning coin
func DrawCoin(ctx *renderer.RenderSystem, pos loommath.Vec2, timer float32) {
	spin := float32(math.Abs(math.Sin(float64(timer)*4.0)))
	width := 24.0 * spin
	if width < 4 { width = 4 }

	cx := pos.X + 20 - width/2
	cy := pos.Y + 6
	
	ctx.Rect(loommath.Vec2{X: cx, Y: cy}, loommath.Vec2{X: width, Y: 28}, loommath.Color{R: 1.0, G: 0.8, B: 0.1, A: 1.0})
	
	if width > 8 {
		ctx.Rect(loommath.Vec2{X: cx + width/2 - 2, Y: cy + 8}, loommath.Vec2{X: 4, Y: 12}, loommath.Color{R: 0.7, G: 0.5, B: 0.0, A: 1.0})
	}
}
