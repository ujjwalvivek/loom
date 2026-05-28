package main

import (
	"github.com/ujjwalvivek/loom/ecs"
	"github.com/ujjwalvivek/loom/math"
	"github.com/ujjwalvivek/loom/physics"
)

const (
	TileWidth  = 40
	TileHeight = 40
	GridRows   = 18
	GridCols   = 150
)

type TileType byte

const (
	TileAir TileType = 0
	TileGround TileType = 1
	TileBrick TileType = 2
	TileQuestion TileType = 3
	TileQuestionEmpty TileType = 4
	TilePipeRimLeft TileType = 5
	TilePipeRimRight TileType = 6
	TilePipeBodyLeft TileType = 7
	TilePipeBodyRight TileType = 8
	TileCoin TileType = 9
	TileFlagpole TileType = 10
	TileFlag TileType = 11
)

type Tile struct {
	Type           TileType
	GridX, GridY   int
	Body           physics.BodyHandle
	VisualYOffset  float32
	BounceVelocity float32
	CoinsRemaining int
}

type Level struct {
	Grid               [GridRows][GridCols]*Tile
	PlayerSpawn        math.Vec2
	GoombaSpawns       []math.Vec2
	CoinEntities       map[math.Vec2]ecs.Entity // Maps grid coord to ECS entities for coins
}

var LevelLayout = []string{
	`......................................................................................................................................................`,
	`......................................................................................................................................................`,
	`......................................................................................................................................................`,
	`......................................................................................................................................................`,
	`......................................................................................................................................................`,
	`......................................................................................................................................................`,
	`......................................................................................................................................................`,
	`......................................................................................................................................................`,
	`......................................................................................................................................................`,
	`...................................................................................................................................f..................`,
	`.......................................................................................?B?B?............CCC........................F..................`,
	`..........................CCC............................CCC.......................................................................F..................`,
	`..............B?B?B......................CCC.......................................................................................F..................`,
	`.......................................................[]..........................................................................F..................`,
	`.........................................[]............()......................[]..................................................F..................`,
	`.......P............G..........G.........()...G........()..........G..G........().................G.G..............................F..................`,
	`######################################################################################################################################################`,
	`######################################################################################################################################################`,
}

func NewLevel(ctx *Context) *Level {
	lvl := &Level{
		GoombaSpawns: make([]math.Vec2, 0),
		CoinEntities: make(map[math.Vec2]ecs.Entity),
	}

	// 1. Parse level layout string matrix
	for y := 0; y < GridRows; y++ {
		rowStr := LevelLayout[y]
		for x := 0; x < GridCols; x++ {
			char := rowStr[x]
			tile := &Tile{
				GridX: x,
				GridY: y,
			}

			switch char {
			case '#':
				tile.Type = TileGround
			case 'B':
				tile.Type = TileBrick
			case '?':
				tile.Type = TileQuestion
				tile.CoinsRemaining = 1
			case '[':
				tile.Type = TilePipeRimLeft
			case ']':
				tile.Type = TilePipeRimRight
			case '(':
				tile.Type = TilePipeBodyLeft
			case ')':
				tile.Type = TilePipeBodyRight
			case 'C':
				tile.Type = TileCoin
			case 'F':
				tile.Type = TileFlagpole
			case 'f':
				tile.Type = TileFlag
			case 'G':
				// Mark goomba spawn point
				lvl.GoombaSpawns = append(lvl.GoombaSpawns, math.Vec2{X: float32(x*TileWidth + 12), Y: float32(y*TileHeight + 12)})
				tile.Type = TileAir
			case 'P':
				// Player spawn point
				lvl.PlayerSpawn = math.Vec2{X: float32(x * TileWidth), Y: float32(y * TileHeight)}
				tile.Type = TileAir
			default:
				tile.Type = TileAir
			}

			lvl.Grid[y][x] = tile
		}
	}

	// 2. Setup physics colliders - MERGE consecutive Ground blocks horizontally to reduce colliders count
	for y := 0; y < GridRows; y++ {
		inRun := false
		startX := 0
		for x := 0; x < GridCols; x++ {
			tile := lvl.Grid[y][x]
			isGround := (tile.Type == TileGround)
			if isGround {
				if !inRun {
					inRun = true
					startX = x
				}
			} else {
				if inRun {
					// Create combined ground body
					width := float32((x - startX) * TileWidth)
					posX := float32(startX * TileWidth)
					posY := float32(y * TileHeight)
					bounds := math.Rect{X: posX, Y: posY, W: width, H: float32(TileHeight)}
					bodyHandle := ctx.Physics.AddBody(0, bounds, 0x3) // LayerMask 0x3 (collides with 0x1 and 0x2)
					
					// Set body handle on all tiles in this run
					for k := startX; k < x; k++ {
						lvl.Grid[y][k].Body = bodyHandle
					}
					inRun = false
				}
			}
		}
		if inRun {
			width := float32((GridCols - startX) * TileWidth)
			posX := float32(startX * TileWidth)
			posY := float32(y * TileHeight)
			bounds := math.Rect{X: posX, Y: posY, W: width, H: float32(TileHeight)}
			bodyHandle := ctx.Physics.AddBody(0, bounds, 0x3)
			for k := startX; k < GridCols; k++ {
				lvl.Grid[y][k].Body = bodyHandle
			}
		}
	}

	// 3. Create physics bodies for other individual solid blocks (Bricks, Questions, Pipes)
	for y := 0; y < GridRows; y++ {
		for x := 0; x < GridCols; x++ {
			tile := lvl.Grid[y][x]
			if tile.Type == TileBrick || tile.Type == TileQuestion ||
				tile.Type == TilePipeRimLeft || tile.Type == TilePipeRimRight ||
				tile.Type == TilePipeBodyLeft || tile.Type == TilePipeBodyRight {
				
				posX := float32(x * TileWidth)
				posY := float32(y * TileHeight)
				bounds := math.Rect{X: posX, Y: posY, W: TileWidth, H: TileHeight}
				
				// Create individual physics collider
				bodyHandle := ctx.Physics.AddBody(0, bounds, 0x3)
				tile.Body = bodyHandle
			}
		}
	}

	return lvl
}

// Update ticks visual block animations
func (lvl *Level) Update(dt float32) {
	for y := 0; y < GridRows; y++ {
		for x := 0; x < GridCols; x++ {
			tile := lvl.Grid[y][x]
			if tile.BounceVelocity != 0 || tile.VisualYOffset != 0 {
				tile.VisualYOffset += tile.BounceVelocity * dt
				tile.BounceVelocity += 900.0 * dt // Pull back down (simulating a spring return)
				if tile.VisualYOffset >= 0 {
					tile.VisualYOffset = 0
					tile.BounceVelocity = 0
				}
			}
		}
	}
}

// BumpTile triggers the visual bounce of a block and handles question block loot
func (lvl *Level) BumpTile(ctx *Context, x, y int) bool {
	if x < 0 || x >= GridCols || y < 0 || y >= GridRows {
		return false
	}
	tile := lvl.Grid[y][x]
	if tile.Type == TileBrick {
		tile.BounceVelocity = -180.0
		ctx.Audio.PlaySound(BumpPatch)
	} else if tile.Type == TileQuestion && tile.CoinsRemaining > 0 {
		tile.BounceVelocity = -180.0
		tile.CoinsRemaining--
		if tile.CoinsRemaining == 0 {
			tile.Type = TileQuestionEmpty
		}
		ctx.Audio.PlaySound(CoinPatch)
		return true
	}
	return false
}
