package physics

import (
	"sync/atomic"
	"time"

	loommath "github.com/ujjwalvivek/loom/math"
)

type CommandType byte

const (
	CmdAddBody    CommandType = 0
	CmdRemoveBody CommandType = 1
	CmdSetVelocity CommandType = 2
	CmdSetPosition CommandType = 3
	CmdStop        CommandType = 4
)

type PhysicsCommand struct {
	Type      CommandType
	Handle    BodyHandle
	Entity    uint32
	Bounds    loommath.Rect
	Velocity  loommath.Vec2
	LayerMask uint32
}

type BodyState struct {
	Handle BodyHandle
	Entity uint32
	Pos    loommath.Vec2
	Size   loommath.Vec2
}

type PhysicsState struct {
	Bodies []BodyState
}

func (ps *PhysicsState) Reset() {
	ps.Bodies = ps.Bodies[:0]
}

type PhysicsSystem struct {
	Commands    chan PhysicsCommand
	StateChan   chan *PhysicsState
	RecycleChan chan *PhysicsState
	
	nextHandle   uint32
	currentState *PhysicsState
}

func NewPhysicsSystem() *PhysicsSystem {
	return &PhysicsSystem{
		Commands:    make(chan PhysicsCommand, 256),
		StateChan:   make(chan *PhysicsState, 2),
		RecycleChan: make(chan *PhysicsState, 2),
	}
}

func (ps *PhysicsSystem) AddBody(entity uint32, bounds loommath.Rect, layerMask uint32) BodyHandle {
	handle := BodyHandle(atomic.AddUint32(&ps.nextHandle, 1))
	ps.Commands <- PhysicsCommand{
		Type:      CmdAddBody,
		Handle:    handle,
		Entity:    entity,
		Bounds:    bounds,
		LayerMask: layerMask,
	}
	return handle
}

func (ps *PhysicsSystem) RemoveBody(handle BodyHandle) {
	ps.Commands <- PhysicsCommand{
		Type:   CmdRemoveBody,
		Handle: handle,
	}
}

func (ps *PhysicsSystem) SetVelocity(handle BodyHandle, velocity loommath.Vec2) {
	ps.Commands <- PhysicsCommand{
		Type:     CmdSetVelocity,
		Handle:   handle,
		Velocity: velocity,
	}
}

func (ps *PhysicsSystem) SetPosition(handle BodyHandle, pos loommath.Vec2) {
	ps.Commands <- PhysicsCommand{
		Type:     CmdSetPosition,
		Handle:   handle,
		Bounds:   loommath.Rect{X: pos.X, Y: pos.Y}, // uses X, Y from Bounds
	}
}

// UpdateState matches main thread's state updates from physics
func (ps *PhysicsSystem) UpdateState(newState *PhysicsState) {
	if ps.currentState != nil {
		select {
		case ps.RecycleChan <- ps.currentState:
		default:
			// drop if recycle full to prevent blocking
		}
	}
	ps.currentState = newState
}

// Query performs a thread-safe bounding box scan against the latest physics state on the main thread.
func (ps *PhysicsSystem) Query(rect loommath.Rect) []BodyState {
	if ps.currentState == nil {
		return nil
	}
	
	var results []BodyState
	for _, body := range ps.currentState.Bodies {
		bRect := loommath.Rect{X: body.Pos.X, Y: body.Pos.Y, W: body.Size.X, H: body.Size.Y}
		if rect.Intersects(bRect) {
			results = append(results, body)
		}
	}
	return results
}

func (ps *PhysicsSystem) Start(fps int, closeChan chan struct{}) {
	bodies := make(map[BodyHandle]*Body)
	grid := NewSpatialGrid(128.0)
	
	// Create a pre-allocated query buffer to avoid allocations in collision checks
	queryBuffer := make([]BodyHandle, 0, 64)

	fixedDt := 1.0 / float32(fps)
	ticker := time.NewTicker(time.Duration(int64(fixedDt * 1e9)))
	defer ticker.Stop()

	for {
		select {
		case <-closeChan:
			return
		case <-ticker.C:
			processCommands(ps.Commands, bodies)

			grid.Clear()
			for _, body := range bodies {
				grid.Insert(body)
			}

			// Continuous swept AABB and resolution with sliding
			for _, body := range bodies {
				if !body.Active || (body.Velocity.X == 0 && body.Velocity.Y == 0) {
					continue
				}

				timeRemaining := fixedDt

				for bump := 0; bump < 4; bump++ {
					if body.Velocity.X == 0 && body.Velocity.Y == 0 {
						break
					}
					if timeRemaining <= 0 {
						break
					}

					// Broadphase: Query surrounding body candidates
					sweepRange := loommath.Rect{
						X: body.Bounds.X + minF(0, body.Velocity.X*timeRemaining),
						Y: body.Bounds.Y + minF(0, body.Velocity.Y*timeRemaining),
						W: body.Bounds.W + absF(body.Velocity.X*timeRemaining),
						H: body.Bounds.H + absF(body.Velocity.Y*timeRemaining),
					}
					queryBuffer = grid.Query(sweepRange, queryBuffer)

					// Narrowphase: Swept AABB checks
					var earliestCollision Collision
					earliestCollision.Time = 1.0

					for _, otherHandle := range queryBuffer {
						if otherHandle == body.Handle {
							continue
						}
						other := bodies[otherHandle]
						if other == nil || !other.Active {
							continue
						}

						// Collision Layer mask check (disjoint layers do not collide)
						if (body.LayerMask & other.LayerMask) == 0 {
							continue
						}

						toi, normal := Sweep(body.Bounds, other.Bounds, body.Velocity.MulScalar(timeRemaining))
						if toi < earliestCollision.Time {
							earliestCollision.Hit = true
							earliestCollision.Time = toi
							earliestCollision.Normal = normal
						}
					}

					// Resolve movement
					if earliestCollision.Hit {
						body.Bounds.X += body.Velocity.X * timeRemaining * earliestCollision.Time
						body.Bounds.Y += body.Velocity.Y * timeRemaining * earliestCollision.Time

						dotprod := body.Velocity.X*earliestCollision.Normal.X + body.Velocity.Y*earliestCollision.Normal.Y
						body.Velocity.X -= dotprod * earliestCollision.Normal.X
						body.Velocity.Y -= dotprod * earliestCollision.Normal.Y

						timeRemaining *= (1.0 - earliestCollision.Time)
					} else {
						body.Bounds.X += body.Velocity.X * timeRemaining
						body.Bounds.Y += body.Velocity.Y * timeRemaining
						break
					}
				}

				// Depenetration pass to resolve overlaps and prevent sticking to walls/ceilings
				queryBuffer = grid.Query(body.Bounds, queryBuffer)
				for _, otherHandle := range queryBuffer {
					if otherHandle == body.Handle {
						continue
					}
					other := bodies[otherHandle]
					if other == nil || !other.Active {
						continue
					}
					if (body.LayerMask & other.LayerMask) == 0 {
						continue
					}

					if overlapping, normal, depth := Overlap(body.Bounds, other.Bounds); overlapping {
						body.Bounds.X += normal.X * (depth + 0.01)
						body.Bounds.Y += normal.Y * (depth + 0.01)

						dot := body.Velocity.X*normal.X + body.Velocity.Y*normal.Y
						if dot < 0 {
							body.Velocity.X -= dot * normal.X
							body.Velocity.Y -= dot * normal.Y
						}
					}
				}
			}

			// Capture current state and push to main thread
			var state *PhysicsState
			select {
			case state = <-ps.RecycleChan:
				state.Reset()
			default:
				state = &PhysicsState{
					Bodies: make([]BodyState, 0, len(bodies)),
				}
			}

			for _, b := range bodies {
				if b.Active {
					state.Bodies = append(state.Bodies, BodyState{
						Handle: b.Handle,
						Entity: b.Entity,
						Pos:    loommath.Vec2{X: b.Bounds.X, Y: b.Bounds.Y},
						Size:   loommath.Vec2{X: b.Bounds.W, Y: b.Bounds.H},
					})
				}
			}

			select {
			case ps.StateChan <- state:
			default:
				// If main thread is falling behind, drop or overwrite to prevent locking physics
			}
		}
	}
}

func processCommands(commands chan PhysicsCommand, bodies map[BodyHandle]*Body) {
	for {
		select {
		case cmd := <-commands:
			switch cmd.Type {
			case CmdAddBody:
				bodies[cmd.Handle] = &Body{
					Handle:    cmd.Handle,
					Entity:    cmd.Entity,
					Bounds:    cmd.Bounds,
					Velocity:  cmd.Velocity,
					LayerMask: cmd.LayerMask,
					Active:    true,
				}
			case CmdRemoveBody:
				if b, ok := bodies[cmd.Handle]; ok {
					b.Active = false
					delete(bodies, cmd.Handle)
				}
			case CmdSetVelocity:
				if b, ok := bodies[cmd.Handle]; ok {
					b.Velocity = cmd.Velocity
				}
			case CmdSetPosition:
				if b, ok := bodies[cmd.Handle]; ok {
					b.Bounds.X = cmd.Bounds.X
					b.Bounds.Y = cmd.Bounds.Y
				}
			case CmdStop:
				return
			}
		default:
			return
		}
	}
}

// Math helpers
func minF(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func maxF(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

func absF(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}
