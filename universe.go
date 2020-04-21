package main

import (
	"errors"
	"github.com/markbates/errx"
	"sync"
)

type Universe struct {
	title  string
	width  int
	height int
	scale  float64
	worlds map[int]*World
}

var universe *Universe

func newUniverse( _width, _height int, _scale float64) *Universe {
	ws := make(map[int]*World)
	u := Universe{
		width:  _width,
		height: _height,
		scale:  _scale,
		worlds: ws,
	}
	return &u
}
func AddWorld() (int, error) {
	if universe == nil {
		return 0, errors.New("universe is nil")
	}

	w := InitWorld()
	universe.worlds[w.MyId] = w
	return w.MyId, nil
}

func GetWorld(idWorld int) (*World, error) {
	if universe == nil {
		return nil, errors.New("universe is nil")
	}
	return universe.worlds[idWorld], nil
}
func BornInWorld(idWorld int) (int, error) {
	w, err := GetWorld(idWorld)
	if err != nil {
		return 0, errx.Wrap(err, "Can not born")
	}
	if idUnit, err := w.AddUnit(); err != nil {
		return idUnit, nil
	}
	return 0, errx.Wrap(err, "Can not born")
}

func KillInWorld(idWorld, idUnit int) error {
	w, err := GetWorld(idWorld)
	if err != nil {
		return errx.Wrap(err, "Can not kill unit")
	}
	if err := w.DeleteUnit(idUnit); err != nil {
		return nil
	}
	return errx.Wrap(err, "Can not kill unit")
}

var theGod *sync.Once

func BigBang(width, height int, scale float64) {
	theGod.Do(func() {
		universe = newUniverse(width, height, scale)
	})

}


