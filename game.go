//go:generate statik -src=resources -include=*.png,*.tmx
package main

import (
	"errors"
	"sort"
	"strconv"
	"sync"

	//"bytes"
	"flag"
	"fmt"
	e "github.com/hajimehoshi/ebiten"
	"strings"

	//"github.com/markbates/pkger"
	"github.com/rakyll/statik/fs"
	//"github.com/gobuffalo/packr/v2"
	//"github.com/markbates/pkger"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/lafriks/go-tiled"
	"github.com/lafriks/go-tiled/render"
	_ "github.com/shipa988/ebitentest/statik" // TODO: Replace with the absolute import path
	"image"
	"image/png"
	"log"
	"math/rand"
	"os"
	"runtime/pprof"
	"time"
)

const (
	UnitActionIdle = "idle"
	UnitActionMove = "run"
)

type Config struct {
	title  string
	width  int
	height int
	scale  float64
}
type Camera struct {
	X       float64
	Y       float64
	Padding float64
}

var config *Config
var camera *Camera
var level *Level
var frames map[string]Frames
var frame int
var unit *Unit
var sprite *Sprite

type Direction int

type Sprite struct {
	op *e.DrawImageOptions
	Frames []*e.Image
	Frame  int
	X      float64
	Y      float64
	Side   Direction
	Config image.Config
}
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
type Frames struct {
	Frames []*e.Image
	render.Coll
	image.Config
}
/*type SpriteImage struct {
	sprite Sprite
	image *e.Image
}
*/
type Level struct {
	levelImage *e.Image
	collisionX map[int][]int
	collisionY map[int][]int
	objects    []*Sprite

}

const (
	Direction_left  Direction = 0
	Direction_right Direction = iota
	Direction_up
	Direction_down
)

var sprites []*Sprite

func Update(screen *e.Image) error {
	handleKeyboard()
	if e.IsDrawingSkipped() {
		return nil
	}
	handleCamera(screen)

	frame++
	sort.Slice(sprites, func(i, j int) bool {
		depth1 := sprites[i].Y + float64(sprites[i].Config.Height)
		depth2 := sprites[j].Y + float64(sprites[j].Config.Height)
		return depth1 < depth2
	})
//	op := &e.DrawImageOptions{}
/*	for _,obj:=range level.objects{

		op.GeoM.Reset()
		op.GeoM.Translate(obj.sprite.X-camera.X, obj.sprite.Y-camera.Y)
		err:= screen.DrawImage(obj.image, op)
		if err != nil {
			log.Println(err)
			return err
		}
	}*/
	for _, sprite := range sprites {
		op := &e.DrawImageOptions{}
		op.GeoM.Reset()
		if sprite.Side == Direction_left {
			op.GeoM.Scale(-1, 1)
			op.GeoM.Translate(float64(sprite.Config.Width), 0)
		}
		op.GeoM.Translate(sprite.X-camera.X, sprite.Y-camera.Y)

		err:= screen.DrawImage(sprite.Frames[(frame/7+sprite.Frame)%len(sprite.Frames)], op)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	ebitenutil.DebugPrint(screen, fmt.Sprintf("fps %0.2f U.x: %0.2f U.y: %0.2f cam.x: %0.2f cam.y: %0.2f", e.CurrentFPS(), unit.X, unit.Y, camera.X, camera.Y))
	return nil
}

func init() {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	skins := []string{"big_demon", "big_zombie", "elf_f"}
	config = &Config{
		title:  "Another Hero",
		width:  420,
		height: 420,
		scale:  2,
	}
	unit = &Unit{
		Id:        1,
		X:         80, //rnd.Float64()*float64(config.width-config.width/16) + 10,
		Y:         128, //rnd.Float64()*float64(config.height-config.height/16) + 10,
		Frame:     int32(rnd.Intn(4)),
		Skin:      skins[rnd.Intn(len(skins))],
		Action:    "idle",
		Speed:     1,
		Direction: Direction_right,
		Side:      Direction_right,
	}
}

//-cpuprofile=havlak1.prof
//go tool pprof havlak1 havlak1.prof
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var memprofile = flag.String("memprofile", "", "write memory profile to this file")

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
	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.WriteHeapProfile(f)
		f.Close()
		return
	}

	done:=make(chan struct{})
	defer close(done)
	wg:=&sync.WaitGroup{}
	var err error
	frames, err = LoadResources()
	if err != nil {
		log.Fatal(err)
	}

	level, err = prepareLevel()
	camera = &Camera{
		X:       unit.X,
		Y:       unit.Y,
		Padding: 30,
	}
	op := &e.DrawImageOptions{}
	sprites = append(sprites, level.objects...)
	sprite=&Sprite{
		Frames: frames[unit.Skin+"_"+unit.Action].Frames,
		op:op,
		Frame:  int(unit.Frame),
		X:      unit.X,
		Y:      unit.Y,
		Side:   unit.Side,
		Config: frames[unit.Skin+"_"+unit.Action].Config,
	}
	sprites = append(sprites, sprite)
	wg.Add(1)
	go func(donech chan struct{}) {
		defer 	wg.Done()
		ticker:=time.NewTicker(time.Second/60)
		for  {
			select {
			case <-ticker.C:
				//spr.
			case<-done:
				return


			}
		}
	}(done)
	if err := e.Run(Update, config.width, config.height, config.scale, config.title); err != nil {
		log.Fatal(err)
	}
	done<- struct{}{}
	wg.Wait()
}
func prepareLevel() (*Level, error) {
	l, err := LoadMapTMX(8)
	if err != nil {
		return nil, err
	}
	//all,ok:=l["background"]
	//all,ok:=l["co"]
	all, ok := l["all_layers"]
	//obj, objok := l["testwalls"]

	if ok {
		op := &e.DrawImageOptions{}
		op.GeoM.Translate(float64(all.Width), float64(all.Height))
		img := all.Frames[0]
		level := Level{
			levelImage: img,
			collisionX: all.Coll.ColmapX,
			collisionY: all.Coll.ColmapY,
		}
		for name,layer:=range l {
			for k,v:= range layer.Coll.ColmapX {
				level.collisionX[k] = append(level.collisionX[k], v...)
			}
			for k,v:= range layer.Coll.ColmapY {
				level.collisionY[k] = append(level.collisionY[k], v...)
			}

			if strings.Index(name,"objects_")>=0{
				for _, object := range layer.TileObjects {
					img, err := e.NewImageFromImage(object.TileImage, e.FilterDefault)
					if err != nil {
						log.Println(err)

					}
					level.objects = append(level.objects,  &Sprite{
						Frames: []*e.Image{img},
						op: &e.DrawImageOptions{},
						Frame:  0,
						X:      float64(object.TilePos.Min.X),
						Y:      float64(object.TilePos.Min.Y),
						Side:   3,
						Config: image.Config{
							ColorModel: object.TileImage.ColorModel(),
							Width:      object.TileImage.Bounds().Max.X,
							Height:     object.TileImage.Bounds().Max.Y,
						},

					} )


				}

			}
		}

		for k,v:=range level.collisionX  {
			m := unique(v)
			sort.Ints(m)
			level.collisionX[k] = m
		}
		for k,v:=range level.collisionY  {
			m := unique(v)
			sort.Ints(m)
			level.collisionY[k] = m
		}
		return &level, nil
	}

	return nil, errors.New("can't load map")
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
	screen.DrawImage(level.levelImage, op)
}

func SearchInt(a []int, x int) bool {
	if x < a[0] || x > a[len(a)-1] {
		return false
	}

	for _, y := range a {
		if y == x {
			return true
		}
		if y > x {
			return false
		}
	}
	return false
}

func handleKeyboard() {
	//event := &game.Event{}
	player := unit

	frame := frames[player.Skin+"_"+player.Action]

	var ismove bool
	if e.IsKeyPressed(e.KeyA) || e.IsKeyPressed(e.KeyLeft) {

		ismove = true
		c, ok := level.collisionX[int(unit.X-1)]
		if ok {
			//for y := int(unit.Y); y < int(unit.Y)+frame.Height; y++ {
			//pos:=//pos!=0 and pos!=len(c)-значит найдено значение внутри массива
			if SearchInt(c, int(unit.Y)+frame.Height) { //found coordinate
				ismove = false
				//break
			}
			//}

		}
		sprite.Side= Direction_left

		unit.Direction = Direction_left
		unit.Side = Direction_left
		if ismove {
			sprite.X-= unit.Speed
			unit.X -= unit.Speed
		}

	}

	if e.IsKeyPressed(e.KeyD) || e.IsKeyPressed(e.KeyRight) {

		ismove = true
		c, ok := level.collisionX[int(unit.X+1)+frame.Width]
		if ok {
			//for y := int(unit.Y); y < int(unit.Y)+frame.Height; y++ {
			//pos:=//pos!=0 and pos!=len(c)-значит найдено значение внутри массива
			if SearchInt(c, int(unit.Y)+frame.Height) { //found coordinate
				ismove = false
				//	break
				//	}
			}

		}
		sprite.Side = Direction_right
		unit.Direction = Direction_right
		unit.Side = Direction_right
		if ismove {
			unit.X += unit.Speed
			sprite.X+= unit.Speed
		}

	}

	if e.IsKeyPressed(e.KeyW) || e.IsKeyPressed(e.KeyUp) {

		ismove = true
		c, ok := level.collisionY[int(unit.Y-1)+frame.Height]
		if ok {
			for x := int(unit.X); x < int(unit.X)+frame.Width; x++ {
				//pos:=//pos!=0 and pos!=len(c)-значит найдено значение внутри массива
				if SearchInt(c, x) { //found coordinate
					ismove = false
					break
				}
			}

		}

		unit.Direction = Direction_up
		unit.Side = unit.Side
		if ismove {
			unit.Y -= unit.Speed
			sprite.Y -= unit.Speed
	}

	}

	if e.IsKeyPressed(e.KeyS) || e.IsKeyPressed(e.KeyDown) {

		ismove = true
		c, ok := level.collisionY[int(unit.Y+1)+frame.Height]
		if ok {
			for x := int(unit.X); x < int(unit.X)+frame.Width; x++ {
				if SearchInt(c, x) { //found coordinate
					ismove = false
					break
				}
			}

		}
		sprite.Side=Direction_down
		unit.Direction = Direction_down
		unit.Side = unit.Side
		if ismove {
			unit.Y += unit.Speed
			sprite.Y += unit.Speed
		}

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
	sprite.Frames= frames[unit.Skin+"_"+unit.Action].Frames
	//return unit
	//prevKey = lastKey
}
func unique(intSlice []int) []int {
	keys := make(map[int]bool)
	list := []int{}
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func LoadMapTMX(mapId int) (map[string]Frames, error) {
	layers := map[string]Frames{}
	statikFS, err := fs.New()
	if err != nil {
		log.Fatal(err)
	}
	fs.Walk(statikFS, "/map/tmx", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		fmt.Println(path)

		if info.IsDir() {
			return nil
		}
		// Access individual files by their paths.
		if !strings.Contains(path, strconv.Itoa(mapId)+`_map.tmx`) {
			return nil
		}
		l := tiled.Loader{FileSystem: statikFS}
		gameMap, err := l.LoadFromFile(path)

		if err != nil {
			fmt.Println("Error parsing map")
			return nil
		}

		//fmt.Println(gameMap)

		// You can also render the map to an in-memory image for direct
		// use with the default Renderer, or by making your own.
		renderer, err := render.NewRenderer(gameMap)
		if err != nil {
			fmt.Println("Error parsing map")
			return nil
		}

		for i, layer := range gameMap.Layers {
			name := layer.Name
			collision, err := renderer.RenderLayer(i)

			if err != nil {
				return err
			}
			for k, v := range collision.ColmapX {
				m := unique(v)
				sort.Ints(m)
				collision.ColmapX[k] = m
			}
			for k, v := range collision.ColmapY {
				m := unique(v)
				sort.Ints(m)
				collision.ColmapY[k] = m
			}

			img, _ := e.NewImageFromImage(renderer.Result, e.FilterDefault)
			layers[name] = Frames{
				Frames: []*e.Image{img},
				Coll:   collision,
				Config: image.Config{
					ColorModel: img.ColorModel(),
					Width:      img.Bounds().Max.X,
					Height:     img.Bounds().Max.Y,
				},
			}
		}
		collision, err := renderer.RenderVisibleLayers()
		if err != nil {
			return err
		}
		for k, v := range collision.ColmapX {
			m := unique(v)
			sort.Ints(m)
			collision.ColmapX[k] = m
		}
		for k, v := range collision.ColmapY {
			m := unique(v)
			sort.Ints(m)
			collision.ColmapY[k] = m
		}
		img, _ := e.NewImageFromImage(renderer.Result, e.FilterDefault)
		layers["all_layers"] = Frames{
			Frames: []*e.Image{img},
			Coll:   collision,
			Config: image.Config{
				ColorModel: img.ColorModel(),
				Width:      img.Bounds().Max.X,
				Height:     img.Bounds().Max.Y,
			},
		}

		// Get a reference to the Renderer's output, an image.NRGBA struct.

		return nil
	})
	return layers, nil
}

func LoadResources() (map[string]Frames, error) {
	images := map[string]*e.Image{}
	cfgs := map[string]image.Config{}
	sprites := map[string]Frames{}
	statikFS, err := fs.New()
	if err != nil {
		log.Fatal(err)
	}

	fs.Walk(statikFS, "/sprites", func(path string, info os.FileInfo, err error) error {
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
		if cfg.Width == 0 {
			cfg.ColorModel = img.ColorModel()
			cfg.Width = int(0.7 * float64(img.Bounds().Max.X))
			cfg.Height = img.Bounds().Max.Y
		}
		eimg, err := e.NewImageFromImage(img, e.FilterDefault)
		if err != nil {
			log.Println(err)
			return nil
		}
		images[info.Name()] = eimg
		cfgs[info.Name()] = cfg
		f.Close()
		return nil
	})

	sprites["big_demon_idle"] = Frames{
		Frames: []*e.Image{
			images["big_demon_idle_anim_f0.png"],
			images["big_demon_idle_anim_f1.png"],
			images["big_demon_idle_anim_f2.png"],
			images["big_demon_idle_anim_f3.png"],
		},
		Config: cfgs["big_demon_idle_anim_f0.png"],
	}
	sprites["big_demon_run"] = Frames{
		Frames: []*e.Image{
			images["big_demon_run_anim_f0.png"],
			images["big_demon_run_anim_f1.png"],
			images["big_demon_run_anim_f2.png"],
			images["big_demon_run_anim_f3.png"],
		},
		Config: cfgs["big_demon_run_anim_f0.png"],
	}

	sprites["big_zombie_idle"] = Frames{
		Frames: []*e.Image{
			images["big_zombie_idle_anim_f0.png"],
			images["big_zombie_idle_anim_f1.png"],
			images["big_zombie_idle_anim_f2.png"],
			images["big_zombie_idle_anim_f3.png"],
		},
		Config: cfgs["big_zombie_idle_anim_f0.png"],
	}
	sprites["big_zombie_run"] = Frames{
		Frames: []*e.Image{
			images["big_zombie_run_anim_f0.png"],
			images["big_zombie_run_anim_f1.png"],
			images["big_zombie_run_anim_f2.png"],
			images["big_zombie_run_anim_f3.png"],
		},
		Config: cfgs["big_zombie_run_anim_f0.png"],
	}

	sprites["elf_f_idle"] = Frames{
		Frames: []*e.Image{
			images["elf_f_idle_anim_f0.png"],
			images["elf_f_idle_anim_f1.png"],
			images["elf_f_idle_anim_f2.png"],
			images["elf_f_idle_anim_f3.png"],
		},
		Config: cfgs["elf_f_idle_anim_f0.png"],
	}
	sprites["elf_f_run"] = Frames{
		Frames: []*e.Image{
			images["elf_f_run_anim_f0.png"],
			images["elf_f_run_anim_f1.png"],
			images["elf_f_run_anim_f2.png"],
			images["elf_f_run_anim_f3.png"],
		},
		Config: cfgs["elf_f_run_anim_f0.png"],
	}
	sprites["floor_1"] = Frames{
		Frames: []*e.Image{images["floor_1.png"]},
		Config: cfgs["floor_1.png"],
	}
	sprites["floor_2"] = Frames{
		Frames: []*e.Image{images["floor_2.png"]},
		Config: cfgs["floor_2.png"],
	}
	sprites["floor_3"] = Frames{
		Frames: []*e.Image{images["floor_3.png"]},
		Config: cfgs["floor_3.png"],
	}
	sprites["floor_4"] = Frames{
		Frames: []*e.Image{images["floor_4.png"]},
		Config: cfgs["floor_4.png"],
	}
	sprites["floor_5"] = Frames{
		Frames: []*e.Image{images["floor_5.png"]},
		Config: cfgs["floor_5.png"],
	}
	sprites["floor_6"] = Frames{
		Frames: []*e.Image{images["floor_6.png"]},
		Config: cfgs["floor_6.png"],
	}
	sprites["floor_7"] = Frames{
		Frames: []*e.Image{images["floor_7.png"]},
		Config: cfgs["floor_7.png"],
	}
	sprites["floor_8"] = Frames{
		Frames: []*e.Image{images["floor_8.png"]},
		Config: cfgs["floor_8.png"],
	}
	sprites["wall_side_front_left"] = Frames{
		Frames: []*e.Image{images["wall_side_front_left.png"]},
		Config: cfgs["wall_side_front_left.png"],
	}
	sprites["wall_side_front_right"] = Frames{
		Frames: []*e.Image{images["wall_side_front_right.png"]},
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
