package client

import (
	"flag"
	"fmt"
	e "github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"image"
	"log"
	"math/rand"
	"net/http"
	et "github.com/shipa988/ebitentest"
	"os"
	"runtime/pprof"
	"time"
)
type Frames struct {
	Frames []image.Image
	image.Config
}

type Camera struct {
	X       float64
	Y       float64
	Padding float64
}

var config *et.Config
var camera *Camera
var levelImage *e.Image
var frames map[string]Frames
var frame int

type Sprite struct {
	Frames []image.Image
	Frame  int
	X      float64
	Y      float64
	Side   et.Direction
	Config image.Config
}

func Update(screen *e.Image) error {
	handleKeyboard()
	if e.IsDrawingSkipped() {
		return nil
	}
	handleCamera(screen)

	sprites := []Sprite{}
	frame++
	op := &e.DrawImageOptions{}

	sprites = append(sprites, Sprite{
		Frames: frames[unit.Skin+"_"+unit.Action].Frames,
		Frame:  int(unit.Frame),
		X:      unit.X,
		Y:      unit.Y,
		Side:   unit.Side,
		Config: frames[unit.Skin+"_"+unit.Action].Config,
	})
	for _, sprite := range sprites {

		if sprite.Side == Direction_left {
			op.GeoM.Scale(-1, 1)
			op.GeoM.Translate(float64(sprite.Config.Width), 0)
		}

		op.GeoM.Translate(sprite.X-camera.X, sprite.Y-camera.Y)
		img, err := e.NewImageFromImage(sprite.Frames[(frame/7+sprite.Frame)%4], e.FilterDefault)
		if err != nil {
			log.Println(err)
			return err
		}

		err = screen.DrawImage(img, op)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	ebitenutil.DebugPrint(screen, fmt.Sprintf("U.x: %0.2f U.y: %0.2f cam.x: %0.2f cam.y: %0.2f", unit.X, unit.Y, camera.X, camera.Y))
	return nil
}

func init() {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	skins := []string{"big_demon", "big_zombie", "elf_f"}
	config = &Config{
		title:  "Another Hero",
		width:  720,
		height: 480,
		scale:  1,
	}
	unit = &Unit{
		Id:        1,
		X:         rnd.Float64()*float64(config.width-config.width/16) + 10,
		Y:         rnd.Float64()*float64(config.height-config.height/16) + 10,
		Frame:     int32(rnd.Intn(4)),
		Skin:      skins[rnd.Intn(len(skins))],
		Action:    "idle",
		Speed:     1,
		Direction: Direction_right,
		Side:      Direction_right,
	}
}

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	var err error
	frames, err = LoadResources()
	if err != nil {
		log.Fatal(err)
	}

	levelImage, err = prepareLevelImage()
	camera = &Camera{
		X:       unit.X,
		Y:       unit.Y,
		Padding: 30,
	}
	if err := e.Run(Update, config.width, config.height, config.scale, config.title); err != nil {
		log.Fatal(err)
	}
}


func handleCamera(screen *e.Image) {
	if camera == nil {
		return
	}

	player := unit
	frame := frames[player.Skin+"_"+player.Action]
	camera.X = player.X - float64(config.width-frame.Config.Width)/2
	camera.Y = player.Y - float64(config.height-frame.Config.Height)/2

	op := &e.DrawImageOptions{}
	op.GeoM.Translate(-camera.X, -camera.Y)
	screen.DrawImage(levelImage, op)
}

func handleKeyboard() {
	//event := &game.Event{}
	var ismove bool
	if e.IsKeyPressed(e.KeyA) || e.IsKeyPressed(e.KeyLeft) {
		ismove = true
		unit.Direction = Direction_left
		unit.Side = Direction_left
		unit.X -= unit.Speed
	}

	if e.IsKeyPressed(e.KeyD) || e.IsKeyPressed(e.KeyRight) {
		unit.Action = "run"
		ismove = true
		unit.Direction = Direction_right
		unit.Side = Direction_right
		unit.X += unit.Speed
	}

	if e.IsKeyPressed(e.KeyW) || e.IsKeyPressed(e.KeyUp) {
		ismove = true
		unit.Direction = Direction_up
		unit.Side = unit.Side
		unit.Y -= unit.Speed
	}

	if e.IsKeyPressed(e.KeyS) || e.IsKeyPressed(e.KeyDown) {
		ismove = true
		unit.Direction = Direction_down
		unit.Side = unit.Side

		unit.Y += unit.Speed
	}

	if ismove {
		unit.Action = UnitActionMove
		/*if prevKey != lastKey {
			message, err := proto.Marshal(event)
			if err != nil {
				log.Println(err)
				return
			}
			c.WriteMessage(websocket.BinaryMessage, message)
		}*/
	} else {

		if unit.Action != UnitActionIdle {
			unit.Action = UnitActionIdle
		}
		/*	event = &game.Event{
				Type: game.Event_type_idle,
				Data: &game.Event_Idle{
					&game.EventIdle{PlayerId: world.MyID},
				},
			}
			message, err := proto.Marshal(event)
			if err != nil {
				log.Println(err)
				return
			}
			c.WriteMessage(websocket.BinaryMessage, message)
			lastKey = -1
		}*/
	}

	//prevKey = lastKey
}
