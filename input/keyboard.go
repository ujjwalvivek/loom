package input

import "github.com/go-gl/glfw/v3.3/glfw"

// KeyDown returns true if the key is currently held down or was pressed this frame.
func (s *InputState) KeyDown(key glfw.Key) bool {
	if key < 0 || int(key) >= len(s.keys) {
		return false
	}
	return s.keys[key] == KeyDown || s.keys[key] == KeyPressed
}

// KeyPressed returns true only if the key transition occurred in the current frame.
func (s *InputState) KeyPressed(key glfw.Key) bool {
	if key < 0 || int(key) >= len(s.keys) {
		return false
	}
	return s.keys[key] == KeyPressed
}

// KeyReleased returns true only if the key was released in the current frame.
func (s *InputState) KeyReleased(key glfw.Key) bool {
	if key < 0 || int(key) >= len(s.keys) {
		return false
	}
	return s.keys[key] == KeyReleased
}
