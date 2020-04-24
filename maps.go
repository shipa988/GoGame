package main

var _ Maper = (*MapbyTMX)(nil)

type Maper interface {
	LoadMap(mapId int) map[string]Frames
}

type MapbyTMX struct {
	mapLayers map[string]Frames
}

func (m *MapbyTMX) LoadMap(mapId int) error {
	if _map, ok := maps[mapId]; ok {
		_map.(*MapbyTMX)
		m=_map()
		return &w, nil
	}
	m.mapLayers = nil
	if m.mapLayers == nil {
		m.mapLayers = make(map[string]Frames)
	}
	
	maps
	return m.mapLayers
}

type Map map[string]Frames 
type Maps map[int]*Map

func ()  {
	
}

func (m *Maps) LoadMap(mapid int) error {
	m2 := m[mapid]
	layers, err := LoadMapTMX(mapid)
	if err != nil {
		return err
	}
	m[mapid]=*Map(layers)
}

