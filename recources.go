package main

import (
	"fmt"
	e "github.com/hajimehoshi/ebiten"
	"github.com/rakyll/statik/fs"
	_ "github.com/shipa988/ebitentest/statik" // TODO: Replace with the absolute import path
	"image"
	"image/png"
	"log"
	"os"
)

func prepareLevelImage() (*e.Image, error) {
	tileSize := config.width / 45
	level := LoadLevel()
	width := len(level[0])
	height := len(level)
	levelImage, _ := e.NewImage(width*tileSize, height*tileSize, e.FilterDefault)

	for i := 0; i < width; i++ {
		for j := 0; j < height; j++ {
			op := &e.DrawImageOptions{}
			op.GeoM.Translate(float64(i*tileSize), float64(j*tileSize))

			img, err := e.NewImageFromImage(frames[level[j][i]].Frames[0], e.FilterDefault)
			if err != nil {
				log.Println(err)
				return levelImage, err
			}
			err = levelImage.DrawImage(img, op)
			if err != nil {
				log.Println(err)
				return levelImage, err
			}
		}
	}

	return levelImage, nil
}
func LoadResources() (map[string]Frames, error) {
	images := map[string]image.Image{}
	cfgs := map[string]image.Config{}
	sprites := map[string]Frames{}
	statikFS, err := fs.New()
	if err != nil {
		log.Fatal(err)
	}

	fs.Walk(statikFS, "/", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		fmt.Println(path)

		if info.IsDir() {
			return nil
		}
		// Access individual files by their paths.
		f, err := statikFS.Open(path)
		if err != nil {
			log.Fatal(err)
		}

		img, err := png.Decode(f)
		if err != nil {
			return err
		}
		cfg, err := png.DecodeConfig(f)

		if cfg.ColorModel == nil {
			cfg.Width = img.Bounds().Max.X
			cfg.Height = img.Bounds().Max.Y
			cfg.ColorModel = img.ColorModel()
		}
		images[info.Name()] = img
		cfgs[info.Name()] = cfg
		f.Close()
		return nil
	})

	sprites["big_demon_idle"] = Frames{
		Frames: []image.Image{
			images["big_demon_idle_anim_f0.png"],
			images["big_demon_idle_anim_f1.png"],
			images["big_demon_idle_anim_f2.png"],
			images["big_demon_idle_anim_f3.png"],
		},
		Config: cfgs["big_demon_idle_anim_f0.png"],
	}
	sprites["big_demon_run"] = Frames{
		Frames: []image.Image{
			images["big_demon_run_anim_f0.png"],
			images["big_demon_run_anim_f1.png"],
			images["big_demon_run_anim_f2.png"],
			images["big_demon_run_anim_f3.png"],
		},
		Config: cfgs["big_demon_run_anim_f0.png"],
	}

	sprites["big_zombie_idle"] = Frames{
		Frames: []image.Image{
			images["big_zombie_idle_anim_f0.png"],
			images["big_zombie_idle_anim_f1.png"],
			images["big_zombie_idle_anim_f2.png"],
			images["big_zombie_idle_anim_f3.png"],
		},
		Config: cfgs["big_zombie_idle_anim_f0.png"],
	}
	sprites["big_zombie_run"] = Frames{
		Frames: []image.Image{
			images["big_zombie_run_anim_f0.png"],
			images["big_zombie_run_anim_f1.png"],
			images["big_zombie_run_anim_f2.png"],
			images["big_zombie_run_anim_f3.png"],
		},
		Config: cfgs["big_zombie_run_anim_f0.png"],
	}

	sprites["elf_f_idle"] = Frames{
		Frames: []image.Image{
			images["elf_f_idle_anim_f0.png"],
			images["elf_f_idle_anim_f1.png"],
			images["elf_f_idle_anim_f2.png"],
			images["elf_f_idle_anim_f3.png"],
		},
		Config: cfgs["elf_f_idle_anim_f0.png"],
	}
	sprites["elf_f_run"] = Frames{
		Frames: []image.Image{
			images["elf_f_run_anim_f0.png"],
			images["elf_f_run_anim_f1.png"],
			images["elf_f_run_anim_f2.png"],
			images["elf_f_run_anim_f3.png"],
		},
		Config: cfgs["elf_f_run_anim_f0.png"],
	}
	sprites["floor_1"] = Frames{
		Frames: []image.Image{images["floor_1.png"]},
		Config: cfgs["floor_1.png"],
	}
	sprites["floor_2"] = Frames{
		Frames: []image.Image{images["floor_2.png"]},
		Config: cfgs["floor_2.png"],
	}
	sprites["floor_3"] = Frames{
		Frames: []image.Image{images["floor_3.png"]},
		Config: cfgs["floor_3.png"],
	}
	sprites["floor_4"] = Frames{
		Frames: []image.Image{images["floor_4.png"]},
		Config: cfgs["floor_4.png"],
	}
	sprites["floor_5"] = Frames{
		Frames: []image.Image{images["floor_5.png"]},
		Config: cfgs["floor_5.png"],
	}
	sprites["floor_6"] = Frames{
		Frames: []image.Image{images["floor_6.png"]},
		Config: cfgs["floor_6.png"],
	}
	sprites["floor_7"] = Frames{
		Frames: []image.Image{images["floor_7.png"]},
		Config: cfgs["floor_7.png"],
	}
	sprites["floor_8"] = Frames{
		Frames: []image.Image{images["floor_8.png"]},
		Config: cfgs["floor_8.png"],
	}
	sprites["wall_side_front_left"] = Frames{
		Frames: []image.Image{images["wall_side_front_left.png"]},
		Config: cfgs["wall_side_front_left.png"],
	}
	sprites["wall_side_front_right"] = Frames{
		Frames: []image.Image{images["wall_side_front_right.png"]},
		Config: cfgs["wall_side_front_right.png"],
	}
	return sprites, nil
}
func LoadLevel() [][]string {
	a := "floor_1"
	b := "floor_2"
	c := "floor_3"
	d := "floor_4"
	e := "wall_side_front_left"
	f := "wall_side_front_right"

	level := [][]string{
		[]string{e, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, f},
		[]string{e, a, b, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, f},
		[]string{e, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, f},
		[]string{e, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, c, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, f},
		[]string{e, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, f},
		[]string{e, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, f},
		[]string{e, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, f},
		[]string{e, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, f},
		[]string{e, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, f},
		[]string{e, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, f},
		[]string{e, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, f},
		[]string{e, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, f},
		[]string{e, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, f},
		[]string{e, a, a, a, a, a, c, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, f},
		[]string{e, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a},
		[]string{e, d, d, d, d, d, a, a, a, a, a, a, a, d, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a},
		[]string{e, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a},
		[]string{e, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a},
		[]string{e, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a},
		[]string{e, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a},
		[]string{e, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a},
		[]string{e, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a},
		[]string{e, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a},
		[]string{e, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a},
		[]string{e, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a},
		[]string{e, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a},
		[]string{e, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a},
		[]string{e, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a},
		[]string{e, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a},
		[]string{e, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a},
		[]string{e, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a},
		[]string{e, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a},
		[]string{e, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a},
		[]string{e, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a, a},
	}

	return level
}
