package termrenderer

import (
	"sync"
	"time"

	"github.com/go-gl/glfw/v3.3/glfw"
	"golang.org/x/term"
)

// TermInput manages non-blocking raw terminal input.
type TermInput struct {
	mu             sync.Mutex
	oldState       *term.State
	lastSeen       map[glfw.Key]time.Time
	pressed        map[glfw.Key]bool
	pendingPressed map[glfw.Key]bool
	held           map[glfw.Key]bool // tracks keys held down (supported on Windows Console API)
	isWindows      bool              // tracks if native windows console events are active
	focused        bool              // tracks console focus state
}

// KeyPressed returns true if the key was freshly pressed this exact frame.
func (ti *TermInput) KeyPressed(key glfw.Key) bool {
	ti.mu.Lock()
	defer ti.mu.Unlock()
	return ti.pressed[key]
}


// Update swaps the pending inputs to the active pressed state and resets pending.
// Call this at the end of the game loop tick.
func (ti *TermInput) Update() {
	ti.mu.Lock()
	defer ti.mu.Unlock()

	// Reset pressed state
	for k := range ti.pressed {
		ti.pressed[k] = false
	}
	// Copy pending to active pressed state
	for k, v := range ti.pendingPressed {
		if v {
			ti.pressed[k] = true
		}
	}
	// Clear pending pressed map
	ti.pendingPressed = make(map[glfw.Key]bool)
}
