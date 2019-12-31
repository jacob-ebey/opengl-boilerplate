// Package game provides interfaces for creating your own assets and actions.
package game

// SceneObject is the base of a scene object.
type SceneObject interface {
	Initialize()
}

// UpdatingSceneObject provides a way for scene objects to participate in the
// asynchronous update loop.
type UpdatingSceneObject interface {
	InitializeUpdate(update chan struct{}) <-chan struct{}
}

// DrawingSceneObject provides a way for scene objects to participate in the
// synchronous drawing loop.
type DrawingSceneObject interface {
	InitializeDraw(draw chan struct{}) <-chan struct{}
}
