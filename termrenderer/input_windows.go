//go:build windows

package termrenderer

import (
	"os"
	"syscall"
	"time"
	"unsafe"

	"github.com/go-gl/glfw/v3.3/glfw"
	"golang.org/x/sys/windows"
	"golang.org/x/term"
)

const (
	KEY_EVENT        = 0x0001
	FOCUS_EVENT      = 0x0010
	STD_INPUT_HANDLE = -10
)

type KEY_EVENT_RECORD struct {
	KeyDown         int32
	RepeatCount     uint16
	VirtualKeyCode  uint16
	VirtualScanCode uint16
	UnicodeChar     uint16
	ControlKeyState uint32
}

type FOCUS_EVENT_RECORD struct {
	SetFocus int32
}

type INPUT_RECORD struct {
	EventType uint16
	_         uint16 // Padding
	Event     [16]byte
}

var (
	kernel32             = syscall.NewLazyDLL("kernel32.dll")
	procReadConsoleInput = kernel32.NewProc("ReadConsoleInputW")

	user32               = syscall.NewLazyDLL("user32.dll")
	procGetAsyncKeyState = user32.NewProc("GetAsyncKeyState")
)

// NewTermInput sets the terminal to raw mode (to disable echo) and starts the Windows Console API polling loop.
func NewTermInput() *TermInput {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}

	// Disable ENABLE_VIRTUAL_TERMINAL_INPUT (0x0200) to ensure we receive
	// standard VirtualKeyCodes for arrow keys and regular keys rather than ANSI sequences.
	handle := windows.Handle(os.Stdin.Fd())
	var mode uint32
	if err := windows.GetConsoleMode(handle, &mode); err == nil {
		mode &^= 0x0200
		windows.SetConsoleMode(handle, mode)
	}

	ti := &TermInput{
		oldState:       oldState,
		lastSeen:       make(map[glfw.Key]time.Time),
		pressed:        make(map[glfw.Key]bool),
		pendingPressed: make(map[glfw.Key]bool),
		held:           make(map[glfw.Key]bool),
		isWindows:      true,
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
	handle := syscall.Handle(os.Stdin.Fd())

	buffer := make([]INPUT_RECORD, 16)
	for {
		var read uint32
		r1, _, _ := procReadConsoleInput.Call(
			uintptr(handle),
			uintptr(unsafe.Pointer(&buffer[0])),
			uintptr(len(buffer)),
			uintptr(unsafe.Pointer(&read)),
		)

		if r1 == 0 || read == 0 {
			time.Sleep(5 * time.Millisecond)
			continue
		}

		ti.mu.Lock()
		now := time.Now()

		for i := 0; i < int(read); i++ {
			rec := buffer[i]
			if rec.EventType == FOCUS_EVENT {
				focusEvent := (*FOCUS_EVENT_RECORD)(unsafe.Pointer(&rec.Event[0]))
				ti.focused = focusEvent.SetFocus != 0
			} else if rec.EventType == KEY_EVENT {
				keyEvent := (*KEY_EVENT_RECORD)(unsafe.Pointer(&rec.Event[0]))
				key := getGlfwKey(keyEvent.VirtualKeyCode, keyEvent.UnicodeChar)
				
				if key != 0 {
					isDown := keyEvent.KeyDown != 0
					if isDown {
						// Trigger KeyPressed only on the initial down transition
						if !ti.held[key] {
							ti.pendingPressed[key] = true
						}
						ti.held[key] = true
						ti.lastSeen[key] = now
					} else {
						ti.held[key] = false
					}
				}
			}
		}
		ti.mu.Unlock()
	}
}

const (
	VK_LSHIFT = 0xA0
	VK_RSHIFT = 0xA1
	VK_SHIFT  = 0x10
	VK_ESCAPE = 0x1B
	VK_SPACE  = 0x20
	VK_RETURN = 0x0D
	VK_LEFT   = 0x25
	VK_UP     = 0x26
	VK_RIGHT  = 0x27
	VK_DOWN   = 0x28
)

func getGlfwKey(vk uint16, char uint16) glfw.Key {
	if vk != 0 {
		key := mapVirtualKey(vk)
		if key != 0 {
			return key
		}
	}
	if char != 0 {
		switch char {
		case 'w', 'W':
			return glfw.KeyW
		case 'a', 'A':
			return glfw.KeyA
		case 's', 'S':
			return glfw.KeyS
		case 'd', 'D':
			return glfw.KeyD
		case ' ':
			return glfw.KeySpace
		case 27: // ESC
			return glfw.KeyEscape
		case 13, 10: // Enter / LF
			return glfw.KeyEnter
		}
	}
	return 0
}

func mapVirtualKey(vk uint16) glfw.Key {
	switch vk {
	case 0x57: // 'W'
		return glfw.KeyW
	case 0x41: // 'A'
		return glfw.KeyA
	case 0x53: // 'S'
		return glfw.KeyS
	case 0x44: // 'D'
		return glfw.KeyD
	case VK_SPACE:
		return glfw.KeySpace
	case VK_RETURN:
		return glfw.KeyEnter
	case VK_ESCAPE:
		return glfw.KeyEscape
	case VK_LEFT:
		return glfw.KeyLeft
	case VK_UP:
		return glfw.KeyUp
	case VK_RIGHT:
		return glfw.KeyRight
	case VK_DOWN:
		return glfw.KeyDown
	case VK_SHIFT, VK_LSHIFT, VK_RSHIFT:
		return glfw.KeyLeftShift
	}
	return 0
}

// KeyDown returns true if the key is currently held down.
func (ti *TermInput) KeyDown(key glfw.Key) bool {
	ti.mu.Lock()
	focused := ti.focused
	ti.mu.Unlock()

	if !focused {
		return false
	}

	vk := glfwToWindowsVk(key)
	if vk == 0 {
		ti.mu.Lock()
		defer ti.mu.Unlock()
		return ti.held[key]
	}

	r1, _, _ := procGetAsyncKeyState.Call(uintptr(vk))
	return (r1 & 0x8000) != 0
}

func glfwToWindowsVk(key glfw.Key) int {
	switch key {
	case glfw.KeyW:
		return 0x57
	case glfw.KeyA:
		return 0x41
	case glfw.KeyS:
		return 0x53
	case glfw.KeyD:
		return 0x44
	case glfw.KeySpace:
		return 0x20
	case glfw.KeyEnter:
		return 0x0D
	case glfw.KeyEscape:
		return 0x1B
	case glfw.KeyLeft:
		return 0x25
	case glfw.KeyUp:
		return 0x26
	case glfw.KeyRight:
		return 0x27
	case glfw.KeyDown:
		return 0x28
	case glfw.KeyLeftShift, glfw.KeyRightShift:
		return 0x10 // VK_SHIFT
	}
	return 0
}
