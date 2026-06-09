package civ6save

import (
	"bytes"
	"encoding/binary"
	"errors"
)

var mapBegin = []byte{
	0x0A, 0, 0, 0,
	0x0B, 0, 0, 0,
	0x0C, 0, 0, 0,
	0x0D, 0, 0, 0,
	0x0E, 0, 0, 0,
	0x0F, 0, 0, 0,
	0x06, 0, 0, 0,
}

type Tile struct {
	Index    int
	Terrain  uint32
	Feature  uint32
	Resource uint32
	Owner    uint16
}

type Map struct {
	Tiles []Tile
	Width int
}

func le32(data []byte, pos int) uint32 {
	return binary.LittleEndian.Uint32(data[pos:])
}

func le16(data []byte, pos int) uint16 {
	return binary.LittleEndian.Uint16(data[pos:])
}

func ParseMap(data []byte) (*Map, error) {
	idx := bytes.Index(data, mapBegin)
	if idx == -1 {
		return nil, errors.New("map begin marker not found")
	}
	idx += 28 // skip the marker

	tileCount := int(le32(data, idx))
	idx += 4

	tiles := make([]Tile, tileCount)

	for i := 0; i < tileCount; i++ {
		t := &tiles[i]
		t.Index = i
		t.Owner = 0xFFFF

		idx += 8 // group[0], group[1]
		idx += 4 // landmass
		t.Terrain = le32(data, idx)
		idx += 4
		t.Feature = le32(data, idx)
		idx += 4
		idx += 2 // unknown int16
		idx += 4 // continent
		idx += 1 // unknown int8
		t.Resource = le32(data, idx)
		idx += 4
		idx += 2 // resourceCount int16
		idx += 4 // improvement
		idx += 1 // unknown int8
		idx += 1 // road
		idx += 1 // roadLvl
		idx += 2 // appeal int16
		idx += 3 // unknown
		idx += 1 // riverCount
		idx += 1 // river
		idx += 1 // cliffbitmap
		idx += 1 // pillage
		found := data[idx]
		idx += 1
		idx += 1 // unknown
		bOverlay := le32(data, idx)
		idx += 4

		if bOverlay != 0 {
			count := int(le32(data, idx))
			idx += 4
			for k := 0; k < count; k++ {
				idx += 11
				value := int(le32(data, idx))
				idx += 4
				idx += 1
				count2 := int(le32(data, idx))
				idx += 4
				if value != 0 {
					idx += count2 * 20
				}
			}
		}

		if found&0x40 != 0 {
			idx += 2 // iCity
			idx += 2 // iCity+1
			idx += 4 // icity-icity+1
			idx += 2 // iDistrict
			idx += 2 // iDistrict+1
			t.Owner = uint16(data[idx])
			idx += 1
			idx += 4 // wonder
		}
	}

	idx += 4
	width := int(le32(data, idx))

	return &Map{Tiles: tiles, Width: width}, nil
}
