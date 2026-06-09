package civ6save

import "encoding/binary"

// GreatPersonRecruit is a recruited great person entry parsed from the global save tail.
type GreatPersonRecruit struct {
	Offset int
	Name   uint32
	Class  uint32
	Era    uint32
	Cost   uint32
	Player int
}

var gpClassSet = map[uint32]bool{
	0xA0426FED: true,
	0x05B84C1A: true,
	0x35A7F88E: true,
	0xD51BB7B6: true,
	0x20FE0A67: true,
	0xFDB5F1E1: true,
	0x9895E813: true,
	0x0E6CA616: true,
	0x90C1D370: true,
}

func u32at(data []byte, i int) uint32 {
	return binary.LittleEndian.Uint32(data[i:])
}

// ParseGreatPeopleRecruitedLoose scans for the contiguous recruited-great-people block.
// It mirrors the C++ struct layout (stride 28):
//
//	[0]=0x05, [4]=name, [8]=class, [12]=era, [16]=cost, [20]=player, [24]=unknown
func ParseGreatPeopleRecruitedLoose(data []byte) []GreatPersonRecruit {
	if len(data) < 28 {
		return nil
	}

	cands := make([]GreatPersonRecruit, 0, 64)
	for i := 0; i+28 <= len(data); i++ {
		if u32at(data, i) != 5 {
			continue
		}
		class := u32at(data, i+8)
		if !gpClassSet[class] {
			continue
		}
		era := u32at(data, i+12)
		if era > 12 {
			continue
		}
		cost := u32at(data, i+16)
		if cost == 0 || cost > 1000000 {
			continue
		}
		player := u32at(data, i+20)
		if player > 63 {
			continue
		}
		cands = append(cands, GreatPersonRecruit{
			Offset: i,
			Name:   u32at(data, i+4),
			Class:  class,
			Era:    era,
			Cost:   cost,
			Player: int(player),
		})
	}
	if len(cands) == 0 {
		return nil
	}

	// Keep the longest contiguous run with fixed stride 28.
	bestStart, bestLen := 0, 1
	start, ln := 0, 1
	for i := 1; i < len(cands); i++ {
		if cands[i].Offset == cands[i-1].Offset+28 {
			ln++
			continue
		}
		if ln > bestLen {
			bestStart, bestLen = start, ln
		}
		start, ln = i, 1
	}
	if ln > bestLen {
		bestStart, bestLen = start, ln
	}

	out := make([]GreatPersonRecruit, bestLen)
	copy(out, cands[bestStart:bestStart+bestLen])
	return out
}
