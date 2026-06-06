# Loom Engine

Loom is a lightweight, concurrent, custom 2D game engine in Go. Built to explore concurrent game system architectures, custom DSP audio synthesis, and real-time garbage collection tuning in a performance-critical game loop. Third in a series alongside Journey (Rust/WASM) and TinyTS (TypeScript).

Loom is built directly on top of **OpenGL 3.3** for graphics, **GLFW 3.3** for window management and inputs, and **Oto v3** for low-level audio streaming. It's concurrent. It's fast. It's bloat-free. It's minimal by design, you get the core systems and add what you need on top.

> Since the engine uses OpenGL and GLFW, your system needs to support graphics rendering. On Windows, GLFW DLLs are loaded dynamically, so you don't need to configure compile toolchains. Make sure you have Go installed (version 1.25 or newer is recommended).

## Philosophy

Loom is built around **systems talking to each other**.

Every major system runs in its own goroutine. Inter-system data crosses boundaries purely via channel messages and double-buffered state pools. The game loop orchestrates them. No shared mutable state between the renderer, physics, and audio goroutines. GC is tunable, short-pause, and cooperative. `GOGC` and `GOMEMLIMIT` are first-class knobs exposed in engine config. The perf system annotates GC pause events on the frame time chart. You can watch the GC respond to tuning in real time.

## Engine Architecture & Features

### Concurrency Model

Loom decouples rendering from simulation and audio by assigning discrete subsystems to long-lived goroutines. Data crosses boundaries purely via channel messages and double-buffered state pools.

```text
Main OS Thread
    ├── GLFW Polling & Window Context
    ├── Render Pipeline (Strict OpenGL 3.3 Core Profile)
    └── Orchestrator
         ├── ECS Parallel System Execution
         └── Telemetry & Debug Overlays

Physics Thread (goroutine)
    ← Listens on Input/Command channel
    → Pushes PhysicsState snapshots via a double-buffered recycle pool 

Audio Thread (goroutine)
    ← Listens on discrete Event channel (play/stop/sequence)
    → Streams Float32 PCM directly to the hardware buffer
```

### State Synchronization

To prevent mutex contention on the hot path, the Physics thread pushes deep-copied state snapshots into a lock-free recycle pool. The Renderer consumes the most recent snapshot for drawing, discarding outdated frames if the simulation outpaces the display.

### SubSystems

The engine is broken down into small, focused modules under the root directory:

* **`renderer/`**: An OpenGL 3.3 batch renderer. It batches sprites, textured shapes, and solids into a single draw call where possible. It also includes an orthographic camera that supports target tracking, screen shake, and smooth zooming.
* **`physics/`**: An AABB swept-collision physics system. It uses a **multi-pass sliding solver** (up to 4 passes per tick). When a player jumps or runs into a wall, it slides their velocity vector smoothly along the surface rather than stopping them dead in their tracks, making it ideal for snappy platformer movement.
* **`audio/`**: A retro-themed procedural audio system. It features a chiptune synthesizer supporting square, triangle, sawtooth, sine, and noise generators. Sound effects utilize custom ADSR envelopes, and background music is driven by a programmable step sequencer.
* **`ecs/`**: A lightweight Entity Component System (ECS) to manage positions, velocities, and physics components.
* **`input/`**: Keyboard and mouse wrappers on top of GLFW to poll current key states or listen for discrete keypresses.
* **`math/`**: Standard 2D vector and rectangle math structures, plus a linear feedback shift register (LFSR) for retro noise generation.
* **`perf/`**: GUI overlays for performance related metrics.
* **`termrenderer/`**: A terminal renderer that draws sprites and shapes by rasterizing triangles and textured quads directly into a character grid. It supports a limited ASCII/Unicode palette, dithering, and orthographic camera effects.

### Garbage Collection Telemetry

Go's GC is treated as a tunable engine feature.

* **Non-STW Polling:** GC pause events are captured using `runtime/metrics` and `debug.ReadGCStats`, avoiding the Stop-The-World penalty of `runtime.ReadMemStats`.
* **Live Tuning:** The debug overlay plots frame times and injects red vertical spikes on the graph exactly when the GC pauses the application. Exposing `debug.SetGCPercent` and `debug.SetMemoryLimit` directly to UI sliders allows developers to visually balance memory consumption against pause latency in real-time.

## Current Roadmap

The architecture is largely complete, but the following features are pending implementation:

1. **Texture and Sprites Atlases (`renderer/sprite.go`)**:
   * UV transform slicing for parsing `.png` sprite sheets, allowing the migration from procedural geometry to standard pixel-art assets.
2. **Dedicated Scene State (`scene/`)**:
   * Extracting scene management into an isolated package with visual transition buffers (e.g., crossfades, push/pop cut effects).
3. **Gamepad Polling (`input/gamepad.go`)**:
   * Implementing joystick axis deadzones and mapping.
4. **File Cleanup**:
   * Some planned independent files were skipped and their logic was just shoved into the main package files to save time:
     * audio/mixer.go: Doesn't exist. The volume groups and distance-attenuation panning logic are currently just baked directly into synth.go and audio.go.
     * physics/layers.go: Doesn't exist. The 32-bit collision layer mask checks are just floating inline inside the physics.go integration loop.
5. **ECS Concurrent Write Safety**:
    * Archetype storage needs read/write separation when parallel systems execute.
    * Decide between per-archetype RWMutex vs channel-based write batching before implementing.

## Install Examples

### Quick install (any platform)

**Unix** (Linux, macOS, WSL):

```sh
curl -sS https://raw.githubusercontent.com/ujjwalvivek/loom/main/examples/scripts/install.sh | sh
loom-mario-term
```

**Windows** (PowerShell):

```powershell
irm https://raw.githubusercontent.com/ujjwalvivek/loom/main/examples/scripts/install.ps1 | iex
loom-mario-term
```

The terminal version runs anywhere and has no dependencies.
The GUI version (`loom-mario`) requires OpenGL 3.3 and GLFW.

### Direct download

Pre-built binaries for all platforms are available on the
[GitHub Releases](https://github.com/ujjwalvivek/loom/releases) page,
with a [download page](https://loom.ujjwalvivek.dev) hosted on Cloudflare Pages.

### Build from source

```sh
go install github.com/ujjwalvivek/loom/examples/mario-term@latest
# or:
go install github.com/ujjwalvivek/loom/examples/mario@latest
```

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE).
