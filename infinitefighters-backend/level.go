package main

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Level is the collision view of a Tiled map. The client renders the same JSON file.
//
// Tiled rules the server relies on:
//   - orthogonal, finite map, square tiles, tile layer data as a plain JSON array (CSV)
//   - tilesets embedded in the map (no external .tsx)
//   - a tileset tile collides when it has the string property collision = "solid" | "oneway";
//     tiles without it are decoration
//   - point objects of type "spawn" give player spawn feet positions, in order P1..P5
type Level struct {
	W, H, Tile int // size in tiles, tile size in px
	tiles      []tileKind
	Spawns     []Point
}

type Point struct{ X, Y float64 }

type tileKind uint8

const (
	tileEmpty tileKind = iota
	tileSolid
	tileOneWay
)

// at treats the left, right and bottom edges of the map as walls; above the map is open.
func (l *Level) at(col, row int) tileKind {
	switch {
	case col < 0 || col >= l.W || row >= l.H:
		return tileSolid
	case row < 0:
		return tileEmpty
	}
	return l.tiles[row*l.W+col]
}

type tiledMap struct {
	Width, Height int
	TileWidth     int `json:"tilewidth"`
	TileHeight    int `json:"tileheight"`
	Infinite      bool
	Layers        []struct {
		Type    string
		Data    []uint32
		Objects []struct {
			Type string
			X, Y float64
		}
	}
	Tilesets []struct {
		FirstGID uint32 `json:"firstgid"`
		Source   string
		Tiles    []struct {
			ID         uint32
			Properties []struct {
				Name  string
				Value any
			}
		}
	}
}

const gidMask = 0x1FFFFFFF // strips Tiled's flip/rotate flag bits

func parseLevel(data []byte) (*Level, error) {
	var m tiledMap
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("tiled map: %w (tile layers must use CSV, not base64)", err)
	}
	if m.Infinite || m.Width <= 0 || m.Height <= 0 {
		return nil, errors.New("tiled map: must be finite with a positive size")
	}
	if m.TileWidth <= 0 || m.TileWidth != m.TileHeight {
		return nil, errors.New("tiled map: tiles must be square")
	}

	kinds := map[uint32]tileKind{} // global tile id → kind
	for _, ts := range m.Tilesets {
		if ts.Source != "" {
			return nil, fmt.Errorf("tiled map: tileset %q must be embedded in the map", ts.Source)
		}
		for _, t := range ts.Tiles {
			for _, p := range t.Properties {
				if p.Name != "collision" {
					continue
				}
				switch p.Value {
				case "solid":
					kinds[ts.FirstGID+t.ID] = tileSolid
				case "oneway":
					kinds[ts.FirstGID+t.ID] = tileOneWay
				default:
					return nil, fmt.Errorf("tiled map: tile %d has unknown collision %v", t.ID, p.Value)
				}
			}
		}
	}

	l := &Level{W: m.Width, H: m.Height, Tile: m.TileWidth, tiles: make([]tileKind, m.Width*m.Height)}
	for _, layer := range m.Layers {
		switch layer.Type {
		case "tilelayer":
			if len(layer.Data) != len(l.tiles) {
				return nil, errors.New("tiled map: tile layer size does not match the map")
			}
			for i, gid := range layer.Data {
				if k := kinds[gid&gidMask]; k != tileEmpty {
					l.tiles[i] = k
				}
			}
		case "objectgroup":
			for _, o := range layer.Objects {
				if o.Type == "spawn" {
					l.Spawns = append(l.Spawns, Point{o.X, o.Y})
				}
			}
		}
	}
	if len(l.Spawns) == 0 {
		return nil, errors.New(`tiled map: needs at least one "spawn" point object`)
	}
	return l, nil
}
