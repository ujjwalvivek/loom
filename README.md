# Loom Engine

Loom is a lightweight, custom 2D game engine in Go. 

Loom is built directly on top of **OpenGL 3.3** for graphics, **GLFW 3.3** for window management and inputs, and **Oto v3** for low-level audio streaming. It's concurrent. It's fast. It's bloat-free. It's everything you would want to build a 2D game.

> Since the engine uses OpenGL and GLFW, your system needs to support graphics rendering. On Windows, GLFW DLLs are loaded dynamically, so you don't need to configure compile toolchains. Make sure you have Go installed (version 1.25 or newer is recommended).

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

### Sub Systems

The engine is broken down into small, focused modules under the root directory:

*   **`renderer/`**: An OpenGL 3.3 batch renderer. It batches sprites, textured shapes, and solids into a single draw call where possible. It also includes an orthographic camera that supports target tracking, screen shake, and smooth zooming.
*   **`physics/`**: An AABB swept-collision physics system. It uses a **multi-pass sliding solver** (up to 4 passes per tick). When a player jumps or runs into a wall, it slides their velocity vector smoothly along the surface rather than stopping them dead in their tracks, making it ideal for snappy platformer movement.
*   **`audio/`**: A retro-themed procedural audio system. It features a chiptune synthesizer supporting square, triangle, sawtooth, sine, and noise generators. Sound effects utilize custom ADSR envelopes, and background music is driven by a programmable step sequencer.
*   **`ecs/`**: A lightweight Entity Component System (ECS) to manage positions, velocities, and physics components.
*   **`input/`**: Keyboard and mouse wrappers on top of GLFW to poll current key states or listen for discrete keypresses.
*   **`math/`**: Standard 2D vector and rectangle math structures, plus a linear feedback shift register (LFSR) for retro noise generation.
*   **`perf/`**: GUI overlays for performance related metrics.

## License

This project is licensed under the MIT License. See the [LICENSE](file:///f:/_Engine/loom/LICENSE).
