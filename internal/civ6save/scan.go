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

const typePlayerFull = 3

type Player struct {
	Index  int
	Leader string
	Pseudo string
	IColor int
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

		// typePlayer
		tp := findNext(data, typePlayerMk, b.pos, blockEnd)
		if tp == -1 {
			seen[b.iPlayer] = true
			continue
		}
		if int(pktInt(data, tp)) != typePlayerFull {
			seen[b.iPlayer] = true
			continue
		}

		// leaderStr
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

		// iColor
		ic := findNext(data, iColorMk, b.pos, blockEnd)
		icolor := 0
		if ic != -1 {
			icolor = int(pktInt(data, ic))
		}

		seen[b.iPlayer] = true
		players = append(players, Player{
			Index:  b.iPlayer,
			Leader: leader,
			Pseudo: pseudo,
			IColor: icolor,
		})
	}

	return players
}
