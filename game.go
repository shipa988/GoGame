//go:generate statik -src=resources -include=*.png,*.txt,*.tmx
package main

import "image"

type Frames struct {
	Frames []image.Image
	image.Config
}
const (
	UnitActionIdle = "idle"
	UnitActionMove = "run"
)

type EventType int

const (
	Init_Event EventType = iota
	Delete_Event
	Move_Event
	Idle_Event
)

type Event struct {
}

type Direction int

const (
	Direction_left  Direction = 0
	Direction_right Direction = iota
	Direction_up
	Direction_down
)