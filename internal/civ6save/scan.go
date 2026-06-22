// Marker-scan extraction of basic player info (leader, pseudo, color).
// Unlike the sequential parsing pipeline in parse.go, this scans the whole
// file for known byte markers and produces lightweight Player records
// rather than the full PlayerState.

package civ6save

import (
	"bytes"
	"encoding/binary"
	"strings"
)

var iPlayerIdxMk = []byte{0x2f, 0x52, 0x96, 0x1a}
var typePlayerMk = []byte{0x95, 0xb9, 0x42, 0xce}
var pseudoMk = []byte{0xfd, 0x6b, 0xb9, 0xda}
var leaderStrMk = []byte{0x5f, 0x5e, 0xcd, 0xe8}
var iColorMk = []byte{0xef, 0x60, 0xaf, 0xcf}
var teamMk = []byte{0x54, 0xb4, 0x8a, 0x0d}    // 0x0D8AB454 — shared-victory team id
var steamIDMk = []byte{0x9a, 0x24, 0x72, 0x8e} // 0x8E72249A — "persona@steamid64"

const typePlayerFull = 3

type Player struct {
	Index  int
	Leader string
	Pseudo string
	IColor int
	// Team is the save's shared-victory team id. Players sharing a Team win
	// together; in FFA each player has a unique Team (alliances do not change it).
	Team int
	// Eliminated is true for major players whose slot is no longer a live human
	// (typePlayer != 3) — this covers both genuinely eliminated civs and humans
	// who left the game and were taken over by the AI. Use LivingCivilization to
	// tell the two apart: a player who left still controls cities.
	Eliminated bool
	// SteamID is the stable SteamID64 extracted from the "persona@steamid64"
	// packet, used to match a save slot to a registered player.
	SteamID string
}

func pktInt(data []byte, offset int) uint32 {
	return binary.LittleEndian.Uint32(data[offset+16:])
}

func pktStr(data []byte, offset int) string {
	b := data[offset+8 : offset+11]
	length := int(b[0]) | int(b[1])<<8 | int(b[2])<<16
	if length <= 0 || offset+16+length > len(data) {
		return ""
	}
	raw := data[offset+16 : offset+16+length]
	for i, b := range raw {
		if b == 0 {
			raw = raw[:i]
			break
		}
	}
	return string(raw)
}

// findNext returns the index of the first occurrence of marker in
// data[start:end], or -1 if absent.
func findNext(data []byte, marker []byte, start, end int) int {
	if start < 0 || start >= end {
		return -1
	}
	i := bytes.Index(data[start:end], marker)
	if i == -1 {
		return -1
	}
	return start + i
}

func ParsePlayers(data []byte) []Player {
	// collect all block start positions
	type block struct {
		pos     int
		iPlayer int
	}
	var blocks []block
	pos := 0
	for {
		idx := findNext(data, iPlayerIdxMk, pos, len(data))
		if idx == -1 {
			break
		}
		if idx+20 > len(data) {
			break
		}
		iPlayer := int(pktInt(data, idx))
		if iPlayer <= 63 {
			blocks = append(blocks, block{idx, iPlayer})
		}
		pos = idx + 1
	}

	seen := make(map[int]bool)
	var players []Player

	for i, b := range blocks {
		if seen[b.iPlayer] {
			continue
		}
		blockEnd := len(data)
		if i+1 < len(blocks) {
			blockEnd = blocks[i+1].pos
		}

		// typePlayer marks the slot kind: 3 = active major (human), 1 = AI.
		// A human who is eliminated has their slot flipped to AI, so we cannot
		// rely on this value alone — we keep alive majors (==3) and, for other
		// values, only those that still carry a human pseudo (eliminated humans),
		// which excludes genuine AI fill, city-states and barbarians.
		tp := findNext(data, typePlayerMk, b.pos, blockEnd)
		if tp == -1 {
			seen[b.iPlayer] = true
			continue
		}
		eliminated := int(pktInt(data, tp)) != typePlayerFull

		// leaderStr — only real major civs pass this filter, which excludes
		// city-states, free cities and barbarians.
		leader := ""
		lm := b.pos
		for {
			lm = findNext(data, leaderStrMk, lm+1, blockEnd)
			if lm == -1 {
				break
			}
			s := pktStr(data, lm)
			if len(s) > 7 && s[:7] == "LEADER_" &&
				!strings.Contains(s, "MINOR") &&
				!strings.Contains(s, "FREE") &&
				!strings.Contains(s, "LOC_") &&
				!strings.Contains(s, "_NAME") {
				leader = s
				break
			}
		}
		if leader == "" {
			seen[b.iPlayer] = true
			continue
		}

		// pseudo
		pm := findNext(data, pseudoMk, b.pos, blockEnd)
		pseudo := ""
		if pm != -1 {
			pseudo = pktStr(data, pm)
		}

		// Skip non-human slots: an eliminated slot with no pseudo is AI fill,
		// not a tracked player.
		if eliminated && pseudo == "" {
			seen[b.iPlayer] = true
			continue
		}

		// iColor
		ic := findNext(data, iColorMk, b.pos, blockEnd)
		icolor := 0
		if ic != -1 {
			icolor = int(pktInt(data, ic))
		}

		// team — shared-victory team id. The team packet sits immediately
		// before the player's iPlayer marker, so scan backwards for it.
		team := b.iPlayer
		lo := b.pos - 64
		if lo < 0 {
			lo = 0
		}
		if rel := bytes.LastIndex(data[lo:b.pos], teamMk); rel != -1 {
			team = int(pktInt(data, lo+rel))
		}

		// SteamID — the "persona@steamid64" packet; keep the part after '@'.
		steamID := ""
		if sm := findNext(data, steamIDMk, b.pos, blockEnd); sm != -1 {
			if s := pktStr(data, sm); s != "" {
				if at := strings.LastIndex(s, "@"); at != -1 {
					steamID = s[at+1:]
				}
			}
		}

		seen[b.iPlayer] = true
		players = append(players, Player{
			Index:      b.iPlayer,
			Leader:     leader,
			Pseudo:     pseudo,
			IColor:     icolor,
			Team:       team,
			Eliminated: eliminated,
			SteamID:    steamID,
		})
	}

	return players
}

// LivingCivilization reports whether player index still controls at least one
// city in the parsed state. Combined with Player.Eliminated this distinguishes a
// player who *left* the game (slot taken over by the AI but the civ is alive,
// cities > 0) from one who was genuinely *eliminated* (no cities left).
func LivingCivilization(state *GameState, index int) bool {
	if state == nil || index < 0 || index >= len(state.Players) {
		return false
	}
	ps := state.Players[index]
	return ps != nil && len(ps.Cities) > 0
}
