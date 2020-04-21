package main

var _ Worlder = (*World)(nil)

type Worlder interface {
	init() error
	AddUnit() (int, error)
	DeleteUnit(id int) error
	HandleEvent(e Event) error
}

type World struct {
	MyId  int
	units []Unit
}

func newWorld() *World {
	var w World
	w.init()
	return &w
}

func (w *World) init() error {
	panic("implement me")
}

func (w *World) AddUnit() (int, error) {
	panic("implement me")
}

func (w *World) DeleteUnit(id int) error {
	panic("implement me")
}

func (w *World) HandleEvent(event Event) error {

}

func InitWorld() *World {
	return newWorld()
}
