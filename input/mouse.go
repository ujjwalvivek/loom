package input

import "github.com/go-gl/glfw/v3.3/glfw"

// MouseDown returns true if the specified button is held down.
func (s *InputState) MouseDown(button glfw.MouseButton) bool {
	if button < 0 || int(button) >= len(s.mouseButtons) {
		return false
	}
	return s.mouseButtons[button] == KeyDown || s.mouseButtons[button] == KeyPressed
}

// MousePressed returns true if the button transition occurred in the current frame.
func (s *InputState) MousePressed(button glfw.MouseButton) bool {
	if button < 0 || int(button) >= len(s.mouseButtons) {
		return false
	}
	return s.mouseButtons[button] == KeyPressed
}

// MouseReleased returns true if the button was released in the current frame.
func (s *InputState) MouseReleased(button glfw.MouseButton) bool {
	if button < 0 || int(button) >= len(s.mouseButtons) {
		return false
	}
	return s.mouseButtons[button] == KeyReleased
}

// MousePos returns the cursor coordinates.
func (s *InputState) MousePos() (float32, float32) {
	return float32(s.mouseX), float32(s.mouseY)
}

// ScrollDelta returns the mouse wheel scroll offsets.
func (s *InputState) ScrollDelta() (float32, float32) {
	return float32(s.scrollX), float32(s.scrollY)
}
