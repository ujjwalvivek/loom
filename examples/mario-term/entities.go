package main

import (
	"github.com/ujjwalvivek/loom/ecs"
	"github.com/ujjwalvivek/loom/math"
	"github.com/ujjwalvivek/loom/physics"
)

type Player struct {
	Entity      ecs.Entity
	Body        physics.BodyHandle
	Coins       int
	Score       int
	IsGrounded  bool
	Jumping     bool
	JumpHold    float32
	FacingLeft  bool
	Alive       bool
	Win         bool
	DeathTimer  float32
	Velocity    math.Vec2
}

type Goomba struct {
	Entity       ecs.Entity
	Body         physics.BodyHandle
	FacingLeft   bool
	Alive        bool
	SquishedTime float32
	Velocity     math.Vec2
}

func SpawnPlayer(ctx *Context, pos math.Vec2) *Player {
	e := ctx.ECS.NewEntity()
	
	// Set initial values
	p := &Player{
		Entity:     e,
		Coins:      0,
		Score:      0,
		IsGrounded: false,
		Jumping:    false,
		FacingLeft: false,
		Alive:      true,
		Win:        false,
		Velocity:   math.Vec2{X: 0, Y: 0},
	}

	// Mario height is 38, width is 30 (representing a 32x40 bounding box minus margins)
	bounds := math.Rect{X: pos.X, Y: pos.Y, W: 32, H: 38}
	p.Body = ctx.Physics.AddBody(uint32(e), bounds, 0x1)

	// Add components
	ctx.ECS.Add(e, ecs.TypePosition, ecs.Position{X: pos.X, Y: pos.Y})
	ctx.ECS.Add(e, ecs.TypeVelocity, ecs.Velocity{X: 0, Y: 0})
	ctx.ECS.Add(e, ecs.TypePhysics, ecs.PhysicsBody{Handle: p.Body})

	return p
}

func SpawnGoomba(ctx *Context, pos math.Vec2) *Goomba {
	e := ctx.ECS.NewEntity()
	
	g := &Goomba{
		Entity:     e,
		FacingLeft: true,
		Alive:      true,
		Velocity:   math.Vec2{X: -60.0, Y: 0},
	}

	// Goomba bounding box: 32x30
	bounds := math.Rect{X: pos.X, Y: pos.Y, W: 32, H: 30}
	g.Body = ctx.Physics.AddBody(uint32(e), bounds, 0x2) // LayerMask 0x2 (does not collide physically with player 0x1)

	ctx.ECS.Add(e, ecs.TypePosition, ecs.Position{X: pos.X, Y: pos.Y})
	ctx.ECS.Add(e, ecs.TypeVelocity, ecs.Velocity{X: -60.0, Y: 0})
	ctx.ECS.Add(e, ecs.TypePhysics, ecs.PhysicsBody{Handle: g.Body})

	return g
}

// UpdateEntities ticks physics, custom logic, and query collisions for player and enemies
func UpdateEntities(ctx *Context, lvl *Level, player *Player, goombas []*Goomba, dt float32) []*Goomba {
	// 1. Get player position from ECS (copied from physics thread outputs)
	posVal := ctx.ECS.Get(player.Entity, ecs.TypePosition)
	var playerPos math.Vec2
	if posVal != nil {
		p := posVal.(*ecs.Position)
		playerPos.X = p.X
		playerPos.Y = p.Y
	}

	// 2. Process Player Logic
	if player.Alive && !player.Win {
		// Query under feet to determine if player is standing on a solid body
		feetRect := math.Rect{X: playerPos.X + 4, Y: playerPos.Y + 38, W: 24, H: 4}
		bodiesBelow := ctx.Physics.Query(feetRect)
		player.IsGrounded = false
		for _, b := range bodiesBelow {
			if b.Handle != player.Body && b.Entity == 0 {
				player.IsGrounded = true
				break
			}
		}

		// Adjust vertical physics on landing
		if player.IsGrounded {
			if player.Velocity.Y > 0 {
				player.Velocity.Y = 0
			}
			player.Jumping = false
			player.JumpHold = 0
		}

		// Head Bump Brick Collision - if moving upwards
		if player.Velocity.Y < 0 {
			headRect := math.Rect{X: playerPos.X + 4, Y: playerPos.Y - 4, W: 24, H: 4}
			bodiesAbove := ctx.Physics.Query(headRect)
			for _, b := range bodiesAbove {
				if b.Handle != player.Body && b.Entity == 0 {
					// Identify grid coordinate
					gridX := int((b.Pos.X + b.Size.X/2.0) / TileWidth)
					gridY := int((b.Pos.Y + b.Size.Y/2.0) / TileHeight)
					
					if lvl.BumpTile(ctx, gridX, gridY) {
						player.Coins++
						player.Score += 200
					}
					player.Velocity.Y = 60.0 // bounce head back down slightly
					player.Jumping = false
					break
				}
			}
		}

		// Win condition checks: hit the Flagpole F
		flagX := float32(139 * TileWidth)
		if playerPos.X >= flagX-10 && playerPos.X <= flagX+40 {
			player.Win = true
			player.Velocity = math.Vec2{}
			ctx.Physics.SetVelocity(player.Body, player.Velocity)
			ctx.Audio.StopMusic()
			ctx.Audio.PlaySound(WinPatch)
		}

		// Send player velocity updates to physics engine
		ctx.Physics.SetVelocity(player.Body, player.Velocity)
	} else if !player.Alive {
		// Death animation: falling down screen space
		player.Velocity.Y += 980.0 * dt
		
		posVal := ctx.ECS.Get(player.Entity, ecs.TypePosition)
		if posVal != nil {
			p := posVal.(*ecs.Position)
			p.Y += player.Velocity.Y * dt
		}
	}

	// 3. Process Goombas Logic
	activeGoombas := make([]*Goomba, 0, len(goombas))

	for _, g := range goombas {
		gPosVal := ctx.ECS.Get(g.Entity, ecs.TypePosition)
		if gPosVal == nil {
			continue
		}
		gPos := gPosVal.(*ecs.Position)

		if !g.Alive {
			g.SquishedTime -= dt
			if g.SquishedTime > 0 {
				activeGoombas = append(activeGoombas, g)
			} else {
				// Delete goomba from ECS registry
				ctx.ECS.Destroy(g.Entity)
			}
			continue
		}

		// Apply Gravity to Goomba
		g.Velocity.Y += 980.0 * dt
		if g.Velocity.Y > 400.0 {
			g.Velocity.Y = 400.0
		}

		// Query under feet to determine if Goomba is standing on a solid body
		gFeetRect := math.Rect{X: gPos.X + 4, Y: gPos.Y + 30, W: 24, H: 4}
		gBodiesBelow := ctx.Physics.Query(gFeetRect)
		gGrounded := false
		for _, b := range gBodiesBelow {
			if b.Handle != g.Body && b.Entity != uint32(player.Entity) {
				gGrounded = true
				break
			}
		}
		if gGrounded {
			if g.Velocity.Y > 0 {
				g.Velocity.Y = 0
			}
		}

		// Move horizontally
		if g.FacingLeft {
			g.Velocity.X = -60.0
		} else {
			g.Velocity.X = 60.0
		}

		// Query walls in front of Goomba to rebound directions
		// Use H=12 and YOffset=2 to ensure the detection box is centered vertically.
		// Use a tight 1px overlap on the edge (W=2, X=gPos.X-1 for left, and X=gPos.X+31 for right)
		// so Goombas visually touch the pipe/enemy before flipping, preventing mid-air flips.
		// Only check side collisions when grounded to avoid false mid-air wall flips.
		if gGrounded {
			var sideRect math.Rect
			if g.FacingLeft {
				sideRect = math.Rect{X: gPos.X - 1.0, Y: gPos.Y + 2, W: 2, H: 12}
			} else {
				sideRect = math.Rect{X: gPos.X + 31.0, Y: gPos.Y + 2, W: 2, H: 12}
			}
			bodiesSide := ctx.Physics.Query(sideRect)
			for _, b := range bodiesSide {
				if b.Handle != g.Body && b.Entity != uint32(player.Entity) {
					g.FacingLeft = !g.FacingLeft
					// Update velocity immediately so it doesn't push into the wall for a frame
					if g.FacingLeft {
						g.Velocity.X = -60.0
					} else {
						g.Velocity.X = 60.0
					}
					break
				}
			}
		}

		ctx.Physics.SetVelocity(g.Body, g.Velocity)
		activeGoombas = append(activeGoombas, g)

		// Sync local position structure for collision evaluation
		gPosition := math.Vec2{X: gPos.X, Y: gPos.Y}

		// 4. Evaluate Collision between Player and Goomba
		if player.Alive && !player.Win {
			marioRect := math.Rect{X: playerPos.X + 4, Y: playerPos.Y + 4, W: 24, H: 30}
			goombaRect := math.Rect{X: gPosition.X + 2, Y: gPosition.Y + 2, W: 28, H: 26}

			if marioRect.Intersects(goombaRect) {
				// Stomp detection: check if Mario is falling onto the Goomba
				if player.Velocity.Y >= 0 && playerPos.Y+32 <= gPosition.Y+10 {
					g.Alive = false
					g.SquishedTime = 0.6 // render squished sprite for 0.6s
					ctx.Physics.RemoveBody(g.Body) // stop physical collisions
					
					// Bounce player upwards
					player.Velocity.Y = -280.0
					player.Jumping = true
					player.Score += 100
					ctx.Audio.PlaySound(StompPatch)
				} else {
					// Side collision: Kill Player
					player.Alive = false
					player.DeathTimer = 3.0
					player.Velocity.X = 0
					player.Velocity.Y = -350.0 // pop up on death
					ctx.Physics.RemoveBody(player.Body)
					ctx.Audio.StopMusic()
					ctx.Audio.PlaySound(BumpPatch)
				}
			}
		}
	}

	return activeGoombas
}
