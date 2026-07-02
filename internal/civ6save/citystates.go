package civ6save

import (
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
		for _, s := range all {
			if strings.Contains(s, "MINOR") || strings.Contains(s, "CITY_STATE") {
				out[b.iPlayer] = s
				break
			}
		}
	}
	return out
}

// CityStateColors maps each city-state's player-slot to the colour of its type
// (trade yellow, science blue, …). The save only carries the leader token (see
// minorCivSlots); resolving a token to its type needs a data table that hasn't
// been built yet, so this returns nil and the renderer draws minor civs with
// its neutral border instead of risking wrong colours.
func CityStateColors(data []byte) map[int]color.RGBA {
	return nil
}
