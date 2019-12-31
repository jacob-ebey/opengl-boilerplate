// Package ux provides abstractions for the UX as well as base implementations.
package ux

import (
	"fmt"
	"sync"

	"github.com/faiface/mainthread"
	"github.com/go-gl/glfw/v3.1/glfw"
)

// KeyAction contains the information about the keypress.
type KeyAction struct {
	Key      glfw.Key
	Scancode int
	Action   glfw.Action
	Mods     glfw.ModifierKey
}

// Window is the base abstraction around a window.
type Window interface {
	// ShouldClose returns true if the window should close; otherwise false.
	ShouldClose() bool
	// Update should be called on the mainthread periodically.
	Update()
	// KeyChannel creates a new channel that will receive key actions.
	KeyChannel() <-chan *KeyAction
	// Destroy destroys the window.
	Destroy()
}

// GlfwWindow is a glfl implementation of Window.
type GlfwWindow struct {
	window         *glfw.Window
	keyChannelsMux sync.Mutex
	keyChannels    []chan *KeyAction
}

// NewGlfwWindow creates a new *GlfwWindow that implements Window.
func NewGlfwWindow(title string, width, height int) (Window, error) {
	var (
		result       = &GlfwWindow{}
		err    error = nil
	)

	mainthread.Call(func() {
		if err = glfw.Init(); err != nil {
			err = fmt.Errorf("failed to initialize glfw: %s", err.Error())
			return
		}

		glfw.WindowHint(glfw.ContextVersionMajor, 3)
		glfw.WindowHint(glfw.ContextVersionMinor, 3)
		glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
		glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
		glfw.WindowHint(glfw.Resizable, glfw.False)

		var win *glfw.Window
		win, err = glfw.CreateWindow(width, height, title, nil, nil)
		if err != nil {
			err = fmt.Errorf("failed to create window: %s", err.Error())
			return
		}
		result.window = win

		win.MakeContextCurrent()
		win.SetKeyCallback(result.keyCallback)
		glfw.SwapInterval(0)
	})

	return result, err
}

func (w *GlfwWindow) ShouldClose() bool {
	return mainthread.CallVal(func() interface{} {
		return w.window.ShouldClose()
	}).(bool)
}

func (w *GlfwWindow) Update() {
	mainthread.Call(func() {
		w.window.SwapBuffers()
		glfw.PollEvents()
	})
}

func (w *GlfwWindow) Destroy() {
	mainthread.Call(func() {
		w.window.Destroy()
		glfw.Terminate()
	})
}

func (w *GlfwWindow) KeyChannel() <-chan *KeyAction {
	c := make(chan *KeyAction)

	w.keyChannelsMux.Lock()
	w.keyChannels = append(w.keyChannels, c)
	w.keyChannelsMux.Unlock()

	return c
}

func (w *GlfwWindow) keyCallback(window *glfw.Window, key glfw.Key, scancode int,
	action glfw.Action, mods glfw.ModifierKey) {
	keyAction := &KeyAction{
		Key:      key,
		Scancode: scancode,
		Action:   action,
		Mods:     mods,
	}

	w.keyChannelsMux.Lock()
	for _, c := range w.keyChannels {
		go func(c chan *KeyAction) {
			c <- keyAction
		}(c)
	}
	w.keyChannelsMux.Unlock()
}
