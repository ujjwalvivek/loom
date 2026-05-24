package main

import (
	"bytes"
	_ "embed"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/ujjwalvivek/loom/ecs"
	"github.com/ujjwalvivek/loom/engine"
	loommath "github.com/ujjwalvivek/loom/math"
	"github.com/ujjwalvivek/loom/renderer"
)

//go:embed icon.png
var iconBytes []byte

// iconBase64 is a fallback for embedding an icon as base64 text.
// To use it: drop a base64-encoded image file as "icon.txt" in this directory
// and uncomment the line below:
// //go:embed icon.txt
var iconBase64 string

func loadIcon() []image.Image {
	// Prefer a direct binary PNG, smaller and faster to decode.
	if len(iconBytes) > 0 {
		img, _, err := image.Decode(bytes.NewReader(iconBytes))
		if err == nil {
			return []image.Image{img}
		}
		fmt.Println("Warning: Failed to decode embedded icon.png, trying base64 fallback:", err)
	}

	// Fallback: base64-encoded image from icon.txt (useful for formats that
	// don't survive binary embedding, or for quick asset swapping without re-encoding).
	data := strings.Join(strings.Fields(iconBase64), "")
	if data == "" {
		fmt.Println("Warning: No window icon set (icon.png missing or corrupt, icon.txt empty)")
		return nil
	}
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		fmt.Println("Warning: Failed to decode embedded icon.txt (invalid base64):", err)
		return nil
	}
	img, _, err := image.Decode(bytes.NewReader(decoded))
	if err != nil {
		fmt.Println("Warning: Failed to decode image from icon.txt:", err)
		return nil
	}
	return []image.Image{img}
}

type GameState int

const (
	StateStartScreen GameState = 0
	StatePlaying     GameState = 1
	StateGameOver    GameState = 2
	StateWin         GameState = 3
)

type MarioScene struct {
	lvl            *Level
	player         *Player
	goombas        []*Goomba
	camX           float32
	coinAnimTimer  float32
	runAnimTimer   float32
	runAnimFrame   int
	levelTimer     float32
	gameOverDelay  float32
	state          GameState
}

func (s *MarioScene) Load(ctx *engine.Context) {
	rand.Seed(time.Now().UnixNano())

	s.lvl = NewLevel(ctx)

	s.player = SpawnPlayer(ctx, s.lvl.PlayerSpawn)
	s.goombas = make([]*Goomba, 0)
	for _, spawn := range s.lvl.GoombaSpawns {
		s.goombas = append(s.goombas, SpawnGoomba(ctx, spawn))
	}

	s.camX = 0
	s.levelTimer = 400.0
	s.gameOverDelay = 0.0
	s.state = StateStartScreen // Start on the Title Screen

	// Start BGM chiptune immediately on load so it plays on start screen
	SetupMarioAudio(ctx)
}

func (s *MarioScene) Update(ctx *engine.Context, dt float32) {
	switch s.state {
	case StateStartScreen:
		// Cinematic panning of the camera
		s.camX += dt * 80.0
		if s.camX > float32(GridCols*TileWidth-640) {
			s.camX = 0 // loop back
		}

		// Toggle flashing animation timer
		s.coinAnimTimer += dt
		if s.coinAnimTimer >= 0.4 {
			s.coinAnimTimer = 0.0
		}

		// Press Space or Enter to start the game
		if ctx.Input.KeyPressed(glfw.KeySpace) || ctx.Input.KeyPressed(glfw.KeyEnter) {
			s.state = StatePlaying
			s.camX = 0 // reset camera for player
		}
		return

	case StateGameOver:
		s.gameOverDelay -= dt
		if s.gameOverDelay <= 0 {
			s.Unload(ctx)
			s.Load(ctx)
			return
		}
		// Let gravity pull dead player down screen
		UpdateEntities(ctx, s.lvl, s.player, s.goombas, dt)
		return

	case StateWin:
		s.gameOverDelay += dt
		if s.gameOverDelay >= 4.0 {
			if ctx.Input.KeyPressed(glfw.KeySpace) || ctx.Input.KeyPressed(glfw.KeyEnter) {
				s.Unload(ctx)
				s.Load(ctx)
				return
			}
		}
		// Hold player static
		s.player.Velocity = loommath.Vec2{}
		ctx.Physics.SetVelocity(s.player.Body, s.player.Velocity)
		return

	case StatePlaying:
		// Decrement game time limit
		s.levelTimer -= dt * 1.5
		if s.levelTimer <= 0 {
			s.levelTimer = 0
			s.triggerDeath(ctx)
			return
		}

		// Read inputs
		leftPressed := ctx.Input.KeyDown(glfw.KeyA) || ctx.Input.KeyDown(glfw.KeyLeft)
		rightPressed := ctx.Input.KeyDown(glfw.KeyD) || ctx.Input.KeyDown(glfw.KeyRight)
		jumpPressed := ctx.Input.KeyDown(glfw.KeySpace) || ctx.Input.KeyDown(glfw.KeyW) || ctx.Input.KeyDown(glfw.KeyUp)
		shiftPressed := ctx.Input.KeyDown(glfw.KeyLeftShift) || ctx.Input.KeyDown(glfw.KeyRightShift)

		// Physics Tuning: Snappy controls
		accel := float32(2200.0)
		friction := float32(2800.0)
		maxSpeed := float32(220.0) // Walk speed

		if shiftPressed {
			maxSpeed = 340.0 // Run speed
			accel = 3200.0
		}

		// Horizontal Movement
		if rightPressed {
			if s.player.Velocity.X < 0 {
				s.player.Velocity.X += accel * 2.5 * dt // Skid turn boost
			} else {
				s.player.Velocity.X += accel * dt
			}
			if s.player.Velocity.X > maxSpeed {
				s.player.Velocity.X = maxSpeed
			}
			s.player.FacingLeft = false
		} else if leftPressed {
			if s.player.Velocity.X > 0 {
				s.player.Velocity.X -= accel * 2.5 * dt // Skid turn boost
			} else {
				s.player.Velocity.X -= accel * dt
			}
			if s.player.Velocity.X < -maxSpeed {
				s.player.Velocity.X = -maxSpeed
			}
			s.player.FacingLeft = true
		} else {
			// Fast friction stop on release
			if s.player.Velocity.X > 0 {
				s.player.Velocity.X -= friction * dt
				if s.player.Velocity.X < 0 {
					s.player.Velocity.X = 0
				}
			} else if s.player.Velocity.X < 0 {
				s.player.Velocity.X += friction * dt
				if s.player.Velocity.X > 0 {
					s.player.Velocity.X = 0
				}
			}
		}

		// Asymmetric Gravity: Fall faster than rising for snappy game feel
		gravity := float32(2800.0) // Rising gravity
		if s.player.Velocity.Y > 0 {
			gravity = 3800.0 // Snappy heavy falls
		}
		s.player.Velocity.Y += gravity * dt
		if s.player.Velocity.Y > 800.0 {
			s.player.Velocity.Y = 800.0 // terminal velocity
		}

		// Jump height variable holding acceleration
		if jumpPressed && s.player.IsGrounded {
			s.player.Velocity.Y = -760.0
			s.player.Jumping = true
			s.player.JumpHold = 0.0
			ctx.Audio.PlaySound(JumpPatch)
		} else if jumpPressed && s.player.Jumping {
			s.player.JumpHold += dt
			if s.player.JumpHold < 0.22 {
				// Counteract gravity to jump higher while holding
				s.player.Velocity.Y -= 900.0 * dt
			} else {
				s.player.Jumping = false
			}
		} else {
			s.player.Jumping = false
		}

		// Query latest coordinates from physics thread
		activeBodies := ctx.Physics.Query(loommath.Rect{X: s.camX - 300, Y: -1000, W: 1500, H: 2500})
		for _, b := range activeBodies {
			if b.Handle == s.player.Body {
				posVal := ctx.ECS.Get(s.player.Entity, ecs.TypePosition)
				if posVal != nil {
					p := posVal.(*ecs.Position)
					p.X = b.Pos.X
					p.Y = b.Pos.Y
				}
			}
			for _, g := range s.goombas {
				if b.Handle == g.Body && g.Alive {
					gPosVal := ctx.ECS.Get(g.Entity, ecs.TypePosition)
					if gPosVal != nil {
						gp := gPosVal.(*ecs.Position)
						gp.X = b.Pos.X
						gp.Y = b.Pos.Y
					}
				}
			}
		}

		// Tick entity loops
		s.goombas = UpdateEntities(ctx, s.lvl, s.player, s.goombas, dt)

		// Tick visual tile bounces
		s.lvl.Update(dt)

		// Check coin collections
		playerPosVal := ctx.ECS.Get(s.player.Entity, ecs.TypePosition)
		if playerPosVal != nil {
			p := playerPosVal.(*ecs.Position)
			px := int((p.X + 16) / TileWidth)
			py := int((p.Y + 19) / TileHeight)
			if px >= 0 && px < GridCols && py >= 0 && py < GridRows {
				tile := s.lvl.Grid[py][px]
				if tile.Type == TileCoin {
					tile.Type = TileAir
					s.player.Coins++
					s.player.Score += 200
					ctx.Audio.PlaySound(CoinPatch)
				}
			}

			// Falling into pit death
			if p.Y > 730.0 {
				s.triggerDeath(ctx)
				return
			}
		}

		// State check for death or win transition
		if !s.player.Alive {
			s.state = StateGameOver
			s.gameOverDelay = 3.0
			return
		}
		if s.player.Win {
			s.state = StateWin
			s.gameOverDelay = 0.0
			return
		}

		// Horizontal Camera scroll clamped to grid margins
		if playerPosVal != nil {
			p := playerPosVal.(*ecs.Position)
			targetCamX := p.X - 640.0/2.0
			if targetCamX < 0 {
				targetCamX = 0
			}
			maxCamX := float32(GridCols*TileWidth - 640)
			if targetCamX > maxCamX {
				targetCamX = maxCamX
			}
			s.camX += (targetCamX - s.camX) * 10.0 * dt
		}

		// Update animation cycles
		s.coinAnimTimer += dt
		if s.coinAnimTimer >= 0.4 {
			s.coinAnimTimer = 0.0
		}

		if float32(math.Abs(float64(s.player.Velocity.X))) > 10.0 && s.player.IsGrounded {
			s.runAnimTimer += dt
			if s.runAnimTimer >= 0.08 {
				s.runAnimTimer = 0.0
				s.runAnimFrame = (s.runAnimFrame + 1) % 2
			}
		} else {
			s.runAnimFrame = 0
		}
	}
}

func (s *MarioScene) triggerDeath(ctx *engine.Context) {
	s.player.Alive = false
	s.player.DeathTimer = 3.0
	s.player.Velocity = loommath.Vec2{X: 0, Y: -350.0}
	ctx.Physics.RemoveBody(s.player.Body)
	ctx.Audio.StopMusic()
	ctx.Audio.PlaySound(BumpPatch)
	s.state = StateGameOver
	s.gameOverDelay = 3.0
}

func (s *MarioScene) Render(ctx *engine.Context) {
	// Clean background with a deep, atmospheric Sunset / Night sky (Indigo/Purple)
	ctx.Draw.Clear(loommath.Color{R: 0.15, G: 0.12, B: 0.22, A: 1.0})

	// Toggle coordinate layout to world projection space
	ctx.Draw.SetUISpace(false)

	// Shift camera matrix offset horizontally and vertically for 2x zoom (640x360 viewport)
	ctx.Draw.Camera.Zoom = 2.0
	ctx.Draw.Camera.Pos = loommath.Vec2{X: s.camX + 320.0, Y: 540.0}

	// 1. Draw procedurally geometric background
	DrawBackground(ctx.Draw, s.camX)

	startCol := int(s.camX) / TileWidth
	endCol := startCol + (1280/TileWidth)/2 + 2
	if startCol < 0 {
		startCol = 0
	}
	if endCol > GridCols {
		endCol = GridCols
	}

	// 2. Draw Visible Grid Tiles
	for y := 0; y < GridRows; y++ {
		for x := startCol; x < endCol; x++ {
			tile := s.lvl.Grid[y][x]
			if tile.Type == TileAir {
				continue // Background is drawn in DrawBackground continuously
			}

			pos := loommath.Vec2{X: float32(x * TileWidth), Y: float32(y*TileHeight) + tile.VisualYOffset}

			switch tile.Type {
			case TileGround:
				DrawGroundBlock(ctx.Draw, pos)
			case TileBrick:
				DrawBrick(ctx.Draw, pos)
			case TileQuestion:
				DrawQuestionBlock(ctx.Draw, pos, false, s.coinAnimTimer)
			case TileQuestionEmpty:
				DrawQuestionBlock(ctx.Draw, pos, true, s.coinAnimTimer)
			case TileCoin:
				DrawCoin(ctx.Draw, pos, s.coinAnimTimer)
			case TileFlagpole:
				ctx.Draw.Rect(loommath.Vec2{X: float32(x*TileWidth) + 18, Y: float32(y * TileHeight)}, loommath.Vec2{X: 4, Y: TileHeight}, loommath.Color{R: 0.8, G: 0.8, B: 0.8, A: 1.0})
			case TileFlag:
				ctx.Draw.Rect(loommath.Vec2{X: float32(x*TileWidth) - 10, Y: float32(y * TileHeight)}, loommath.Vec2{X: 28, Y: 24}, loommath.Color{R: 0.1, G: 0.7, B: 0.15, A: 1.0})
			}

			// Pipe rendering
			switch tile.Type {
				case TilePipeRimLeft:
					DrawPipe(ctx.Draw, pos, float32(TileWidth)*2, float32(TileHeight))
				case TilePipeBodyLeft:
					DrawPipe(ctx.Draw, pos, float32(TileWidth)*2, float32(TileHeight))
			}
		}
	}

	// 3. Draw Goombas
	for _, g := range s.goombas {
		gPosVal := ctx.ECS.Get(g.Entity, ecs.TypePosition)
		if gPosVal == nil { continue }
		gp := gPosVal.(*ecs.Position)
		DrawGoomba(ctx.Draw, loommath.Vec2{X: gp.X, Y: gp.Y}, g.Alive)
	}

	// 4. Draw Player
	playerPosVal := ctx.ECS.Get(s.player.Entity, ecs.TypePosition)
	if playerPosVal != nil {
		p := playerPosVal.(*ecs.Position)
		DrawMario(ctx.Draw, loommath.Vec2{X: p.X, Y: p.Y}, s.player.Velocity, s.player.FacingLeft)
	}

	// 5. Toggle coordinate layout to static UI space (screen coordinates 0..1280)	// UI Rendering
	ctx.Draw.SetUISpace(true)

	if s.state == StateStartScreen {
		// Clean Stylized Title
		// Full Screen Dark Cinematic Overlay
		ctx.Draw.Rect(loommath.Vec2{X: 0, Y: 0}, loommath.Vec2{X: 1280, Y: 720}, loommath.Color{R: 0, G: 0, B: 0, A: 0.4})

		// Clean Stylized Title (Perfectly centered)
		// "LOOM MARIO" is 10 chars. 10 * 6 * size = 60 * 7 = 420. Half is 210. 640 - 210 = 430.
		DrawString(ctx.Draw, "LOOM MARIO", 430, 250, 7.0, loommath.Color{R: 0.1, G: 0.1, B: 0.2, A: 1}) // Deep drop shadow
		DrawString(ctx.Draw, "LOOM MARIO", 425, 245, 7.0, loommath.Color{R: 1.0, G: 0.9, B: 0.2, A: 1}) // Vibrant gold

		// Blinking press space to start
		pulse := float32(math.Abs(math.Sin(float64(s.coinAnimTimer) * 5.0)))
		DrawString(ctx.Draw, "PRESS SPACE TO PLAY", 469, 420, 3.0, loommath.Color{R: 1, G: 1, B: 1, A: pulse*0.8 + 0.2})
	} else {
		// Draw HUD during gameplay
		// 3. Draw Game HUD Top Bar
		DrawStringWithShadow(ctx.Draw, "MARIO", 40, 40, 3.0, loommath.Color{R: 0.95, G: 0.25, B: 0.25, A: 1.0})
		scoreStr := fmt.Sprintf("%06d", s.player.Score)
		DrawStringWithShadow(ctx.Draw, scoreStr, 40, 75, 3.0, loommath.Color{R: 1, G: 1, B: 1, A: 1})

		DrawStringWithShadow(ctx.Draw, "COINS", 400, 40, 3.0, loommath.Color{R: 0.95, G: 0.75, B: 0.1, A: 1.0})
		DrawStringWithShadow(ctx.Draw, "x", 400, 75, 3.0, loommath.Color{R: 0.95, G: 0.75, B: 0.1, A: 1.0})
		coinValStr := fmt.Sprintf("%02d", s.player.Coins)
		DrawStringWithShadow(ctx.Draw, coinValStr, 418, 75, 3.0, loommath.Color{R: 1, G: 1, B: 1, A: 1.0})

		DrawStringWithShadow(ctx.Draw, "WORLD", 800, 40, 3.0, loommath.Color{R: 0.35, G: 0.75, B: 0.95, A: 1.0})
		DrawStringWithShadow(ctx.Draw, "1-1", 800, 75, 3.0, loommath.Color{R: 1, G: 1, B: 1, A: 1.0})

		DrawStringWithShadow(ctx.Draw, "TIME", 1100, 40, 3.0, loommath.Color{R: 0.35, G: 0.85, B: 0.35, A: 1.0})
		timeStr := fmt.Sprintf("%03d", int(s.levelTimer))
		DrawStringWithShadow(ctx.Draw, timeStr, 1100, 75, 3.0, loommath.Color{R: 1, G: 1, B: 1, A: 1.0})

		if s.state == StateGameOver {
			DrawStringWithShadow(ctx.Draw, "GAME OVER", 440, 360, 5.0, loommath.Color{R: 1, G: 0.2, B: 0.2, A: 1})
		}
		if s.state == StateWin {
			DrawStringWithShadow(ctx.Draw, "LEVEL CLEARED!", 380, 360, 5.0, loommath.Color{R: 1, G: 0.9, B: 0.2, A: 1})
		}
	}
}

func (s *MarioScene) Unload(ctx *engine.Context) {
	ctx.Audio.StopMusic()
}

func main() {
	engine.Run(&MarioScene{}, engine.Config{
		Width:     1280,
		Height:    720,
		Title:     "loom Mario",
		TargetFPS: 60, // 60 FPS standard retro pacing
		Pixelated: true,
		GC: engine.GCConfig{
			GOGC:             100,
			GoMemLimit:       256,
			PauseAnnotations: false,
		},
		Icon:      loadIcon(),
	})
}

// DrawString renders textual glyphs using small visual rectangle blocks in static screen coordinates
func DrawString(rs *renderer.RenderSystem, text string, startX, startY float32, size float32, color loommath.Color) {
	for charIdx, char := range text {
		glyph, ok := Font[char]
		if !ok {
			glyph = Font[' ']
		}

		charX := startX + float32(charIdx)*6.0*size
		for row := 0; row < 7; row++ {
			line := glyph[row]
			for col := 0; col < 5; col++ {
				if line[col] == '*' {
					px := charX + float32(col)*size
					py := startY + float32(row)*size
					rs.Rect(loommath.Vec2{X: px, Y: py}, loommath.Vec2{X: size, Y: size}, color)
				}
			}
		}
	}
}

// DrawStringWithShadow draws a string with a dark drop shadow for high readability on any background
func DrawStringWithShadow(rs *renderer.RenderSystem, text string, startX, startY float32, size float32, color loommath.Color) {
	// Draw shadow (dark deep indigo/black, offset by 2 pixels down and right)
	DrawString(rs, text, startX+2, startY+2, size, loommath.Color{R: 0.02, G: 0.01, B: 0.05, A: 1.0})
	// Draw foreground text
	DrawString(rs, text, startX, startY, size, color)
}

// Pixel Font data for HUD text rendering
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
