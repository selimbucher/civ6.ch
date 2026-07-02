package civ6save

import (
	"fmt"
	"image/color"
	"strings"
)

// minorCivSlots scans the player blocks and returns, for each city-state /
// minor-civ player slot, its identifying string (e.g. a CIVILIZATION_* or
// LEADER_MINOR_* token), keyed by the player-slot index used as a tile Owner.
func minorCivSlots(data []byte) map[int]string {
	out := map[int]string{}
	pos := 0
	var starts []struct {
		pos, iPlayer int
	}
	for {
		idx := findNext(data, iPlayerIdxMk, pos, len(data))
		if idx == -1 || idx+20 > len(data) {
			break
		}
		ip := int(pktInt(data, idx))
		if ip <= 63 {
			starts = append(starts, struct{ pos, iPlayer int }{idx, ip})
		}
		pos = idx + 1
	}
	for i, b := range starts {
		end := len(data)
		if i+1 < len(starts) {
			end = starts[i+1].pos
		}
		lm := b.pos
		var all []string
		for {
			lm = findNext(data, leaderStrMk, lm+1, end)
			if lm == -1 {
				break
			}
			s := pktStr(data, lm)
			if s != "" {
				all = append(all, s)
			}
		}
		fmt.Printf("DEBUG slot=%d strings=%q\n", b.iPlayer, all)
		for _, s := range all {
			if strings.Contains(s, "MINOR") || strings.Contains(s, "CITY_STATE") {
				out[b.iPlayer] = s
				break
			}
		}
	}
	return out
}

// CityStateColors maps each city-state's player-slot to the colour of its type.
func CityStateColors(data []byte) map[int]color.RGBA {
	slots := minorCivSlots(data)
	for idx, s := range slots {
		fmt.Printf("DEBUG citystate slot=%d str=%q\n", idx, s)
	}
	return nil
}
