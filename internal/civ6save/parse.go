package civ6save

import (
	"bytes"
	"encoding/binary"
	"regexp"
	"strings"
)

// GameSettings holds lobby/setup parameters read from the pre-compressed save header.
type GameSettings struct {
	Map              string   // e.g. "Shuffle", "Pangaea Ultima"
	MapSize          string   // e.g. "MAPSIZE_SMALL"
	GameSpeed        string   // e.g. "GAMESPEED_ONLINE"
	Difficulty       string   // e.g. "DIFFICULTY_PRINCE"
	Ruleset          string   // e.g. "RULESET_EXPANSION_2" (Gathering Storm)
	CurrentEra       string   // e.g. "ERA_INDUSTRIAL" (current game era)
	Options          []string // e.g. ["GAMEOPTION_NO_BARBARIANS"]
	Modes            []string // e.g. ["MONOPOLIES", "SECRETSOCIETIES"]
	EnabledVictories []string // e.g. ["VICTORY_CONQUEST", "VICTORY_CULTURE"]
	MapScript        string   // e.g. "StandardMaps/Shuffle.lua"
	MultiPlayer      bool
}

var (
	reMapText  = regexp.MustCompile(`"text":"([^"]+)"`)
	reGameMode = regexp.MustCompile(`LOC_GAMEMODE_([A-Z0-9_]+?)_NAME`)
)

// ParseTurn reads the current turn number from the raw save bytes.
func ParseTurn(data []byte) int {
	return int(binary.LittleEndian.Uint32(data[377:]))
}

// ParseSettings reads game lobby settings from the pre-compressed header region.
func ParseSettings(data []byte) GameSettings {
	var s GameSettings

	// Work only in the pre-zlib region.
	modIdx := bytes.LastIndex(data, modTitle)
	if modIdx < 0 {
		modIdx = len(data)
	}
	zlibIdx := bytes.Index(data[modIdx:], zlibMagic)
	if zlibIdx < 0 {
		zlibIdx = modIdx
	} else {
		zlibIdx += modIdx
	}
	pre := data[:zlibIdx]

	// Game modes — extracted from {"modes":[...]} JSON blob.
	// Each active mode has a "LOC_GAMEMODE_FOO_NAME" key; we extract "FOO".
	if idx := bytes.Index(pre, []byte(`"modes"`)); idx >= 0 {
		// Back up to the opening brace of the enclosing object.
		brace := bytes.LastIndexByte(pre[:idx], '{')
		if brace >= 0 {
			blob := pre[brace:]
			// Find the closing brace of this top-level object.
			depth := 0
			for i, c := range blob {
				if c == '{' {
					depth++
				} else if c == '}' {
					depth--
					if depth == 0 {
						blob = blob[:i+1]
						break
					}
				}
			}
			seen := map[string]bool{}
			for _, m := range reGameMode.FindAllSubmatch(blob, -1) {
				key := string(m[1])
				if !seen[key] {
					s.Modes = append(s.Modes, key)
					seen[key] = true
				}
			}
		}
	}

	// Map name — extracted from the JSON localisation blob at marker 0x584C6027.
	if idx := bytes.Index(pre, binary.LittleEndian.AppendUint32(nil, 0x584C6027)); idx >= 0 {
		blob := pre[idx+16:]
		if m := reMapText.FindSubmatch(blob); m != nil {
			s.Map = string(m[1])
		}
	}

	// Map script path e.g. "StandardMaps/Shuffle.lua"
	for _, prefix := range [][]byte{[]byte("StandardMaps/"), []byte("Maps/")} {
		if idx := bytes.Index(pre, prefix); idx >= 0 {
			end := bytes.IndexByte(pre[idx:], 0x00)
			if end < 0 {
				end = 60
			}
			s.MapScript = string(pre[idx : idx+end])
			break
		}
	}

	// Null-terminated string helper.
	readStr := func(prefix string) string {
		p := []byte(prefix)
		idx := bytes.Index(pre, p)
		if idx < 0 {
			return ""
		}
		end := bytes.IndexByte(pre[idx:], 0x00)
		if end < 0 {
			end = 60
		}
		return string(pre[idx : idx+end])
	}

	s.MapSize = readStr("MAPSIZE_")
	s.GameSpeed = readStr("GAMESPEED_")

	// Ruleset: first RULESET_ string e.g. "RULESET_EXPANSION_2".
	s.Ruleset = readStr("RULESET_")

	// Current era: marker 0xE7170E55 + type=5 packet, string follows at +16.
	eraMarker := binary.LittleEndian.AppendUint32(nil, 0xE7170E55)
	if idx := bytes.Index(pre, eraMarker); idx >= 0 && idx+20 <= len(pre) {
		end := bytes.IndexByte(pre[idx+16:], 0x00)
		if end < 0 {
			end = 40
		}
		s.CurrentEra = string(pre[idx+16 : idx+16+end])
	}

	// Difficulty: find "DIFFICULTY_XXXX_NAME…" and trim the suffix.
	if d := readStr("DIFFICULTY_"); d != "" {
		if cut := strings.Index(d, "_NAME"); cut > 0 {
			d = d[:cut]
		}
		s.Difficulty = d
	}

	// Game options (e.g. GAMEOPTION_NO_BARBARIANS).
	pos := 0
	for {
		p := []byte("GAMEOPTION_")
		idx := bytes.Index(pre[pos:], p)
		if idx < 0 {
			break
		}
		idx += pos
		end := bytes.IndexByte(pre[idx:], 0x00)
		if end < 0 {
			end = 60
		}
		s.Options = append(s.Options, string(pre[idx:idx+end]))
		pos = idx + end + 1
	}

	// Multiplayer: marker 0x7C546F81, type=3, value == 0x453F23E2.
	if idx := bytes.Index(pre, binary.LittleEndian.AppendUint32(nil, 0x7C546F81)); idx >= 0 && idx+20 <= len(pre) {
		val := binary.LittleEndian.Uint32(pre[idx+16:])
		s.MultiPlayer = val == 0x453F23E2
	}

	// Enabled victory conditions: each is stored as CRC(name) + u32(1=enabled).
	// CRCs are derived from internal Civ6 victory type names.
	victoryMarkers := []struct {
		crc  uint32
		name string
	}{
		{0x1843ff8c, "VICTORY_CONQUEST"},   // Domination
		{0x150c2d79, "VICTORY_TECHNOLOGY"}, // Science
		{0xeabc48eb, "VICTORY_CULTURE"},    // Culture
		{0x5529a9bb, "VICTORY_SCORE"},      // Score
		{0xe2898d23, "VICTORY_DIPLOMATIC"}, // Diplomatic (GS)
		{0x18c44790, "VICTORY_RELIGIOUS"},  // Religion
	}
	for _, v := range victoryMarkers {
		pat := binary.LittleEndian.AppendUint32(nil, v.crc)
		if idx := bytes.Index(pre, pat); idx >= 0 && idx+8 <= len(pre) {
			enabled := binary.LittleEndian.Uint32(pre[idx+4:])
			if enabled == 1 {
				s.EnabledVictories = append(s.EnabledVictories, v.name)
			}
		}
	}

	return s
}
