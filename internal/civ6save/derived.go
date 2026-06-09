package civ6save

import "encoding/binary"

func applyDerivedState(gs *GameState, data []byte) {
	// Religion founder and foreign followers.
	founderBySymbol := make(map[uint32]int, len(gs.Religions))
	for _, rel := range gs.Religions {
		founderBySymbol[rel.Symbol] = rel.FounderPlayer
		if rel.FounderPlayer >= 0 && rel.FounderPlayer < len(gs.Players) && gs.Players[rel.FounderPlayer] != nil {
			gs.Players[rel.FounderPlayer].ReligionFounded = true
		}
	}

	// Foreign follower score uses majority religion per city:
	// count foreign cities whose majority religion symbol matches the founded religion.
	for owner, ps := range gs.Players {
		if ps == nil {
			continue
		}
		for _, c := range ps.Cities {
			if c.Religion == 0 || c.Religion == 0xFFFFFFFF {
				continue
			}
			founder, ok := founderBySymbol[c.Religion]
			if !ok {
				continue
			}
			if founder == owner {
				continue
			}
			if founder < 0 || founder >= len(gs.Players) || gs.Players[founder] == nil {
				continue
			}
			gs.Players[founder].ForeignCitiesFollowingReligion++
		}
	}

	// Era score (loose pattern parse from global tail).
	era := parseEraScoresLoose(data)
	for i, v := range era {
		if gs.Players[i] != nil {
			gs.Players[i].EraScore = v
		}
	}
}

func parseEraScoresLoose(data []byte) map[int]int {
	out := map[int]int{}
	if len(data) < 700 {
		return out
	}

	u32 := func(i int) uint32 { return binary.LittleEndian.Uint32(data[i:]) }
	isBool64 := func(off int) bool {
		if off+64 > len(data) {
			return false
		}
		for i := 0; i < 64; i++ {
			b := data[off+i]
			if b != 0 && b != 1 {
				return false
			}
		}
		return true
	}

	for off := 0; off+700 < len(data); off++ {
		if u32(off) != 0x40 {
			continue
		}
		if !isBool64(off + 4) {
			continue
		}
		if u32(off+68) != 0x40 {
			continue
		}
		if !isBool64(off + 72) {
			continue
		}

		pos := off + 136
		if u32(pos) != 0x40 {
			continue
		}
		pos += 4

		ok := true
		// First 0x40 block: commemorations (variable lengths per player).
		for i := 0; i < 64; i++ {
			if pos+16 > len(data) || u32(pos) != 0x1E {
				ok = false
				break
			}
			pos += 4
			count2 := int(u32(pos))
			pos += 4 + count2*4
			if pos+4 > len(data) {
				ok = false
				break
			}
			pos += 4 // 00
			if pos+4 > len(data) {
				ok = false
				break
			}
			count2 = int(u32(pos))
			pos += 4 + count2*4
			if pos+4 > len(data) {
				ok = false
				break
			}
			pos += 4 // choices to make
		}
		if !ok || pos+4 > len(data) || u32(pos) != 0x40 {
			continue
		}
		pos += 4

		tmp := map[int]int{}
		for i := 0; i < 64; i++ {
			if pos+184 > len(data) {
				ok = false
				break
			}
			if u32(pos) != 0x1E {
				ok = false
				break
			}
			if u32(pos+4) != 0x16 {
				ok = false
				break
			}
			e := int(u32(pos + 8))
			if e < 0 || e > 10000 {
				ok = false
				break
			}
			tmp[i] = e
			pos += 184
		}
		if !ok {
			continue
		}

		nz := 0
		for _, v := range tmp {
			if v > 0 {
				nz++
			}
		}
		if nz == 0 {
			continue
		}
		return tmp
	}

	return out
}
