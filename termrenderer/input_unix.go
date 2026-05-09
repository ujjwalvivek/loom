//go:build !windows

package termrenderer

import (
	"os"
	"time"

	"github.com/go-gl/glfw/v3.3/glfw"
	"golang.org/x/term"
)

// NewTermInput sets the terminal to raw mode and starts the input polling loop.
func NewTermInput() *TermInput {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}

	ti := &TermInput{
		oldState:       oldState,
		lastSeen:       make(map[glfw.Key]time.Time),
		pressed:        make(map[glfw.Key]bool),
		pendingPressed: make(map[glfw.Key]bool),
		held:           make(map[glfw.Key]bool),
		isWindows:      false,
		focused:        true,
	}

	go ti.poll()

	return ti
}

// Shutdown restores the terminal state.
func (ti *TermInput) Shutdown() {
	term.Restore(int(os.Stdin.Fd()), ti.oldState)
}

func (ti *TermInput) poll() {
	buf := make([]byte, 32)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil || n == 0 {
			time.Sleep(1 * time.Millisecond)
			continue
		}

		ti.mu.Lock()
		now := time.Now()

		registerKey := func(key glfw.Key) {
			if key != 0 {
				if time.Since(ti.lastSeen[key]) > 150*time.Millisecond {
					ti.pendingPressed[key] = true
				}
				ti.lastSeen[key] = now
			}
		}

		i := 0
		for i < n {
			if buf[i] == 27 { // Escape sequence or ESC key
				if i+2 < n && buf[i+1] == '[' {
					var key glfw.Key
					switch buf[i+2] {
					case 'A':
						key = glfw.KeyUp
					case 'B':
						key = glfw.KeyDown
					case 'C':
						key = glfw.KeyRight
					case 'D':
						key = glfw.KeyLeft
					}
					if key != 0 {
						registerKey(key)
						i += 3
						continue
					}
				}
				// Standalone ESC
				registerKey(glfw.KeyEscape)
				i++
			} else {
				var key glfw.Key
				switch buf[i] {
				case 'w', 'W':
					key = glfw.KeyW
				case 'a', 'A':
					key = glfw.KeyA
				case 's', 'S':
					key = glfw.KeyS
				case 'd', 'D':
					key = glfw.KeyD
				case ' ':
					key = glfw.KeySpace
				case '\r', '\n':
					key = glfw.KeyEnter
				case 3: // Ctrl+C
					key = glfw.KeyEscape
				}
				registerKey(key)
				i++
			}
		}
		ti.mu.Unlock()
	}
}

// KeyDown returns true if the key is currently held down.
func (ti *TermInput) KeyDown(key glfw.Key) bool {
	ti.mu.Lock()
	defer ti.mu.Unlock()
	return time.Since(ti.lastSeen[key]) < 150*time.Millisecond
}
