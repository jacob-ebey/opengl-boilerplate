// Package scene is where you define your own scene objects.
package scene

import (
	"time"

	"github.com/faiface/mainthread"
)

const (
	gameObjectDelayMicroseconds = 10
)

// DelayGameObject is an example scene object that delays for 10ms
// in both it's update and render logic.
type DelayGameObject struct{}

// Initialize is where you would add your initialization logic.
func (o *DelayGameObject) Initialize() {
}

// Shows how to span a goroutine and participate in the asynchronous update loop.
func (o *DelayGameObject) InitializeUpdate(update chan struct{}) <-chan struct{} {
	doneUpdating := make(chan struct{})

	go func() {
		for {
			// wait for the next update loop
			<-update

			time.Sleep(time.Microsecond * gameObjectDelayMicroseconds)

			// report your're done
			doneUpdating <- struct{}{}
		}
	}()

	return doneUpdating
}

// Shows how to span a goroutine and participate in the synchronous draw loop.
func (o *DelayGameObject) InitializeDraw(draw chan struct{}) <-chan struct{} {
	doneDrawing := make(chan struct{})

	go func() {
		for {
			// wait for the next draw loop
			<-draw

			mainthread.Call(func() {
				time.Sleep(time.Microsecond * gameObjectDelayMicroseconds)

				// notify we are done drawing. do it inside the mainthread.Call
				// so we don't have to wait for the context switch.
				doneDrawing <- struct{}{}
			})
		}
	}()

	return doneDrawing
}
