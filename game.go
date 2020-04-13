package game

import (

	"log"
	"math/rand"
	"os"
	"time"

	//"bytes"
	"image"
	"image/png"
	e "github.com/hajimehoshi/ebiten"
	"github.com/markbates/pkger"

)

const (
	screenWidth  = 320
	screenHeight = 240

	frameOX     = 0
	frameOY     = 32
	frameWidth  = 32
	frameHeight = 32
	frameNum    = 8
)



var frames map[string]Frames
var frame int
var unit *Unit

func Update(screen *e.Image) error {
	frame++

	if e.IsDrawingSkipped() {
		return nil
	}

	op := &e.DrawImageOptions{}
	op.GeoM.Translate(-float64(frameWidth)/2, -float64(frameHeight)/2)
	op.GeoM.Translate(screenWidth/2, screenHeight/2)

	sprites = append(sprites, Sprite{
		Frames: frames[unit.Skin+"_"+unit.Action].Frames,
		Frame:  int(unit.Frame),
		X:      unit.X,
		Y:      unit.Y,
		//Side:   unit.Side,
		Config: frames[unit.Skin+"_"+unit.Action].Config,
	})
	for _,sprite:=range sprites  {
		img, err := e.NewImageFromImage(sprite.Frames[(frame/7+sprite.Frame)%4], e.FilterDefault)
		if err != nil {
			log.Println(err)
			return err
		}
		err= screen.DrawImage(img, op)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}
type Sprite struct {
	Frames []image.Image
	Frame  int
	X      float64
	Y      float64
	//Side   game.Direction
	Config image.Config
}

type Unit struct {
	Id int
	X float64
	Y float64
	Frame int32
	Skin string
	Action string
	Speed int


}
type Frames struct {
	Frames []image.Image
	image.Config
}
var sprites []Sprite
func LoadResources() (map[string]Frames, error) {
	images := map[string]image.Image{}
	cfgs := map[string]image.Config{}
	sprites := map[string]Frames{}

	pkger.Walk("./resources/sprites", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		f,err:=pkger.Open(info.Name())
		if err != nil {
			return err
		}
		img, err := png.Decode(f)
		if err != nil {
			return  err
		}
		cfg, err := png.DecodeConfig(f)

		images[info.Name()] = img
		cfgs[info.Name()] = cfg
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

	return sprites, nil
}

func init() {
	// Decode image from a byte slice instead of a file so that
	// this example works in any working directory.
	// If you want to use a file, there are some options:
	// 1) Use os.Open and pass the file to the image decoder.
	//    This is a very regular way, but doesn't work on browsers.
	// 2) Use ebitenutil.OpenFile and pass the file to the image decoder.
	//    This works even on browsers.
	// 3) Use ebitenutil.NewImageFromFile to create an ebiten.Image directly from a file.
	//    This also works on browsers.
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	skins := []string{"big_demon", "big_zombie", "elf_f"}
	unit = &Unit{
		Id:     1,
		X:      rnd.Float64()*300 + 10,
		Y:      rnd.Float64()*220 + 10,
		Frame:  int32(rnd.Intn(4)),
		Skin:   skins[rnd.Intn(len(skins))],
		Action: "idle",
		Speed:  1,
	}
	var err error
	frames, err =LoadResources()
	if err != nil {
		log.Fatal(err)
	}

}