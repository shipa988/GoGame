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

	"github.com/satori/go.uuid"
	_ "github.com/shipa988/ebitentest/statik" // TODO: Replace with the absolute import path
	"github.com/shipa988/go-tiled"
	"github.com/shipa988/go-tiled/render"
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

var myUnitId uuid.UUID

//var unit *Unit

//var sprite *Sprite

var units Units

var lastKey e.Key
var prevKey e.Key

type Direction int

type Sprite struct {
	op     *e.DrawImageOptions
	Frames []*e.Image
	Frame  int
	X      float64
	Y      float64
	Side   Direction
	Config image.Config
}

type Unit struct {
	sprite    Sprite
	Id        uuid.UUID
	X         float64
	Y         float64
	Frame     int32
	Skin      string
	Action    string
	EAction   EvenType
	Speed     float64
	Direction Direction
	Side      Direction
}
type Frames struct {
	Frames []*e.Image
	render.Coll
	image.Config
}

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

type EvenType int

const (
	idle EvenType = iota
	move
	jump
)

type Event struct {
	idunit    uuid.UUID
	direction Direction
	etype     EvenType
}

var direction Direction
var etype EvenType
var isEvent bool

var sprites []*Sprite
var rwmx *sync.RWMutex
var eventChan chan Event

type UnitSprite struct {
	sprite *Sprite
	unit   *Unit
}

type Units struct {
	unitsprites map[uuid.UUID]*UnitSprite
	mx          sync.RWMutex
}

func NewUnits() Units {
	var u = Units{}
	u.unitsprites = make(map[uuid.UUID]*UnitSprite)
	return u
}

func (u *Units) Get(id uuid.UUID) *UnitSprite {
	//u.mx.RLock()
	s, ok := u.unitsprites[id]
	//u.mx.RUnlock()
	if ok {
		return s
	}
	return nil
}

func (u *Units) GetUnit(id uuid.UUID) *Unit {
	//u.mx.RLock()
	s, ok := u.unitsprites[id]
	//u.mx.RUnlock()
	if ok {
		return s.unit
	}
	return nil
}

func (u *Units) GetSprite(id uuid.UUID) *Sprite {
	//u.mx.RLock()
	s, ok := u.unitsprites[id]
	//u.mx.RUnlock()
	if ok {
		return s.sprite
	}
	return nil
}

func (u *Units) HandleEvent(event Event) {
	u.mx.Lock()
	player := units.GetUnit(event.idunit)
	switch event.etype {
	case move:
		player.Side = event.direction
		player.Direction = event.direction
		player.Action = UnitActionMove

	case idle:
		player.Action = UnitActionIdle
	}
	u.mx.Unlock()

}

func (u *Units) Add(uuid uuid.UUID, sp *Sprite, un *Unit) {
	u.mx.Lock()
	defer u.mx.Unlock()
	u.unitsprites[uuid] = &UnitSprite{
		sprite: sp,
		unit:   un,
	}
	rwmx.Lock()
	sprites = append(sprites, sp)
	rwmx.Unlock()
}

func (u *Units) Update() {
	u.mx.Lock()
	rwmx.Lock()
	//defer u.mx.Unlock()
	var ismove bool
	for _, unitsprite := range units.unitsprites {
		player := unitsprite.unit
		sprite := unitsprite.sprite
		frame := frames[player.Skin+"_"+player.Action]

		if player.Action == UnitActionMove {
			ismove = true
			player.Action = UnitActionMove
			switch player.Direction {
			case Direction_right:
				c, ok := level.collisionX[int(player.X+1)+frame.Width]
				if ok {
					if SearchInt(c, int(player.Y)+frame.Height) { //found coordinate
						ismove = false
						break
					}
				}
				if ismove {
					sprite.Side = player.Direction
					player.X += player.Speed
					sprite.X += player.Speed
				}
			case Direction_left:
				c, ok := level.collisionX[int(player.X-1)]
				if ok {
					if SearchInt(c, int(player.Y)+frame.Height) { //found coordinate
						ismove = false
						break
					}
				}
				if ismove {
					sprite.Side = player.Direction
					sprite.X -= player.Speed
					player.X -= player.Speed
				}
			case Direction_up:
				c, ok := level.collisionY[int(player.Y-1)+frame.Height]
				if ok {
					for x := int(player.X); x < int(player.X)+frame.Width; x++ {
						if SearchInt(c, x) { //found coordinate Значит есть пересечение по оси х
							ismove = false
							break
						}
					}
				}
				if ismove {
					sprite.Y -= player.Speed
					player.Y -= player.Speed
				}

			case Direction_down:
				c, ok := level.collisionY[int(player.Y+1)+frame.Height]
				if ok {
					for x := int(player.X); x < int(player.X)+frame.Width; x++ {
						if SearchInt(c, x) { //found coordinate
							ismove = false
							break
						}
					}

				}
				if ismove {
					sprite.Y += player.Speed
					player.Y += player.Speed
				}
			}
		}

		sprite.Frames = frames[player.Skin+"_"+player.Action].Frames

	}
	rwmx.Unlock()
	u.mx.Unlock()
	//}
	//	u.unitsprites[event.idunit].Frames = frames[unit.Skin+"_"+unit.Action].Frames
}

func Update(screen *e.Image) error {
	units.mx.RLock()
	player := units.GetUnit(myUnitId)

	handleKeyboard(player, eventChan)
	units.mx.RUnlock()
	if e.IsDrawingSkipped() {
		return nil
	}
	units.mx.RLock()
	handleCamera(player, screen)
	units.mx.RUnlock()
	rwmx.RLock()
	frame++
	sort.Slice(sprites, func(i, j int) bool {
		depth1 := sprites[i].Y + float64(sprites[i].Config.Height)
		depth2 := sprites[j].Y + float64(sprites[j].Config.Height)
		return depth1 < depth2
	})

	for _, sprite := range sprites {
		op := &e.DrawImageOptions{}
		op.GeoM.Reset()
		if sprite.Side == Direction_left {
			op.GeoM.Scale(-1, 1)
			op.GeoM.Translate(float64(sprite.Config.Width)*1.1-1, 0)
		}
		op.GeoM.Translate(sprite.X-camera.X, sprite.Y-camera.Y)
		err := screen.DrawImage(sprite.Frames[(frame/7+sprite.Frame)%len(sprite.Frames)], op)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	rwmx.RUnlock()
	//s := units.GetSprite(myUnitId)
	//log.Println(unit.Action)
	//ebitenutil.DebugPrint(screen, fmt.Sprintf("fps %0.2f U.x: %0.2f U.y: %0.2f population %v", e.CurrentFPS(), player.X, player.Y, len(units.unitsprites)))
	return nil
}

func init() {
	config = &Config{
		title:  "Another Hero",
		width:  420,
		height: 420,
		scale:  2,
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
	eventChan = make(chan Event, 0)
	done := make(chan struct{})
	wg := &sync.WaitGroup{}
	rwmx = &sync.RWMutex{}
	var err error
	frames, err = LoadResources()
	if err != nil {
		log.Fatal(err)
	}
	level, err = prepareLevel()

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	skins := []string{"chort"}
	myUnitId = uuid.Must(uuid.NewV4(), err)
	unit := &Unit{
		Id:        myUnitId,
		X:         200, //rnd.Float64()*float64(config.width-config.width/16) + 10,
		Y:         200, //rnd.Float64()*float64(config.height-config.height/16) + 10,
		Frame:     int32(rnd.Intn(4)),
		Skin:      skins[rnd.Intn(len(skins))],
		Action:    "idle",
		Speed:     1,
		Direction: Direction_right,
		Side:      Direction_right,
	}
	sprite := &Sprite{
		Frames: frames[unit.Skin+"_"+unit.Action].Frames,
		op:     &e.DrawImageOptions{},
		Frame:  int(unit.Frame),
		X:      unit.X - 1,
		Y:      unit.Y,
		Side:   unit.Side,
		Config: frames[unit.Skin+"_"+unit.Action].Config,
	}
	camera = &Camera{
		X:       unit.X,
		Y:       unit.Y,
		Padding: 30,
	}
	units = NewUnits()
	sprites = append(sprites, level.objects...)

	units.Add(myUnitId, sprite, unit)
	wg.Add(1)
	go func(donech chan struct{}) {
		event := Event{}
		defer wg.Done()
		for {
			select {
			case event = <-eventChan:
				units.HandleEvent(event)
			case <-donech:
				return
			}
		}
	}(done)
	wg.Add(1)
	go func() {
		defer wg.Done()
		rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
		skins := []string{"big_demon", "big_zombie", "goblin", "elf_f"}
		for {
			rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
			time.Sleep(time.Millisecond * 100)
			otherUnitId := uuid.Must(uuid.NewV4(), err)
			unit := &Unit{
				Id:        otherUnitId,
				X:         float64(rnd.Intn(level.levelImage.Bounds().Max.X)) / 2, //(rnd.Float64()*float64(level.levelImage.Bounds().Max.X-20)/float64(rnd.Intn(2)))+10,
				Y:         float64(rnd.Intn(level.levelImage.Bounds().Max.Y)) / 2, //(rnd.Float64()*float64(level.levelImage.Bounds().Max.Y-20)/float64(rnd.Intn(2)))+10,
				Frame:     int32(rnd.Intn(4)),
				Skin:      skins[rnd.Intn(len(skins))],
				Action:    "idle",
				Speed:     1,
				Direction: Direction_right,
				Side:      Direction_right,
			}
			sprite = &Sprite{
				Frames: frames[unit.Skin+"_"+unit.Action].Frames,
				op:     &e.DrawImageOptions{},
				Frame:  int(unit.Frame),
				X:      unit.X - 1,
				Y:      unit.Y,
				Side:   unit.Side,
				Config: frames[unit.Skin+"_"+unit.Action].Config,
			}
			rwmx.Lock()
			sprites = append(sprites, sprite)
			rwmx.Unlock()
			units.Add(otherUnitId, sprite, unit)
			wg.Add(1)
			go func(uid uuid.UUID, echan chan Event) {
				defer wg.Done()
				rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
				ticker := time.NewTicker(time.Second * 2)
				for {
					select {
					case <-ticker.C:
						event := Event{
							idunit:    uid,
							etype:     move,
							direction: Direction(rnd.Intn(4)),
						}
						echan <- event
					}
				}
			}(otherUnitId, eventChan)
		}

	}()
	wg.Add(1)
	go func(donech chan struct{}) {
		defer wg.Done()
		ticker := time.NewTicker(time.Second / 60)
		for {
			select {
			case <-ticker.C:
				units.Update()
			case <-donech:
				return
			}
		}
	}(done)
	if err := e.Run(Update, config.width, config.height, config.scale, config.title); err != nil {
		log.Fatal(err)
	}
	fmt.Println("все")
	close(done)
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
		for name, layer := range l {
			for k, v := range layer.Coll.ColmapX {
				level.collisionX[k] = append(level.collisionX[k], v...)
			}
			for k, v := range layer.Coll.ColmapY {
				level.collisionY[k] = append(level.collisionY[k], v...)
			}

			if strings.Index(name, "objects_") >= 0 {
				for _, object := range layer.TileObjects {
					img, err := e.NewImageFromImage(object.TileImage, e.FilterDefault)
					if err != nil {
						log.Println(err)

					}
					level.objects = append(level.objects, &Sprite{
						Frames: []*e.Image{img},
						op:     &e.DrawImageOptions{},
						Frame:  0,
						X:      float64(object.TilePos.Min.X),
						Y:      float64(object.TilePos.Min.Y),
						Side:   3,
						Config: image.Config{
							ColorModel: object.TileImage.ColorModel(),
							Width:      object.TileImage.Bounds().Max.X,
							Height:     object.TileImage.Bounds().Max.Y,
						},
					})

				}

			}
		}

		for k, v := range level.collisionX {
			m := unique(v)
			sort.Ints(m)
			level.collisionX[k] = m
		}
		for k, v := range level.collisionY {
			m := unique(v)
			sort.Ints(m)
			level.collisionY[k] = m
		}
		return &level, nil
	}

	return nil, errors.New("can't load map")
}

func handleCamera(player *Unit, screen *e.Image) {
	if camera == nil {
		return
	}
	frame := frames[player.Skin+"_"+player.Action]
	camera.X = player.X - float64(config.width-frame.Config.Width)/2
	camera.Y = player.Y - float64(config.height-frame.Config.Height)/2

	op := &e.DrawImageOptions{}
	op.GeoM.Translate(-camera.X, -camera.Y)
	screen.DrawImage(level.levelImage, op)
}

func handleKeyboard(player *Unit, echan chan Event) {
	isEvent = false
	//frame := frames[player.Skin+"_"+player.Action]
	etype = idle
	lastKey = -1
	/*if e.IsKeyPressed(prevKey)==true{
		return
	}*/
	//e.IsKeyPressed(e.MouseButtonLeft)
	if e.IsKeyPressed(e.KeyA) || e.IsKeyPressed(e.KeyLeft) {
		isEvent = true
		direction = Direction_left
		etype = move
		if lastKey != e.KeyLeft {
			lastKey = e.KeyLeft
		}
	}

	if e.IsKeyPressed(e.KeyD) || e.IsKeyPressed(e.KeyRight) {

		isEvent = true
		direction = Direction_right
		etype = move
		if lastKey != e.KeyRight {
			lastKey = e.KeyRight
		}
	}

	if e.IsKeyPressed(e.KeyW) || e.IsKeyPressed(e.KeyUp) {

		isEvent = true
		etype = move
		direction = Direction_up

		if lastKey != e.KeyUp {
			lastKey = e.KeyUp
		}
	}

	if e.IsKeyPressed(e.KeyS) || e.IsKeyPressed(e.KeyDown) {

		isEvent = true
		direction = Direction_down
		etype = move
		if lastKey != e.KeyDown {
			lastKey = e.KeyDown
		}
	}
	if (isEvent && prevKey != lastKey) || player.Action == UnitActionMove {
		{
			event := Event{
				idunit:    player.Id,
				etype:     etype,
				direction: direction,
			}
			echan <- event
			prevKey = lastKey
		}
	}

	//if event.etype!=move {

	//	if unit.Action != UnitActionIdle {
	//		unit.Action = UnitActionIdle
	//	}
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
			}
		//}
		//	sprite.Frames = frames[unit.Skin+"_"+unit.Action].Frames*/
	//return unit
	//prevKey = lastKey
	//}
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
		collision, err := renderer.RenderLayer(0)
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
		if err != nil || cfg.Width == 0 {
			cfg.ColorModel = img.ColorModel()
			cfg.Width = int(0.8 * float64(img.Bounds().Max.X))
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
	sprites["goblin_idle"] = Frames{
		Frames: []*e.Image{
			images["goblin_idle_anim_f0.png"],
			images["goblin_idle_anim_f1.png"],
			images["goblin_idle_anim_f2.png"],
			images["goblin_idle_anim_f3.png"],
		},
		Config: cfgs["goblin_idle_anim_f0.png"],
	}
	sprites["goblin_run"] = Frames{
		Frames: []*e.Image{
			images["goblin_run_anim_f0.png"],
			images["goblin_run_anim_f1.png"],
			images["goblin_run_anim_f2.png"],
			images["goblin_run_anim_f3.png"],
		},
		Config: cfgs["goblin_run_anim_f0.png"],
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
	sprites["chort_idle"] = Frames{
		Frames: []*e.Image{
			images["chort_idle_anim_f0.png"],
			images["chort_idle_anim_f1.png"],
			images["chort_idle_anim_f2.png"],
			images["chort_idle_anim_f3.png"],
		},
		Config: cfgs["chort_idle_anim_f0.png"],
	}
	sprites["chort_run"] = Frames{
		Frames: []*e.Image{
			images["chort_run_anim_f0.png"],
			images["chort_run_anim_f1.png"],
			images["chort_run_anim_f2.png"],
			images["chort_run_anim_f3.png"],
		},
		Config: cfgs["chort_run_anim_f0.png"],
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
