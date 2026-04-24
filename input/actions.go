package input

import "github.com/go-gl/glfw/v3.3/glfw"

// BindAction associates an action name with one or more physical keys.
func (is *InputSystem) BindAction(action string, keys ...glfw.Key) {
	is.mu.Lock()
	defer is.mu.Unlock()
	is.actions[action] = append(is.actions[action], keys...)
}

// UpdateActions updates action state evaluations based on mapped keys.
func (is *InputSystem) UpdateActions() {
	is.mu.Lock()
	defer is.mu.Unlock()

	for action, keys := range is.actions {
		var maxState KeyState = KeyNone
		for _, key := range keys {
			if key < 0 || int(key) >= len(is.keys) {
				continue
			}
			keyState := is.keys[key]
			if keyState == KeyPressed {
				maxState = KeyPressed
				break
			} else if keyState == KeyDown && maxState != KeyPressed {
				maxState = KeyDown
			} else if keyState == KeyReleased && maxState == KeyNone {
				maxState = KeyReleased
			}
		}
		is.actionStates[action] = maxState
	}
}

// ActionDown returns true if the mapped action is active.
func (s *InputState) ActionDown(action string) bool {
	state, ok := s.actionStates[action]
	if !ok {
		return false
	}
	return state == KeyDown || state == KeyPressed
}

// ActionPressed returns true if the action was triggered this frame.
func (s *InputState) ActionPressed(action string) bool {
	state, ok := s.actionStates[action]
	if !ok {
		return false
	}
	return state == KeyPressed
}

// ActionReleased returns true if the action was ended this frame.
func (s *InputState) ActionReleased(action string) bool {
	state, ok := s.actionStates[action]
	if !ok {
		return false
	}
	return state == KeyReleased
}
