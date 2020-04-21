package main

type Unit struct {
	Id        int
	X         float64
	Y         float64
	Frame     int32
	Skin      string
	Action    string
	Speed     float64
	Direction Direction
	Side      Direction
}
