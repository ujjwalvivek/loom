package engine

import (
	"github.com/ujjwalvivek/loom/audio"
	"github.com/ujjwalvivek/loom/ecs"
	"github.com/ujjwalvivek/loom/input"
	"github.com/ujjwalvivek/loom/perf"
	"github.com/ujjwalvivek/loom/physics"
	"github.com/ujjwalvivek/loom/renderer"
)

type Context struct {
	Draw    *renderer.RenderSystem
	Physics *physics.PhysicsSystem
	Audio   *audio.AudioSystem
	Input   *input.InputState
	ECS     *ecs.World
	Perf    *perf.PerfSystem
	Scene   *SceneManager
}

func NewContext(width, height int32) *Context {
	ctx := &Context{
		Draw:    renderer.NewRenderSystem(width, height),
		Physics: physics.NewPhysicsSystem(),
		Audio:   audio.NewAudioSystem(),
		ECS:     ecs.NewWorld(),
		Perf:    perf.NewPerfSystem(),
	}
	ctx.Scene = NewSceneManager(ctx)
	return ctx
}
