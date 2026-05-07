package engine

import (
	"log"
	"runtime"
	"runtime/debug"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/ujjwalvivek/loom/input"
	"github.com/ujjwalvivek/loom/perf"
)

func Run(initialScene Scene, config Config) {
	// Lock main thread for GLFW and OpenGL context (mandatory)
	runtime.LockOSThread()

	// Configure initial dynamic GC parameters
	if config.GC.GOGC > 0 {
		debug.SetGCPercent(config.GC.GOGC)
	}
	if config.GC.GoMemLimit > 0 {
		debug.SetMemoryLimit(config.GC.GoMemLimit * 1024 * 1024) // MB to Bytes
	}

	// Initialize GLFW
	if err := glfw.Init(); err != nil {
		log.Fatalln("Failed to initialize GLFW:", err)
	}
	defer glfw.Terminate()

	// Window Hints (OpenGL 3.3 Core Profile)
	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	// Create GLFW Window
	window, err := glfw.CreateWindow(config.Width, config.Height, config.Title, nil, nil)
	if err != nil {
		log.Fatalln("Failed to create GLFW window:", err)
	}
	defer window.Destroy()
	
	window.MakeContextCurrent()

	// Initialize OpenGL bindings
	if err := gl.Init(); err != nil {
		log.Fatalln("Failed to initialize OpenGL bindings:", err)
	}

	// Enable alpha blending for transparent sprites
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	// Disable VSync to control framerate pacing in engine.Loop
	glfw.SwapInterval(0)

	// Initialize input system
	inputSys := input.NewInputSystem()
	inputSys.RegisterCallbacks(window)

	// Initialize context
	ctx := NewContext(int32(config.Width), int32(config.Height))

	// Start physics thread goroutine (180 Hz target rate for low-latency platforming)
	physicsClose := make(chan struct{})
	go ctx.Physics.Start(180, physicsClose)
	defer close(physicsClose)

	// Load initial scene
	ctx.Scene.Push(initialScene)

	// Timing loop
	loop := NewLoop(config.TargetFPS)
	loop.Reset()

	showOverlay := config.GC.PauseAnnotations
	
	// Main execution loop
	for !window.ShouldClose() {
		// Poll window events (must occur on the main thread)
		glfw.PollEvents()

		// Read latest physics state snapshot if available (non-blocking)
		select {
		case newState := <-ctx.Physics.StateChan:
			ctx.Physics.UpdateState(newState)
		default:
		}

		// Update actions and snapshot input state for the current frame
		inputSys.UpdateActions()
		ctx.Input = inputSys.Snapshot()

		// F3 key toggles performance stats overlay panel
		if ctx.Input.KeyPressed(glfw.KeyF3) {
			showOverlay = !showOverlay
		}

		stats := ctx.Perf.GetStats()

		// Dynamic GC parameter tuning via arrow keys
		if ctx.Input.KeyPressed(glfw.KeyUp) {
			newGOGC := stats.GOGCPercent + 10
			if newGOGC > 300 {
				newGOGC = 300
			}
			ctx.Perf.SetGCTuning(newGOGC, int64(stats.GOMemLimitMB)*1024*1024)
		}
		if ctx.Input.KeyPressed(glfw.KeyDown) {
			newGOGC := stats.GOGCPercent - 10
			if newGOGC < 10 {
				newGOGC = 10
			}
			ctx.Perf.SetGCTuning(newGOGC, int64(stats.GOMemLimitMB)*1024*1024)
		}
		if ctx.Input.KeyPressed(glfw.KeyRight) {
			newLimit := stats.GOMemLimitMB + 16
			if newLimit > 512 {
				newLimit = 512
			}
			ctx.Perf.SetGCTuning(stats.GOGCPercent, int64(newLimit)*1024*1024)
		}
		if ctx.Input.KeyPressed(glfw.KeyLeft) {
			newLimit := stats.GOMemLimitMB - 16
			if newLimit < 16 {
				newLimit = 16
			}
			ctx.Perf.SetGCTuning(stats.GOGCPercent, int64(newLimit)*1024*1024)
		}

		// Advance timer
		_, dt, _ := loop.Tick()
		ctx.Perf.AddFrame(dt)

		// Update active scene
		currentScene := ctx.Scene.Current()
		if currentScene == nil {
			break
		}
		
		// Update rendering system (camera tick, etc.)
		ctx.Draw.Update(dt)

		currentScene.Update(ctx, dt)

		// Render scene wrapped in Post-FX pass
		ctx.Draw.Begin()
		currentScene.Render(ctx)

		if showOverlay {
			perf.RenderOverlay(ctx.Draw, stats, ctx.Perf.FrameTimes, float32(config.Width), float32(config.Height))
		}

		ctx.Draw.End(float32(glfw.GetTime()))

		// Present
		window.SwapBuffers()

		// Transition transient input states
		inputSys.PostUpdate()
	}

	// Unload all scenes on shutdown
	for ctx.Scene.Current() != nil {
		ctx.Scene.Pop()
	}
}
