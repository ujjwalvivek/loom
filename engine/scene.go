package engine

type Scene interface {
	Load(ctx *Context)
	Update(ctx *Context, dt float32)
	Render(ctx *Context)
	Unload(ctx *Context)
}

type SceneManager struct {
	stack []Scene
	ctx   *Context
}

func NewSceneManager(ctx *Context) *SceneManager {
	return &SceneManager{
		stack: make([]Scene, 0),
		ctx:   ctx,
	}
}

func (sm *SceneManager) Push(scene Scene) {
	sm.stack = append(sm.stack, scene)
	scene.Load(sm.ctx)
}

func (sm *SceneManager) Pop() {
	if len(sm.stack) == 0 {
		return
	}
	idx := len(sm.stack) - 1
	sm.stack[idx].Unload(sm.ctx)
	sm.stack = sm.stack[:idx]
}

func (sm *SceneManager) Replace(scene Scene) {
	sm.Pop()
	sm.Push(scene)
}

func (sm *SceneManager) Current() Scene {
	if len(sm.stack) == 0 {
		return nil
	}
	return sm.stack[len(sm.stack)-1]
}
