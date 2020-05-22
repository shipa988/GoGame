//go:generate statik -src=resources -include=*.png,*.tmx
package main

import (
	"errors"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"runtime"

	//"github.com/hajimehoshi/ebiten/ebitenutil"
	"image/color"
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

const (
	idle EvenType = iota
	move
)
const (
	Direction_left  Direction = 0
	Direction_right Direction = iota
	Direction_up
	Direction_down
)

var config *Config

var camera *Camera

var level *Level

var unit *Unit
var myUnitId uuid.UUID
var sprite *Sprite

var frames map[string]Frames

var frame int

var units Units

var lastKey e.Key

var prevKey e.Key

var direction Direction
var etype EvenType

//var isEvent bool
var sprites []*Sprite
var eventChan chan Event

type Direction int

type Frames struct {
	images []*e.Image
	config ImageConfig
}
type TmxMap struct{
	image *e.Image
	config ImageConfig
	render.Coll
	objectGroups []*tiled.ObjectGroup
}

type Point struct {
	x,y float64
}

type Route struct {
	name string
	path []Point
}

type Level struct {
	levelImage *e.Image
	collisionX map[float64][]float64
	collisionY map[float64][]float64
	objects    []*Sprite
	animations []*Sprite
	heroroutes [] *Route
}

type Config struct {
	title  string
	width  float64
	height float64
	scale  float64
}

type Camera struct {
	X       float64
	Y       float64
	Padding float64
}

type EvenType int

type Event struct {
	idunit    uuid.UUID
	direction Direction
	etype     EvenType
}

type ImageConfig struct {
	ColorModel color.Model
	Width      float64
	Height     float64
}

type Sprite struct {
	op *e.DrawImageOptions
	//Frames map[EvenType][]*e.Image
	Frames []*e.Image
	Frame  int
	X      float64
	Y      float64
	Side   Direction
	Config ImageConfig
}

type Unit struct {
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

type UnitSprite struct {
	sprite *Sprite
	unit   *Unit
}
type MyMutex struct {
	mx sync.RWMutex
}

func (m *MyMutex) Lock() {
	m.mx.Lock()
	fmt.Println("Lock")
}
func (m *MyMutex) RLock() {
	m.mx.RLock()
	fmt.Println("RLock")
}
func (m *MyMutex) Unlock() {
	m.mx.Unlock()
	fmt.Println("Unlock")
}
func (m *MyMutex) RUnlock() {
	m.mx.RUnlock()
	fmt.Println("RUnlock")
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

func (u *Units) HandleEvent(eventChan chan Event, donech chan struct{}, wg *sync.WaitGroup) {
	wg.Add(1)
	func() {
		event := Event{}
		player := &Unit{}
		defer wg.Done()
		for {
			select {
			case event = <-eventChan:
				u.mx.Lock()
				player = units.GetUnit(event.idunit)
				switch event.etype {
				case move:
					player.Side = event.direction
					player.Direction = event.direction
					player.Action = UnitActionMove
				case idle:
					player.Action = UnitActionIdle
				}
				u.mx.Unlock()
			case <-donech:
				return
			}
		}
	}()
}

//Add adds new UnitSprite object to Units and sprite to sprites and sort it after then.
func (u *Units) Add(uuid uuid.UUID, sp *Sprite, un *Unit) {
	u.mx.Lock()

	u.unitsprites[uuid] = &UnitSprite{
		sprite: sp,
		unit:   un,
	}
	sprites = append(sprites, sp)
	sort.Slice(sprites, func(i, j int) bool {
		depth1 := sprites[i].Y + sprites[i].Config.Height
		depth2 := sprites[j].Y + sprites[j].Config.Height
		return depth1 < depth2
	})

	u.mx.Unlock()
}

func (u *Units) Update(donech chan struct{}, wg *sync.WaitGroup) {
	wg.Add(1)
	func() {
		defer wg.Done()
		var ismove bool
		sortsprites := func(i, j int) bool {
			depth1 := sprites[i].Y + sprites[i].Config.Height
			depth2 := sprites[j].Y + sprites[j].Config.Height
			return depth1 < depth2
		}
		ticker := time.NewTicker(time.Second / 60)
		for {
			select {
			case <-ticker.C:
				u.mx.Lock()
				for _, unitsprite := range u.unitsprites {
					player := unitsprite.unit
					sprite := unitsprite.sprite
					//sprite.Frames[]
					frame := frames[player.Skin+"_"+player.Action]
					sprite.Frames = frame.images
					if player.Action == UnitActionMove {
						ismove = true
						player.Action = UnitActionMove
						switch player.Direction {
						case Direction_right:
							c, ok := level.collisionX[(player.X+1)+frame.config.Width]
							if ok {
								if SearchFloat(c, (player.Y)+frame.config.Height) { //found coordinate
									ismove = false
									break
								}
							}
							if ismove {
								player.X += player.Speed
								sprite.Side = player.Direction
								sprite.X += player.Speed
							}
						case Direction_left:
							c, ok := level.collisionX[(player.X - 1)]
							if ok {
								if SearchFloat(c, (player.Y)+frame.config.Height) { //found coordinate
									ismove = false
									break
								}
							}
							if ismove {
								player.X -= player.Speed
								sprite.Side = player.Direction
								sprite.X -= player.Speed
							}
						case Direction_up:
							c, ok := level.collisionY[(player.Y-1)+frame.config.Height]
							if ok {
								for x := (player.X); x < (player.X)+frame.config.Width; x++ {
									if SearchFloat(c, x) { //found coordinate Значит есть пересечение по оси х
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
							c, ok := level.collisionY[(player.Y+1)+frame.config.Height]
							if ok {
								for x := (player.X); x < (player.X)+frame.config.Width; x++ {
									if SearchFloat(c, x) { //found coordinate
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

				}
				sort.Slice(sprites, sortsprites)
				u.mx.Unlock()
			case <-donech:
				return
			}
		}
	}()

}

func Update(screen *e.Image) error {
	if e.IsDrawingSkipped() {
		return nil
	}

	handleKeyboard()
	units.mx.RLock()
	handleCamera(screen)

	frame++

	for _, sprite := range sprites {
		op := &e.DrawImageOptions{}
		op.GeoM.Reset()
		if sprite.Side == Direction_left {
			op.GeoM.Scale(-1, 1)
			op.GeoM.Translate(sprite.Config.Width*1.1-1, 0)
		}
		op.GeoM.Translate(sprite.X-camera.X, sprite.Y-camera.Y)
		err := screen.DrawImage(sprite.Frames[(frame/7+sprite.Frame)%len(sprite.Frames)], op)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	//s := units.GetSprite(myUnitId)
	//log.Println(unit.Action)

	ebitenutil.DebugPrint(screen, fmt.Sprintf("fps %0.2f population %v", e.CurrentFPS(), len(units.unitsprites)))
	units.mx.RUnlock()
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
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	// ... rest of the program ...

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		runtime.GC()    // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}
	eventChan = make(chan Event, 0)
	done := make(chan struct{})
	wg := &sync.WaitGroup{}
	units = NewUnits()
	var err error
	frames, err = LoadResources()
	if err != nil {
		log.Fatal(err)
	}
	level, err = prepareLevel()
	if err != nil {
		log.Fatal(err)
	}
	skins := []string{"chort"}
	myUnitId = uuid.Must(uuid.NewV4(), err)
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	unit = &Unit{
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
	sprite = &Sprite{
		Frames: frames[unit.Skin+"_"+unit.Action].images,
		op:     &e.DrawImageOptions{},
		Frame:  int(unit.Frame),
		X:      unit.X - 1,
		Y:      unit.Y,
		Side:   unit.Side,
		Config: frames[unit.Skin+"_"+unit.Action].config,
	}
	camera = &Camera{
		X:       unit.X,
		Y:       unit.Y,
		Padding: 30,
	}
	sprites = append(sprites, level.objects...)

	units.Add(myUnitId, sprite, unit)
	go units.HandleEvent(eventChan, done, wg)
	go units.Update(done, wg)

	wg.Add(1)
	go func(donech chan struct{}) {
		defer wg.Done()
		var rnd *rand.Rand
		skins := []string{"big_demon", "big_zombie", "goblin", "elf_f"}
		ticker := time.NewTicker(time.Second * 2)
		for {
			select {
			case <-ticker.C:
				rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
				//	time.Sleep(time.Second * 2
				otherUnitId := uuid.Must(uuid.NewV4(), err)
				otherUnit := &Unit{
					Id:        otherUnitId,
					X:         212, //float64(rnd.Intn(level.levelImage.Bounds().Max.X)) , //(rnd.Float64()*float64(level.levelImage.Bounds().Max.X-20)/float64(rnd.Intn(2)))+10,
					Y:         132, //float64(rnd.Intn(level.levelImage.Bounds().Max.Y)) , //(rnd.Float64()*float64(level.levelImage.Bounds().Max.Y-20)/float64(rnd.Intn(2)))+10,
					Frame:     int32(rnd.Intn(4)),
					Skin:      skins[rnd.Intn(len(skins))],
					Action:    "idle",
					Speed:     1,
					Direction: Direction_right,
					Side:      Direction_right,
				}
				otherSprite := &Sprite{
					Frames: frames[otherUnit.Skin+"_"+otherUnit.Action].images,
					op:     &e.DrawImageOptions{},
					Frame:  int(otherUnit.Frame),
					X:      otherUnit.X - 1,
					Y:      otherUnit.Y,
					Side:   otherUnit.Side,
					Config: frames[otherUnit.Skin+"_"+otherUnit.Action].config,
				}
				units.Add(otherUnitId, otherSprite, otherUnit)

				wg.Add(1)
				go func(uid uuid.UUID, echan chan Event) {
					defer wg.Done()
					rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
					ticker := time.NewTicker(time.Second * 1)
					for {
						select {
						case <-ticker.C:
							event := Event{
								idunit:    uid,
								etype:     EvenType(rnd.Intn(2)),
								direction: Direction(rnd.Intn(4)),
							}
							echan <- event
						case <-donech:
							return
						}
					}
				}(otherUnitId, eventChan)

			case <-donech:
				return
			}
		}

	}(done)

	if err := e.Run(Update, int(config.width), int(config.height), config.scale, config.title); err != nil {
		log.Fatal(err)
	}
	fmt.Println("все")
	close(done)
	wg.Wait()
}

func prepareLevel() (*Level, error) {
	m, err := LoadMapTMX(10)
	if err != nil {
		return nil, err
	}

		op := &e.DrawImageOptions{}
		op.GeoM.Translate(m.config.Width, m.config.Height)
		img := m.image
		level := Level{
			levelImage: img,
			collisionX: m.Coll.ColmapX,
			collisionY: m.Coll.ColmapY,
		}

		for _, object := range m.TileObjects {
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
				Config: ImageConfig{
					ColorModel: object.TileImage.ColorModel(),
					Width:      float64(object.TileImage.Bounds().Max.X),
					Height:     float64(object.TileImage.Bounds().Max.Y),
				},
			})

		}

		for k, v := range level.collisionX {
			m := uniqueFloat64(v)
			sort.Float64s(m)
			level.collisionX[k] = m
		}
		for k, v := range level.collisionY {
			m := uniqueFloat64(v)
			sort.Float64s(m)
			level.collisionY[k] = m
		}
		return &level, nil

	return nil, errors.New("can't load map")
}

func handleCamera(screen *e.Image) {
	frame := frames[unit.Skin+"_"+unit.Action]
	camera.X = unit.X - (config.width-frame.config.Width)/2
	camera.Y = unit.Y - (config.height-frame.config.Height)/2

	op := &e.DrawImageOptions{}
	op.GeoM.Translate(-camera.X, -camera.Y)
	screen.DrawImage(level.levelImage, op)
}

var event Event

func handleKeyboard() {
	etype = idle
	lastKey = e.Key0
	if e.IsKeyPressed(e.KeyA) || e.IsKeyPressed(e.KeyLeft) {
		direction = Direction_left
		etype = move
		if lastKey != e.KeyLeft {
			lastKey = e.KeyLeft
		}
	}

	if e.IsKeyPressed(e.KeyD) || e.IsKeyPressed(e.KeyRight) {
		direction = Direction_right
		etype = move
		if lastKey != e.KeyRight {
			lastKey = e.KeyRight
		}
	}

	if e.IsKeyPressed(e.KeyW) || e.IsKeyPressed(e.KeyUp) {
		etype = move
		direction = Direction_up
		if lastKey != e.KeyUp {
			lastKey = e.KeyUp
		}
	}

	if e.IsKeyPressed(e.KeyS) || e.IsKeyPressed(e.KeyDown) {
		direction = Direction_down
		etype = move
		if lastKey != e.KeyDown {
			lastKey = e.KeyDown
		}
	}

	if prevKey != lastKey {
		event = Event{
			idunit:    myUnitId,
			etype:     etype,
			direction: direction,
		}
		eventChan <- event
		prevKey = lastKey
	}
}
//sed -Ein "s/([- ,])([0-9]+)\.[0-9]+/\1\2/"g %mapfile
func LoadMapTMX(mapId int) (TmxMap, error) {
	var tmxmap TmxMap
	statikFS, err := fs.New()
	if err != nil {
		log.Fatal(err)
	}
	err = fs.Walk(statikFS, "/map/tmx", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}
		// Access individual files by their paths.
		if !strings.Contains(path, strconv.Itoa(mapId)+`_map.tmx`) { //todo: replace on _map.tmx
			return nil
		}
		l := tiled.Loader{FileSystem: statikFS}
		gameMap, err := l.LoadFromFile(path)

		if err != nil {
			return err
		}
/*		for i, group := range gameMap.ObjectGroups{
			for i2, object := range group.Objects {
				for i3, polygon := range object.Polygons {

				}
			}
		}*/

		renderer, err := render.NewRenderer(gameMap)
		if err != nil {
			fmt.Println("Error parsing map")
			return nil
		}

		collision, err := renderer.RenderVisibleLayers()//todo: return collision tileobjects in go-tiled
		if err != nil {
			return err
		}
		_, err= renderer.RenderLayer(0)
		if err != nil {
			return err
		}
		img, err := e.NewImageFromImage(renderer.Result, e.FilterDefault)
		if err != nil {
			return err
		}
		tmxmap=TmxMap{
			image:        img,
			Coll:         collision,
			objectGroups: gameMap.ObjectGroups,
			config: ImageConfig{
				ColorModel: img.ColorModel(),
				Width:      float64(img.Bounds().Max.X),
				Height:     float64(img.Bounds().Max.Y),
			},
		}
		return nil
	})
	return tmxmap, err
}

func LoadResources() (map[string]Frames, error) {
	images := map[string]*e.Image{}
	cfgs := map[string]ImageConfig{}
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
		cfgs[info.Name()] = ImageConfig{
			ColorModel: cfg.ColorModel,
			Width:      float64(cfg.Width),
			Height:     float64(cfg.Height),
		}
		f.Close()
		return nil
	})

	sprites["big_demon_idle"] = Frames{
		images: []*e.Image{
			images["big_demon_idle_anim_f0.png"],
			images["big_demon_idle_anim_f1.png"],
			images["big_demon_idle_anim_f2.png"],
			images["big_demon_idle_anim_f3.png"],
		},
		config: cfgs["big_demon_idle_anim_f0.png"],
	}
	sprites["big_demon_run"] = Frames{
		images: []*e.Image{
			images["big_demon_run_anim_f0.png"],
			images["big_demon_run_anim_f1.png"],
			images["big_demon_run_anim_f2.png"],
			images["big_demon_run_anim_f3.png"],
		},
		config: cfgs["big_demon_run_anim_f0.png"],
	}
	sprites["goblin_idle"] = Frames{
		images: []*e.Image{
			images["goblin_idle_anim_f0.png"],
			images["goblin_idle_anim_f1.png"],
			images["goblin_idle_anim_f2.png"],
			images["goblin_idle_anim_f3.png"],
		},
		config: cfgs["goblin_idle_anim_f0.png"],
	}
	sprites["goblin_run"] = Frames{
		images: []*e.Image{
			images["goblin_run_anim_f0.png"],
			images["goblin_run_anim_f1.png"],
			images["goblin_run_anim_f2.png"],
			images["goblin_run_anim_f3.png"],
		},
		config: cfgs["goblin_run_anim_f0.png"],
	}
	sprites["big_zombie_idle"] = Frames{
		images: []*e.Image{
			images["big_zombie_idle_anim_f0.png"],
			images["big_zombie_idle_anim_f1.png"],
			images["big_zombie_idle_anim_f2.png"],
			images["big_zombie_idle_anim_f3.png"],
		},
		config: cfgs["big_zombie_idle_anim_f0.png"],
	}
	sprites["big_zombie_run"] = Frames{
		images: []*e.Image{
			images["big_zombie_run_anim_f0.png"],
			images["big_zombie_run_anim_f1.png"],
			images["big_zombie_run_anim_f2.png"],
			images["big_zombie_run_anim_f3.png"],
		},
		config: cfgs["big_zombie_run_anim_f0.png"],
	}
	sprites["chort_idle"] = Frames{
		images: []*e.Image{
			images["chort_idle_anim_f0.png"],
			images["chort_idle_anim_f1.png"],
			images["chort_idle_anim_f2.png"],
			images["chort_idle_anim_f3.png"],
		},
		config: cfgs["chort_idle_anim_f0.png"],
	}
	sprites["chort_run"] = Frames{
		images: []*e.Image{
			images["chort_run_anim_f0.png"],
			images["chort_run_anim_f1.png"],
			images["chort_run_anim_f2.png"],
			images["chort_run_anim_f3.png"],
		},
		config: cfgs["chort_run_anim_f0.png"],
	}
	sprites["elf_f_idle"] = Frames{
		images: []*e.Image{
			images["elf_f_idle_anim_f0.png"],
			images["elf_f_idle_anim_f1.png"],
			images["elf_f_idle_anim_f2.png"],
			images["elf_f_idle_anim_f3.png"],
		},
		config: cfgs["elf_f_idle_anim_f0.png"],
	}
	sprites["elf_f_run"] = Frames{
		images: []*e.Image{
			images["elf_f_run_anim_f0.png"],
			images["elf_f_run_anim_f1.png"],
			images["elf_f_run_anim_f2.png"],
			images["elf_f_run_anim_f3.png"],
		},
		config: cfgs["elf_f_run_anim_f0.png"],
	}
	sprites["floor_1"] = Frames{
		images: []*e.Image{images["floor_1.png"]},
		config: cfgs["floor_1.png"],
	}
	sprites["floor_2"] = Frames{
		images: []*e.Image{images["floor_2.png"]},
		config: cfgs["floor_2.png"],
	}
	sprites["floor_3"] = Frames{
		images: []*e.Image{images["floor_3.png"]},
		config: cfgs["floor_3.png"],
	}
	sprites["floor_4"] = Frames{
		images: []*e.Image{images["floor_4.png"]},
		config: cfgs["floor_4.png"],
	}
	sprites["floor_5"] = Frames{
		images: []*e.Image{images["floor_5.png"]},
		config: cfgs["floor_5.png"],
	}
	sprites["floor_6"] = Frames{
		images: []*e.Image{images["floor_6.png"]},
		config: cfgs["floor_6.png"],
	}
	sprites["floor_7"] = Frames{
		images: []*e.Image{images["floor_7.png"]},
		config: cfgs["floor_7.png"],
	}
	sprites["floor_8"] = Frames{
		images: []*e.Image{images["floor_8.png"]},
		config: cfgs["floor_8.png"],
	}
	sprites["wall_side_front_left"] = Frames{
		images: []*e.Image{images["wall_side_front_left.png"]},
		config: cfgs["wall_side_front_left.png"],
	}
	sprites["wall_side_front_right"] = Frames{
		images: []*e.Image{images["wall_side_front_right.png"]},
		config: cfgs["wall_side_front_right.png"],
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
func uniqueFloat64(intSlice []float64) []float64 {
	keys := make(map[float64]bool)
	list := []float64{}
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
func SearchFloat(a []float64, x float64) bool {
	if x < a[0] || x > a[len(a)-1] {
		return false
	}

	for i := 0; i < len(a); i++ {
		if a[i] == x {
			return true
		}
		if a[i] > x {
			return false
		}
	}
	//for _, y := range a {
	//
	//}
	return false
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
