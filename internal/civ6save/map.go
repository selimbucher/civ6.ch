package civ6save

import (
	"encoding/binary"
	"errors"
)

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

// findMapStart locates the tile-count field that precedes the map's tile array.
//
// The tile array is preceded by the plot-data layer table, serialized as:
//
//	[count][lo][hi]  [count ascending layer IDs: lo, lo+1, … hi]  [6]  [tileCount]
//
// The layer range always ends at 15, but the number of layers varies by save:
// 12 (IDs 4–15) in older saves, 6 (10–15) in current base-game saves, and as
// few as 3 (13–15) in some modded/Secret-Societies saves. The previous fixed
// 28-byte marker assumed the six trailing IDs 10–15, so it failed to find the
// map in saves with fewer layers. We match the structure instead, returning the
// offset of the tileCount dword (or -1 if not found).
func findMapStart(data []byte) int {
	n := len(data)
	for i := 0; i+12 < n; i++ {
		count := le32(data, i)
		if count < 1 || count > 16 {
			continue
		}
		lo, hi := le32(data, i+4), le32(data, i+8)
		if hi != 15 || hi < lo || hi-lo+1 != count {
			continue
		}
		list := i + 12
		ok := true
		for k := uint32(0); k < count; k++ {
			if list+int(k)*4+4 > n || le32(data, list+int(k)*4) != lo+k {
				ok = false
				break
			}
		}
		if !ok {
			continue
		}
		after := list + int(count)*4
		if after+8 > n || le32(data, after) != 6 {
			continue
		}
		return after + 4 // tileCount dword
	}
	return -1
}

func ParseMap(data []byte) (*Map, error) {
	idx := findMapStart(data)
	if idx == -1 {
		return nil, errors.New("map begin marker not found")
	}

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
