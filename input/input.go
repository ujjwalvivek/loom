package input

import (
	"sync"

	"github.com/go-gl/glfw/v3.3/glfw"
)

type KeyState byte

const (
	KeyNone     KeyState = 0
	KeyDown     KeyState = 1 // Key is held down (second frame onwards)
	KeyPressed  KeyState = 2 // Key was pressed this frame
	KeyReleased KeyState = 3 // Key was released this frame
)

type InputSystem struct {
	mu           sync.RWMutex
	keys         [glfw.KeyLast + 1]KeyState
	mouseButtons [glfw.MouseButtonLast + 1]KeyState
	mouseX       float64
	mouseY       float64
	scrollX      float64
	scrollY      float64

	// Action mappings
	actions      map[string][]glfw.Key
	actionStates map[string]KeyState
}

type InputState struct {
	keys         [glfw.KeyLast + 1]KeyState
	mouseButtons [glfw.MouseButtonLast + 1]KeyState
	mouseX       float64
	mouseY       float64
	scrollX      float64
	scrollY      float64
	actionStates map[string]KeyState
}

func NewInputSystem() *InputSystem {
	return &InputSystem{
		actions:      make(map[string][]glfw.Key),
		actionStates: make(map[string]KeyState),
	}
}

// RegisterCallbacks links input listener routines to the GLFW window context.
func (is *InputSystem) RegisterCallbacks(window *glfw.Window) {
	window.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		if key < 0 || int(key) >= len(is.keys) {
			return
		}
		is.mu.Lock()
		defer is.mu.Unlock()
		switch action {
		case glfw.Press:
			is.keys[key] = KeyPressed
		case glfw.Release:
			is.keys[key] = KeyReleased
		}
	})

	window.SetMouseButtonCallback(func(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
		if button < 0 || int(button) >= len(is.mouseButtons) {
			return
		}
		is.mu.Lock()
		defer is.mu.Unlock()
		switch action {
		case glfw.Press:
			is.mouseButtons[button] = KeyPressed
		case glfw.Release:
			is.mouseButtons[button] = KeyReleased
		}
	})

	window.SetCursorPosCallback(func(w *glfw.Window, xpos float64, ypos float64) {
		is.mu.Lock()
		defer is.mu.Unlock()
		is.mouseX = xpos
		is.mouseY = ypos
	})

	window.SetScrollCallback(func(w *glfw.Window, xoff float64, yoff float64) {
		is.mu.Lock()
		defer is.mu.Unlock()
		is.scrollX = xoff
		is.scrollY = yoff
	})
}

// Snapshot returns a copy of the current input state suitable for concurrent read access.
func (is *InputSystem) Snapshot() *InputState {
	is.mu.RLock()
	defer is.mu.RUnlock()

	snap := &InputState{
		mouseX:       is.mouseX,
		mouseY:       is.mouseY,
		scrollX:      is.scrollX,
		scrollY:      is.scrollY,
		actionStates: make(map[string]KeyState),
	}
	copy(snap.keys[:], is.keys[:])
	copy(snap.mouseButtons[:], is.mouseButtons[:])
	for action, state := range is.actionStates {
		snap.actionStates[action] = state
	}
	return snap
}

// PostUpdate resolves key press triggers to active states and cleans temporary deltas.
func (is *InputSystem) PostUpdate() {
	is.mu.Lock()
	defer is.mu.Unlock()

	is.scrollX = 0
	is.scrollY = 0

	for i, state := range is.keys {
		switch state {
		case KeyPressed:
			is.keys[i] = KeyDown
		case KeyReleased:
			is.keys[i] = KeyNone
		}
	}

	for i, state := range is.mouseButtons {
		switch state {
		case KeyPressed:
			is.mouseButtons[i] = KeyDown
		case KeyReleased:
			is.mouseButtons[i] = KeyNone
		}
	}
}
