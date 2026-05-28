package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/ujjwalvivek/loom/audio"
	"github.com/ujjwalvivek/loom/ecs"
	"github.com/ujjwalvivek/loom/engine"
	loommath "github.com/ujjwalvivek/loom/math"
	"github.com/ujjwalvivek/loom/physics"
	"github.com/ujjwalvivek/loom/termrenderer"
)

type GameState int

const (
	StateStartScreen GameState = 0
	StatePlaying     GameState = 1
	StateGameOver    GameState = 2
	StateWin         GameState = 3
)

// We redefine Context locally to use TermRenderer and TermInput
type Context struct {
	Draw    *termrenderer.TermRenderer
	Input   *termrenderer.TermInput
	Physics *physics.PhysicsSystem
	Audio   *audio.AudioSystem
	ECS     *ecs.World
}

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

func (s *MarioScene) Load(ctx *Context) {
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
	s.state = StateStartScreen

	SetupMarioAudio(ctx)
}

func (s *MarioScene) Update(ctx *Context, dt float32) {
	switch s.state {
	case StateStartScreen:
		s.camX += dt * 80.0
		if s.camX > float32(GridCols*TileWidth-640) {
			s.camX = 0
		}

		s.coinAnimTimer += dt
		if s.coinAnimTimer >= 0.4 {
			s.coinAnimTimer = 0.0
		}

		if ctx.Input.KeyPressed(glfw.KeySpace) || ctx.Input.KeyPressed(glfw.KeyEnter) {
			s.state = StatePlaying
			s.camX = 0
		}
		return

	case StateGameOver:
		s.gameOverDelay -= dt
		if s.gameOverDelay <= 0 {
			s.Unload(ctx)
			s.Load(ctx)
			return
		}
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
		s.player.Velocity = loommath.Vec2{}
		ctx.Physics.SetVelocity(s.player.Body, s.player.Velocity)
		return

	case StatePlaying:
		s.levelTimer -= dt * 1.5
		if s.levelTimer <= 0 {
			s.levelTimer = 0
			s.triggerDeath(ctx)
			return
		}

		leftPressed := ctx.Input.KeyDown(glfw.KeyA) || ctx.Input.KeyDown(glfw.KeyLeft)
		rightPressed := ctx.Input.KeyDown(glfw.KeyD) || ctx.Input.KeyDown(glfw.KeyRight)
		jumpPressed := ctx.Input.KeyDown(glfw.KeySpace) || ctx.Input.KeyDown(glfw.KeyW) || ctx.Input.KeyDown(glfw.KeyUp)
		shiftPressed := ctx.Input.KeyDown(glfw.KeyLeftShift) || ctx.Input.KeyDown(glfw.KeyRightShift)

		accel := float32(2200.0)
		friction := float32(2800.0)
		maxSpeed := float32(220.0)

		if shiftPressed {
			maxSpeed = 340.0
			accel = 3200.0
		}

		if rightPressed {
			if s.player.Velocity.X < 0 {
				s.player.Velocity.X += accel * 2.5 * dt
			} else {
				s.player.Velocity.X += accel * dt
			}
			if s.player.Velocity.X > maxSpeed {
				s.player.Velocity.X = maxSpeed
			}
			s.player.FacingLeft = false
		} else if leftPressed {
			if s.player.Velocity.X > 0 {
				s.player.Velocity.X -= accel * 2.5 * dt
			} else {
				s.player.Velocity.X -= accel * dt
			}
			if s.player.Velocity.X < -maxSpeed {
				s.player.Velocity.X = -maxSpeed
			}
			s.player.FacingLeft = true
		} else {
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

		gravity := float32(2800.0)
		if s.player.Velocity.Y > 0 {
			gravity = 3800.0
		}
		s.player.Velocity.Y += gravity * dt
		if s.player.Velocity.Y > 800.0 {
			s.player.Velocity.Y = 800.0
		}

		if jumpPressed && s.player.IsGrounded {
			s.player.Velocity.Y = -760.0
			s.player.Jumping = true
			s.player.JumpHold = 0.0
			ctx.Audio.PlaySound(JumpPatch)
		} else if jumpPressed && s.player.Jumping {
			s.player.JumpHold += dt
			if s.player.JumpHold < 0.22 {
				s.player.Velocity.Y -= 900.0 * dt
			} else {
				s.player.Jumping = false
			}
		} else {
			s.player.Jumping = false
		}

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

		s.goombas = UpdateEntities(ctx, s.lvl, s.player, s.goombas, dt)
		s.lvl.Update(dt)

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

			if p.Y > 730.0 {
				s.triggerDeath(ctx)
				return
			}
		}

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

func (s *MarioScene) triggerDeath(ctx *Context) {
	s.player.Alive = false
	s.player.DeathTimer = 3.0
	s.player.Velocity = loommath.Vec2{X: 0, Y: -350.0}
	ctx.Physics.RemoveBody(s.player.Body)
	ctx.Audio.StopMusic()
	ctx.Audio.PlaySound(BumpPatch)
	s.state = StateGameOver
	s.gameOverDelay = 3.0
}

func (s *MarioScene) Render(ctx *Context) {
	// Terminal is roughly 120x36 columns. We need to downscale standard 1280x720 coordinates!
	// We'll scale coordinates by dividing them by 10.
	scale := float32(10.0)

	// Clear with a dark slate clear color
	ctx.Draw.Clear(loommath.Color{R: 0.05, G: 0.05, B: 0.1, A: 1.0})

	// Scale camera
	termCamX := s.camX / scale

	// Draw sky background
	DrawBackground(ctx.Draw, termCamX, scale)

	startCol := int(s.camX) / TileWidth
	endCol := startCol + (1280/TileWidth)/2 + 2
	if startCol < 0 {
		startCol = 0
	}
	if endCol > GridCols {
		endCol = GridCols
	}

	for y := 0; y < GridRows; y++ {
		for x := startCol; x < endCol; x++ {
			tile := s.lvl.Grid[y][x]
			if tile.Type == TileAir {
				continue
			}

			// Screen Space position
			wx := float32(x * TileWidth) - s.camX
			wy := float32(y * TileHeight) + tile.VisualYOffset

			// Map to 120x72 pixel grid with aspect ratio correction:
			// Subtract 360 camera Y offset, scale by 5.33, and add vertical offset of 4.5
			pos := loommath.Vec2{
				X: wx / 5.33,
				Y: (wy - 360.0)/5.33 + 4.5,
			}

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
				DrawFlagpole(ctx.Draw, pos)
			case TileFlag:
				// Draw flagpole behind the flag first, then the flag locked to integer pos
				DrawFlagpole(ctx.Draw, pos)
				DrawFlag(ctx.Draw, loommath.Vec2{X: pos.X - 2, Y: pos.Y + 2})
			case TilePipeRimLeft, TilePipeBodyLeft:
				DrawPipe(ctx.Draw, pos, float32(TileWidth)*2/5.33, float32(TileHeight)/5.33)
			}
		}
	}

	for _, g := range s.goombas {
		gPosVal := ctx.ECS.Get(g.Entity, ecs.TypePosition)
		if gPosVal == nil { continue }
		gp := gPosVal.(*ecs.Position)
		wx := gp.X - s.camX
		pos := loommath.Vec2{
			X: wx / 5.33,
			Y: (gp.Y - 360.0)/5.33 + 4.5,
		}
		// Goomba walk cycle based on world X
		walkCycle := int(gp.X / 12.0) % 2 == 0
		DrawGoomba(ctx.Draw, pos, g.Alive, walkCycle)
	}

	playerPosVal := ctx.ECS.Get(s.player.Entity, ecs.TypePosition)
	if playerPosVal != nil {
		p := playerPosVal.(*ecs.Position)
		wx := p.X - s.camX
		pos := loommath.Vec2{
			X: wx / 5.33,
			Y: (p.Y - 360.0)/5.33 + 4.5,
		}
		// Mario walk cycle based on world X position (prevents camera lock freeze)
		walkCycle := int(p.X / 16.0) % 2 == 0
		DrawMario(ctx.Draw, pos, s.player.Velocity, s.player.FacingLeft, walkCycle)
	}

	// Draw HUD using 3x5 small pixel font on a single row (Y = 2) to look extremely compact (5px high)
	hudRed := loommath.Color{R: 0.95, G: 0.2, B: 0.2, A: 1.0}
	hudGold := loommath.Color{R: 0.95, G: 0.75, B: 0.1, A: 1.0}

	// Score: M 000000
	DrawSmallPixelText(ctx.Draw, "M", 2, 2, hudRed)
	DrawSmallPixelText(ctx.Draw, fmt.Sprintf("%06d", s.player.Score), 7, 2, loommath.ColorWhite)

	// Coins: C x00
	DrawSmallPixelText(ctx.Draw, "C", 44, 2, hudGold)
	DrawSmallPixelText(ctx.Draw, fmt.Sprintf("x%02d", s.player.Coins), 49, 2, loommath.ColorWhite)

	// World: 1-1
	DrawSmallPixelText(ctx.Draw, "1-1", 74, 2, loommath.ColorWhite)

	// Time: 000
	DrawSmallPixelText(ctx.Draw, fmt.Sprintf("%03d", int(s.levelTimer)), 98, 2, loommath.ColorWhite)

	// Draw start screen overlay if in Start state (no credits, clean logo, pulsing subtitle at bottom)
	if s.state == StateStartScreen {
		// Large 2x logo titles with drop shadows (shifted down by 1 character row)
		DrawPixelTextScale(ctx.Draw, "LOOM", 37, 13, 2, loommath.ColorBlack)
		DrawPixelTextScale(ctx.Draw, "LOOM", 36, 12, 2, loommath.Color{R: 0.98, G: 0.78, B: 0.05, A: 1.0})

		DrawPixelTextScale(ctx.Draw, "MARIO", 31, 29, 2, loommath.ColorBlack)
		DrawPixelTextScale(ctx.Draw, "MARIO", 30, 28, 2, loommath.Color{R: 0.92, G: 0.16, B: 0.08, A: 1.0})

		// Pulsing "PRESS SPACE TO PLAY" with drop shadow (shifted up by 1 character row)
		pulse := float32(math.Abs(math.Sin(float64(s.coinAnimTimer) * 5.0)))
		textColor := loommath.Color{R: pulse, G: pulse, B: pulse, A: 1.0}
		DrawPixelText(ctx.Draw, "PRESS SPACE TO PLAY", 4, 63, loommath.ColorBlack)
		DrawPixelText(ctx.Draw, "PRESS SPACE TO PLAY", 3, 62, textColor)
	}

	ctx.Draw.Present()
}

func (s *MarioScene) Unload(ctx *Context) {
	ctx.Audio.StopMusic()
}

func main() {
	// Set terminal window title (cross-platform ANSI escape sequence)
	fmt.Print("\033]0;Loom Mario Terminal\007")

	// Initialize custom terminal engine context
	// Terminal Width: 120 columns, Height: 36 rows (yielding a 120x72 pixel grid)
	ctx := &Context{
		Draw:    termrenderer.NewTermRenderer(120, 36),
		Input:   termrenderer.NewTermInput(),
		Physics: physics.NewPhysicsSystem(),
		Audio:   audio.NewAudioSystem(),
		ECS:     ecs.NewWorld(),
	}

	defer ctx.Draw.Shutdown()
	defer ctx.Input.Shutdown()

	physicsClose := make(chan struct{})
	go ctx.Physics.Start(180, physicsClose)
	defer close(physicsClose)

	scene := &MarioScene{}
	scene.Load(ctx)

	// Intercept Ctrl+C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	loop := engine.NewLoop(60)
	loop.Reset()

	running := true
	for running {
		frameStart := time.Now()

		select {
		case <-c:
			running = false
		default:
		}

		if ctx.Input.KeyDown(glfw.KeyEscape) {
			running = false
		}

		select {
		case newState := <-ctx.Physics.StateChan:
			ctx.Physics.UpdateState(newState)
		default:
		}

		_, dt, _ := loop.Tick()
		scene.Update(ctx, dt)
		scene.Render(ctx)

		ctx.Input.Update()

		// Limit to 60 FPS to prevent CPU saturation and terminal output buffering
		elapsed := time.Since(frameStart)
		if elapsed < 16*time.Millisecond {
			time.Sleep(16*time.Millisecond - elapsed)
		}
	}
}
