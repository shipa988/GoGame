package main

import "errors"

var _ Worlder = (*World)(nil)

type Worlder interface {
	AddUnit() (int, error)
	DeleteUnit(id int) error
	HandleEvent(e Event) error
	InitMap() (*Maper, error)
	LoadMap() error
}

type World struct {
	MyId    int
	MyUnits []Unit
	MyMap   Maper
}

/*func (w *World) initMap(mapid int) (*Maper, error) {
	var m MapbyTMX
	m.LoadMap(mapid)
	return *m,nil
}

func (w *World) loadMap(m Maper) error {
	if m == nil {
		return errors.New("can't load map in world, map is nil")
	}
	w.MyMap = m
	return nil
}*/

func InitWorld(mapid int) (*World, error) {
	var w World

	m,err:=w.initMap(mapid)
	if err!=nil {
		return nil, errors.New("can't InitWorld")
	}
	w.loadMap(*m)
	return &w, nil
}

func (w *World) AddUnit() (int, error) {
	panic("implement me")
}

func (w *World) DeleteUnit(id int) error {
	panic("implement me")
}

func (w *World) HandleEvent(event Event) error {
	panic("implement me")
}
