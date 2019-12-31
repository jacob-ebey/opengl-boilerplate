package main

import (
	"fmt"
	"time"

	"github.com/faiface/glhf"
	"github.com/faiface/mainthread"
	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/paulbellamy/ratecounter"

	"github.com/jacob-ebey/opengl-boilerplate/game"
	"github.com/jacob-ebey/opengl-boilerplate/scene"
	"github.com/jacob-ebey/opengl-boilerplate/ux"
)

const (
	windowWidth  = 500
	windowHeight = 200
	targetFps    = 300

	timerInc = 1
	clearR   = 0
	clearG   = 1
	clearB   = 1
	clearA   = 1
)

type runtimeObject struct {
	update       chan struct{}
	draw         chan struct{}
	doneUpdating <-chan struct{}
	doneDrawing  <-chan struct{}
}

func createRuntimeObjects(sceneObjects []game.SceneObject) []*runtimeObject {
	runtimeObjects := make([]*runtimeObject, len(sceneObjects))

	for index, obj := range sceneObjects {
		var (
			update       chan struct{}   = nil
			draw         chan struct{}   = nil
			doneUpdating <-chan struct{} = nil
			doneDrawing  <-chan struct{} = nil
		)

		obj.Initialize()

		if obj, ok := obj.(game.UpdatingSceneObject); ok {
			update = make(chan struct{})
			doneUpdating = obj.InitializeUpdate(update)
		}

		if obj, ok := obj.(game.DrawingSceneObject); ok {
			draw = make(chan struct{})
			doneDrawing = obj.InitializeDraw(draw)
		}

		runtimeObjects[index] = &runtimeObject{
			update:       update,
			draw:         draw,
			doneUpdating: doneUpdating,
			doneDrawing:  doneDrawing,
		}
	}

	return runtimeObjects
}

func updateRuntimeObjects(runtimeObjects []*runtimeObject, updateValue struct{}) {
	for _, obj := range runtimeObjects {
		if obj.update != nil {
			obj.update <- updateValue
		}
	}

	for _, obj := range runtimeObjects {
		<-obj.doneUpdating
	}
}

func drawRuntimeObjects(runtimeObjects []*runtimeObject, drawValue struct{}) {
	for _, obj := range runtimeObjects {
		if obj.draw != nil && obj.doneDrawing != nil {
			obj.draw <- drawValue
			<-obj.doneDrawing
		}
	}
}

func shouldExit(window ux.Window, keyChan <-chan *ux.KeyAction) bool {
	if window.ShouldClose() {
		return true
	}

	select {
	case event := <-keyChan:
		if event.Action == glfw.Press && event.Key == glfw.KeyEscape {
			return true
		}

	default:
		break
	}

	return false
}

func run() {
	window, err := ux.NewGlfwWindow("opengl-boilerplate", windowWidth, windowHeight)
	if err != nil {
		panic(err)
	}

	defer window.Destroy()

	keyChan := window.KeyChannel()

	mainthread.Call(func() {
		glhf.Init()
	})

	renderTick := time.Tick(time.Second / targetFps)

	sceneObjects := []game.SceneObject{
		&scene.DelayGameObject{},
		&scene.DelayGameObject{},
	}

	runtimeObjects := createRuntimeObjects(sceneObjects)

	gameLoopTimer := ratecounter.NewRateCounter(time.Second)
	renderLoopTimer := ratecounter.NewRateCounter(time.Second)

	updateValue := struct{}{}
	drawValue := struct{}{}
	i := 0

	for !shouldExit(window, keyChan) {
		gameLoopTimer.Incr(timerInc)

		updateRuntimeObjects(runtimeObjects, updateValue)

		select {
		case <-renderTick:
			i++

			renderLoopTimer.Incr(timerInc)

			drawRuntimeObjects(runtimeObjects, drawValue)

			mainthread.Call(func() {
				glhf.Clear(clearR, clearG, clearB, clearA)
			})

			window.Update()

			if i%60 == 0 {
				fmt.Printf("\ngameloop updates per sec  : %d", gameLoopTimer.Rate())
				fmt.Printf("\nrenderloop updates per sec: %d\n", renderLoopTimer.Rate())
			}

		default:
			continue
		}
	}
}

func main() {
	mainthread.Run(run)
}
