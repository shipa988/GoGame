package main

import (
	e "github.com/hajimehoshi/ebiten"
	"testing"
)

func BenchmarkUpdate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		e.Run(Update, config.width, config.height, config.scale, config.title)
	}

}
