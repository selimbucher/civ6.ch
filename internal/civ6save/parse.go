package civ6save

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var (
	reMapText  = regexp.MustCompile(`"text":"([^"]+)"`)
	reGameMode = regexp.MustCompile(`LOC_GAMEMODE_([A-Z0-9_]+?)_NAME`)
)

var territoryBuilder = []byte{
	0x54, 0x65, 0x72, 0x72, 0x69, 0x74, 0x6F, 0x72,
	0x79, 0x42, 0x75, 0x69, 0x6C, 0x64, 0x65, 0x72,
}

// Anchors used to resync the reader inside the diplomatic-state zone.
var (
	endIAStuff     = []byte{0xBA, 0xF1, 0xBF, 0x93}
	endPlayerStuff = []byte{0xBC, 0x0A, 0x2B, 0xDE}
)

// ── public API ────────────────────────────────────────────────────────────────

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
	gameOption := []byte("GAMEOPTION_")
	pos := 0
	for {
		idx := bytes.Index(pre[pos:], gameOption)
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

// ParseState parses the main game body into a GameState.
func ParseState(data []byte) (*GameState, error) {
	idx := bytes.Index(data, territoryBuilder)
	if idx == -1 {
		return nil, errors.New("TERRITORY_BUILDER not found")
	}

	r := newReader(data, idx+16)

	r.skip(8) // 00 07
	r.skip(8) // ?

	gs := &GameState{}

	religions, err := parseReligions(r)
	if err != nil {
		return nil, err
	}
	gs.Religions = religions

	if err := r.assertU32(1); err != nil {
		return nil, errors.New("expected 1 before playerCount")
	}
	r.skip(4)

	playerCount := int(r.readU32())

	for iPlayer := 0; iPlayer < playerCount; iPlayer++ {
		trueiPlayer := int(r.readU32())
		ps, err := parsePlayer(r, trueiPlayer)
		if err != nil {
			return nil, err
		}
		gs.Players[trueiPlayer] = ps
	}

	r.skip(28)

	// Global recruited great people block from the post-player tail.
	recruited, err := parseGreatPeopleRecruitedFromTail(data, r.pos, playerCount)
	if err != nil {
		// Fallback to loose scan; better to return partial state than fail the full parse.
		recruited = ParseGreatPeopleRecruitedLoose(data)
	}
	gs.RecruitedGreatPeople = recruited
	for _, gp := range recruited {
		if gp.Player >= 0 && gp.Player < len(gs.Players) && gs.Players[gp.Player] != nil {
			gs.Players[gp.Player].GreatPeopleRecruited++
		}
	}

	applyDerivedState(gs, data)
	return gs, nil
}

// parseReligions reads the religion section that precedes the player blocks.
// ── private sub-parsers ───────────────────────────────────────────────────────

func parseReligions(r *reader) ([]ReligionState, error) {
	religionCount := int(r.readU32())
	religions := make([]ReligionState, 0, religionCount)
	for i := 0; i < religionCount; i++ {
		if err := r.assertU32(7); err != nil {
			return nil, errors.New("religion: expected 7")
		}
		r.skip(4)
		symbol := r.readU32()
		founder := int(r.readU32())
		r.skip(12)
		name := r.readString()

		beliefCount := int(r.readU32())
		beliefs := make([]uint32, beliefCount)
		for j := 0; j < beliefCount; j++ {
			beliefs[j] = r.readU32()
		}

		buildingCount := int(r.readU32())
		buildings := make([]uint32, buildingCount)
		for j := 0; j < buildingCount; j++ {
			buildings[j] = r.readU32()
		}

		unitCount := int(r.readU32())
		units := make([]uint32, unitCount)
		for j := 0; j < unitCount; j++ {
			units[j] = r.readU32()
		}
		r.skip(11)

		religions = append(religions, ReligionState{
			FounderPlayer: founder,
			Symbol:        symbol,
			Name:          name,
			Beliefs:       beliefs,
			Buildings:     buildings,
			Units:         units,
		})
	}
	return religions, nil
}

// parsePlayer reads one full player block (cities, districts, units, yields, …).
func parsePlayer(r *reader, trueiPlayer int) (*PlayerState, error) {
	ps := &PlayerState{IPlayer: trueiPlayer}

	if err := r.assertU32(47); err != nil {
		return nil, errors.New("expected 47 at player start")
	}
	r.skip(4) // 2F 00 00 00
	r.skip(4) // iPlayer copy

	r.skip(4)
	r.skip(4) // init pos ?
	r.skip(56)
	for i := 0; i < 10; i++ {
		r.skip(1)
	}
	r.skip(4)
	for i := 0; i < 7; i++ {
		r.skip(4)
	}

	// ── Units (position only) ─────────────────────────────────────────
	unitCount := int(r.readU32())
	for iUnit := 0; iUnit < unitCount; iUnit++ {
		r.skip(12) // 3x int
		r.skip(4)  // x
		r.skip(4)  // y
		r.skip(8)
		r.skip(12)
		count2 := int(r.readU32())
		r.skip(8)
		r.skip(count2 * 20)
		count2 = int(r.readU32())
		r.skip(count2 * 4)
	}

	r.skip(4)
	count := int(r.readU32())
	r.skip(count * 4)
	r.skip(12) // 06 07 01

	r.readMapDiscard() // goodyHuts (08)
	r.skip(4)
	r.readMapDiscard() // 0b
	if err := r.assertU32(0x0c); err != nil {
		return nil, errors.New("expected 0x0c before diploFavor")
	}
	r.skip(4)
	ps.DiploFavor = int(r.readU32())
	r.skip(20)

	r.readMapDiscard() // 3 maps
	r.readMapDiscard()
	r.readMapDiscard()
	r.readArrayDiscard(4, 0) // 05
	r.readMapDiscard()       // 06

	// ── Cities ───────────────────────────────────────────────────────
	r.skip(33)
	if err := r.assertU32(0x10); err != nil {
		return nil, errors.New("expected 0x10 before cityCount")
	}
	r.skip(4)
	r.skip(4)
	cityCount := int(r.readU32())
	ps.Cities = make([]CityState, cityCount)

	for iCity := 0; iCity < cityCount; iCity++ {
		if err := parseCity(r, &ps.Cities[iCity]); err != nil {
			return nil, err
		}
	}

	r.skip(4)
	count = int(r.readU32())
	r.skip(count * 4)
	r.skip(8) // 06 07
	r.skip(5) // 00
	r.skip(4) // continent
	count = int(r.readU32())
	r.skip(count * 4)
	r.skip(4)
	r.skip(8)
	count = int(r.readU32())
	r.skip(count * 4)
	r.skip(4)
	count = int(r.readU32())
	r.skip(count * 4)
	r.skip(4)
	count = int(r.readU32())
	r.skip(count * 8)

	if err := r.assertU32(0x32); err != nil {
		return nil, errors.New("expected 0x32 after continent block")
	}
	r.skip(4)
	r.skip(3137)

	r.readMapDiscard()
	r.readMapDiscard()
	r.skip(53)

	count = int(r.readU32())
	r.skip(count * 264)
	count = int(r.readU32())
	r.skip(count * 72)

	// TOOLTIP
	count = int(r.peek32())
	if count != 64 {
		return nil, errors.New("expected 64 in tooltip")
	}
	r.skip(4)
	for i := 0; i < count; i++ {
		count2 := int(r.readU32())
		for k := 0; k < count2; k++ {
			r.skip(4)
			r.skip(4)
			r.skip(33)
			r.readString()
			r.skip(4)
		}
	}
	r.skip(12)
	r.skip(64 * 16)
	r.skip(64 * 4)
	r.skip(64 * 4)
	r.skip(768)

	count = int(r.readU32())
	r.skip(count * 8)

	for i := 0; i < 64; i++ {
		if err := r.assertU32(0x32); err != nil {
			return nil, errors.New("expected 0x32 in 64-loop")
		}
		r.skip(13)
	}

	count = int(r.readU32())
	r.skip(count * 21)
	count = int(r.readU32())
	r.skip(count * 16)
	count = int(r.readU32())
	r.skip(count * 4)
	count = int(r.readU32())
	r.skip(count * 16)

	count = int(r.readU32())
	for i := 0; i < count; i++ {
		r.skip(4)
		r.skip(8)
	}
	r.skip(86)

	count = int(r.peek32())
	if count != 64 {
		return nil, errors.New("expected 64 in player bitmap 1")
	}
	r.skip(4)
	r.skip(64 * 4)
	count = int(r.peek32())
	if count != 64 {
		return nil, errors.New("expected 64 in player bitmap 2")
	}
	r.skip(4)
	r.skip(64 * 4)

	count = int(r.readU32())
	r.skip(count * 16)
	count = int(r.readU32())
	for i := 0; i < count; i++ {
		r.skip(8)
		count2 := int(r.readU32())
		r.skip(count2 * 8)
	}
	count = int(r.readU32())
	r.skip(count * 14)
	count = int(r.readU32())
	for i := 0; i < count; i++ {
		r.skip(8)
		count2 := int(r.readU32())
		r.skip(count2 * 20)
	}
	count = int(r.readU32())
	for i := 0; i < count; i++ {
		r.skip(8)
		count2 := int(r.readU32())
		for j := 0; j < count2; j++ {
			r.skip(4)
			count3 := int(r.readU32())
			r.skip(count3 * 4)
		}
	}
	count = int(r.readU32())
	for i := 0; i < count; i++ {
		r.skip(8)
		count2 := int(r.readU32())
		r.skip(count2 * 20)
	}

	r.readArrayDiscard(4, 1) // 8c
	r.readArrayDiscard(4, 0) // 40 * 0
	r.readMapDiscard()       // 2a

	r.skip(16)

	// ── Districts ────────────────────────────────────────────────────
	districtCount := int(r.readU32())
	ps.Districts = make([]DistrictState, districtCount)
	for i := 0; i < districtCount; i++ {
		if err := parseDistrict(r, &ps.Districts[i]); err != nil {
			return nil, err
		}
	}

	r.skip(4)
	count = int(r.readU32())
	r.skip(count * 4)
	r.skip(8)
	r.skip(8)

	r.readMapDiscard()
	r.readMapDiscard()
	r.readMapDiscard()
	r.readMapDiscard()
	r.readMapDiscard()

	count = int(r.readU32())
	r.skip(count * 22)
	count = int(r.readU32())
	r.skip(count * 12)
	count = int(r.readU32())
	r.skip(count * 16)
	count = int(r.readU32())
	r.skip(count * 12)

	r.skip(4)
	r.skip(4)
	count = int(r.peek32())
	r.skip(4)
	if err := r.assertU32(uint32(count)); err != nil {
		return nil, errors.New("tile count mismatch")
	}
	r.skip(4)
	r.skip(count * 2)
	r.skip(4) // 2d

	ps.Government = r.readU32()
	r.skip(8)       // FF
	_ = r.readU32() // lastTurnChangeGovernment
	r.skip(4)
	r.skip(1)
	r.readArrayDiscard(4, 1)
	r.readMapDiscard()
	r.readMapDiscard()
	for i := 0; i < 5; i++ {
		r.readArrayDiscard(4, 1)
	}

	count = int(r.readU32())
	r.skip(4)
	r.skip(count * 8)

	count = int(r.readU32())
	r.skip(count * 8)
	r.skip(4)
	r.skip(8)

	ps.CivicsResearched = r.readMapBool() // dogme found
	r.readMapBoolDiscard()                // dogme boost
	r.readMapFloat()                      // dogme research
	r.readArrayDiscard(4, 0)              // dogme current (sep=0 per C++ readArray(it, 4, 0))

	r.skip(21)
	r.skip(4)
	count = int(r.peek32())
	if count != 64 {
		return nil, errors.New("expected 64 in tourism by player")
	}
	r.skip(4)
	r.skip(64 * 4)

	count = int(r.peek32())
	if count != 64 {
		return nil, errors.New("expected 64 in touristUnknown")
	}
	r.skip(4)
	r.skip(64)
	count = int(r.peek32())
	if count != 64 {
		return nil, errors.New("expected 64 in touristByPlayer")
	}
	r.skip(4)
	r.skip(64 * 4)

	r.skip(12)
	r.skip(12)
	r.skip(4)
	r.readU32() // internalTourist
	r.skip(8)

	count = int(r.peek32())
	if count != 64 {
		return nil, errors.New("expected 64 in tourism count")
	}
	r.skip(4)
	r.skip(64 * 4)
	r.skip(8)

	count = int(r.readU32())
	var totalPerTurnTourism int
	for i := 0; i < count; i++ {
		r.skip(4)                               // id location tourism generator
		totalPerTurnTourism += int(r.readU32()) // generated tourism per turn
		count2 := int(r.readU32())
		for j := 0; j < count2; j++ {
			r.skip(4)
			r.skip(4) // externalTourist
		}
	}
	ps.Tourism = float32(totalPerTurnTourism)

	count = int(r.readU32())
	r.skip(count * 17)
	count = int(r.readU32())
	r.skip(count * 12)
	r.skip(1)

	count = int(r.readU32())
	for i := 0; i < count; i++ {
		if err := r.assertU32(0x2D); err != nil {
			return nil, errors.New("expected 0x2D in civic loop")
		}
		r.skip(12)
	}
	r.skip(4)

	count = int(r.readU32())
	for i := 0; i < count; i++ {
		if err := r.assertU32(0x2D); err != nil {
			return nil, errors.New("expected 0x2D in civic loop 2")
		}
		r.skip(13)
	}
	r.skip(16)
	r.skip(4)
	count = int(r.readU32())
	r.skip(count * 8)
	r.skip(4)
	r.skip(8)
	r.skip(8)
	count = int(r.readU32())
	r.skip(count * 13)
	count = int(r.readU32())
	r.skip(count * 16)

	// SPY
	if err := r.assertU32(0x11); err != nil {
		return nil, errors.New("expected 0x11 before spy")
	}
	r.skip(8)
	count = int(r.readU32())
	for i := 0; i < count; i++ {
		r.skip(24)
		count2 := int(r.readU32())
		r.skip(count2 * 4)
	}
	count = int(r.readU32())
	for i := 0; i < count; i++ {
		r.skip(40)
		count2 := int(r.readU32())
		r.skip(count2 * 4)
		r.skip(4)
	}
	count = int(r.readU32())
	r.skip(count * 16)
	r.skip(4)
	r.skip(4)
	r.skip(4)
	count = int(r.readU32())
	r.skip(count * 44)
	count = int(r.readU32())
	r.skip(count * 16)
	count = int(r.readU32())
	r.skip(count * 12)

	if err := r.assertU32(0x14); err != nil {
		return nil, errors.New("expected 0x14 before faith")
	}
	r.skip(4) // past 0x14
	r.skip(1) // it++
	ps.Faith = int(r.readU32())
	r.skip(3)
	r.readU32() // pantheon
	r.skip(4)
	r.skip(8)

	count = int(r.readU32())
	r.skip(count * 24)
	count = int(r.readU32())
	r.skip(count * 13)
	r.skip(4)
	r.skip(44)
	r.skip(2)
	r.skip(4)
	r.skip(20)

	r.readMapDiscard() // 15
	r.skip(16)
	r.readArrayDiscard(4, 0) // 19
	r.skip(4)                // 10
	r.readArrayDiscard(4, 1) // 36
	r.skip(13)

	// Strategic resources
	stratCount := int(r.readU32())
	for i := 0; i < stratCount; i++ {
		r.readU32() // type
		if err := r.assertU32(3); err != nil {
			return nil, errors.New("expected 3 in strat resource")
		}
		r.skip(4)
		r.readU32() // current
		r.readU32() // get
		r.readU32() // give
	}
	if err := r.assertU32(uint32(stratCount)); err != nil {
		return nil, errors.New("strat count mismatch")
	}
	r.skip(4)
	for i := 0; i < stratCount; i++ {
		r.skip(4) // type
		r.readFloat()
	}
	count = int(r.readU32())
	r.skip(count * 8)

	r.readMapDiscard()
	r.readMapDiscard()
	r.readMapDiscard()
	r.readMapDiscard()

	count = int(r.readU32())
	for i := 0; i < count; i++ {
		r.skip(4)
		r.readFloat()
	}
	r.readMapDiscard()
	r.readMapDiscard()
	count = int(r.readU32())
	for i := 0; i < count; i++ {
		r.skip(4)
		r.readU32()
	}
	r.readMapDiscard()
	r.readMapDiscard()

	r.skip(32)
	r.skip(65)
	r.readInt(3) // diploPoint
	r.skip(3)

	count = int(r.readU32())
	for i := 0; i < count; i++ {
		r.skip(13)
		r.readString()
		r.skip(22)
	}
	r.skip(20)

	r.readMapDiscard() // unitsCount 8C
	r.readMapDiscard() // 23
	count = int(r.readU32())
	r.skip(count * 4)
	r.readMapDiscard() // 06

	r.skip(1888)

	r.readArrayDiscard(4, 0) // tech current
	r.skip(12)
	ps.TechsResearched = r.readMapBool() // tech found
	r.readMapBoolDiscard()               // tech boost
	r.readMapFloat()                     // tech research

	r.skip(8)
	r.skip(4)
	r.skip(4)
	r.skip(1)
	ps.Gold = int(r.readU32())
	r.skip(4)
	r.skip(4)
	r.skip(20)
	r.skip(4)
	r.skip(3)

	r.readMapDiscard() // 3A

	count = int(r.readU32())
	r.skip(count * 8)
	r.skip(16)

	// ── Full units (detailed) ──────────────────────────────────────────
	unitCount = int(r.readU32())
	for iUnit := 0; iUnit < unitCount; iUnit++ {
		if err := parseDetailedUnit(r); err != nil {
			return nil, err
		}
	}

	r.skip(4)
	count = int(r.readU32())
	r.skip(count * 4)
	r.skip(8) // 06 07

	r.readMapDiscard() // 8C
	r.readMapDiscard()
	r.readMapDiscard()
	r.readMapDiscard()       // 15
	r.readArrayDiscard(4, 1) // 8C

	r.skip(3)
	count = int(r.readU32())
	for i := 0; i < count; i++ {
		r.skip(4)
		count2 := int(r.readU32())
		r.skip(count2 * 8)
	}
	count = int(r.readU32())
	r.skip(count * 4) // traders
	count = int(r.readU32())
	r.skip(count * 4)
	r.skip(4)
	// hack: conditional skip
	if r.peek32() == 0 {
		r.skip(4)
	}
	r.readMapDiscard() // 8C
	r.skip(12)
	count = int(r.readU32())
	r.skip(count * 12)
	count = int(r.readU32())
	r.skip(count * 12)
	count = int(r.readU32())
	r.skip(count * 12)
	count = int(r.readU32())
	r.skip(count * 4)
	r.skip(32 + 44)
	r.readMapDiscard()       // 8C
	r.readMapDiscard()       // 36
	r.readArrayDiscard(4, 0) // 02 eras
	r.skip(4 + 4)
	r.readU32() // availableRoads
	r.readU32() // unknownRoads
	r.skip(52)
	for i := 0; i < 2; i++ {
		r.skip(4 + 64)
	}
	for i := 0; i < 13; i++ {
		if err := r.assertU32(6); err != nil {
			return nil, errors.New("expected 6 in 13-map loop post-units")
		}
		r.readMapDiscard()
	}
	r.skip(4)
	for i := 0; i < 72; i++ {
		if err := r.assertU32(6); err != nil {
			return nil, errors.New("expected 6 in 72-map loop post-units")
		}
		r.readMapDiscard()
	}
	// improvement yields
	for k := 0; k < 3; k++ {
		count = int(r.readU32())
		for i := 0; i < count; i++ {
			r.skip(4)
			if err := r.assertU32(6); err != nil {
				return nil, errors.New("expected 6 in improvement yields")
			}
			r.readMapDiscard()
		}
	}

	// terrain yields
	r.readMapDiscard() // 06
	for k := 0; k < 4; k++ {
		count = int(r.readU32())
		for i := 0; i < count; i++ {
			r.skip(4)
			if err := r.assertU32(6); err != nil {
				return nil, errors.New("expected 6 in terrain yields")
			}
			r.readMapDiscard()
		}
	}
	r.skip(12 + 8)
	gpCurrent := r.readMapFloat() // greatPeopleCurrent
	gpProd := r.readMapFloat()    // greatPeopleProd
	r.readMapFloat()              // greatPeopleUnknown

	for _, v := range gpCurrent {
		if v > 0 {
			ps.GreatPeopleCurrent += v
		}
	}
	for _, v := range gpProd {
		if v > 0 {
			ps.GreatPeoplePerTurn += v
		}
	}

	// NOTE: greatPeopleCurrent is progress toward current recruit, not count of
	// recruited great people. Counting those values inflates score massively.
	// GreatPeopleRecruited stays 0 unless parsed from the dedicated recruited list.
	r.skip(10)
	r.readMapDiscard()
	r.readMapDiscard()
	r.readMapDiscard()
	r.skip(4)
	// Scan forward for the skip41 boundary: 0x16000000 followed by 64 at +41 bytes
	// This is more reliable than parsing the variable great-person class loops
	{
		found := false
		for off := 0; off < 2000; off++ {
			if r.pos+off+45 > len(r.data) {
				break
			}
			if binary.LittleEndian.Uint32(r.data[r.pos+off:]) == 0x16 &&
				binary.LittleEndian.Uint32(r.data[r.pos+off+41:]) == 64 {
				r.pos = r.pos + off + 41
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("player %d: skip41 anchor not found", trueiPlayer)
		}
	}

	// emissaries
	count = int(r.readU32())
	for i := 0; i < count; i++ {
		r.readU32()
	}
	count = int(r.readU32())
	r.skip(count * 4)
	count = int(r.readU32())
	r.skip(count) // 1 byte each

	r.readMapDiscard()
	r.skip(16)
	r.readMapDiscard()
	r.readMapDiscard()
	count = int(r.readU32())
	r.skip(count * 4)
	count = int(r.readU32())
	r.skip(count * 4)
	count = int(r.readU32())
	r.skip(count * 4)
	count = int(r.readU32())
	r.skip(count * 4)
	r.skip(20)
	count = int(r.readU32())
	r.skip(count * 4)
	count = int(r.readU32())
	r.skip(count * 4)
	r.skip(12)
	r.readMapDiscard()

	// governors
	r.skip(4)
	count = int(r.readU32())
	r.skip(count * 4)
	r.skip(8 + 16)
	govCount := int(r.readU32())
	for i := 0; i < govCount; i++ {
		r.skip(4 + 4 + 4) // id + 0F + i+1
		r.readU32()       // type
		r.readString()    // name
		r.skip(2 + 2 + 2 + 2 + 62 + 4 + 4)
		count2 := int(r.readU32())
		r.skip(count2 * 12)
		r.readMapDiscardSV(1) // promotions
		r.skip(4 + 1)
	}

	// post-governor section
	r.skip(4)
	r.readArrayDiscard(4, 0)
	r.skip(8 + 4 + 1 + 8 + 64 + 4 + 4)
	count = int(r.peek32())
	if count != 64 {
		return nil, errors.New("expected 64 in gov post bitmap a")
	}
	r.skip(4 + 64*4)
	count = int(r.peek32())
	if count != 64 {
		return nil, errors.New("expected 64 in gov post bitmap b")
	}
	r.skip(4 + 64*4)
	r.skip(8)
	r.readArrayDiscard(4, 1) // 8C
	r.readArrayDiscard(4, 1) // 46
	r.readArrayDiscard(4, 1) // 2B
	r.skip(24)
	count = int(r.readU32())
	r.skip(count * 20)
	r.skip(4)
	count = int(r.readU32())
	r.skip(count * 4)
	r.skip(8 + 2)

	// variable-type count loop
	count = int(r.readU32())
	for i := 0; i < count; i++ {
		tp := int(r.readU32())
		if err := r.assertU32(3); err != nil {
			return nil, errors.New("expected 3 in type loop")
		}
		r.skip(4)
		switch tp {
		case 0:
			r.skip(61)
		case 1:
			r.skip(48)
		case 2, 3, 4:
			r.skip(44)
		default:
			return nil, errors.New("unknown type in type loop")
		}
	}

	if err := r.assertU32(2); err != nil {
		return nil, errors.New("expected 2 before tileCount in yields")
	}
	r.skip(4)
	tileCount := int(r.readU32())
	r.skip(tileCount * 2)
	if err := r.assertU32(uint32(tileCount)); err != nil {
		return nil, errors.New("tile count mismatch in yields")
	}
	r.skip(4 + tileCount*2 + 24 + 4)

	if err := parseCityYields(r, ps, trueiPlayer); err != nil {
		return nil, err
	}

	// post city-yield section

	// post city yields: skip4 + c*4 + skip(8+4+4) + c*12 + skip8
	r.skip(4)
	count = int(r.readU32())
	r.skip(count * 4)
	r.skip(8 + 4 + 4)
	count = int(r.readU32())
	r.skip(count * 12)
	r.skip(8)
	if err := r.assertU32(3); err != nil {
		return nil, fmt.Errorf("player %d: expected 3, got %d at pos %d", trueiPlayer, r.peek32(), r.pos)
	}
	r.skip(4)
	count = int(r.readU32())
	r.skip(count * 8)
	r.skip(1)
	r.readMapDiscard() // map35a
	r.readMapDiscard() // map35b

	// Line 2811: readMap //05
	r.readMapDiscard()

	// Lines 2813-2828: tech dogme turnTo
	count = int(r.readU32())
	for i := 0; i < count; i++ {
		key := r.readU32()
		if err := r.assertU32(2); err != nil {
			return nil, errors.New("expected 2 in tech dogme turnTo")
		}
		r.skip(4) // advance past 2
		r.skip(4) // extra
		if err := r.assertU32(key); err != nil {
			return nil, errors.New("key mismatch in tech dogme turnTo")
		}
		r.skip(4) // advance past key
		r.skip(4) // extra
		r.skip(4) // extra
		r.skip(8)
		r.readU32() // turnCount
	}

	// Lines 2830-2835: count*34
	count = int(r.readU32())
	r.skip(count * 34)

	// Lines 2837-2842: count*32
	count = int(r.readU32())
	r.skip(count * 32)

	// Line 2843: readMap //0D
	r.readMapDiscard()

	// Lines 2846-2860: improvements count*(4+4+4+x+y+district+type+1)
	count = int(r.readU32())
	r.skip(count * (4 + 4 + 4 + 4 + 4 + 4 + 4 + 1))

	// Line 2862: skip8
	r.skip(8)
	// Lines 2864-2869: count*8, skip4
	count = int(r.readU32())
	r.skip(count * 8)
	r.skip(4)
	// Lines 2872-2879: count*4, skip1, skip4
	count = int(r.readU32())
	r.skip(count * 4)
	r.skip(1 + 4)

	// Line 2881: skip1844
	r.skip(1844)

	// Now search END_IA_STUFF — log position for debugging — we're now past the fixed sections
	// and into the diplomatic state + skip593 zone
	idxIA := bytes.Index(r.data[r.pos:], endIAStuff)
	if idxIA == -1 {
		return nil, errors.New("END_IA_STUFF anchor not found")
	}
	r.pos = r.pos + idxIA - 4
	r.readMapDiscard() // 23

	count = int(r.readU32())
	r.skip(count * 13)
	count = int(r.readU32())
	r.skip(count * 16)
	count = int(r.readU32())
	r.skip(count * 4)

	// C++: count2 = readInt(3); it+=4; if count2!=0: search END_PLAYER_STUFF
	count2Raw := r.readInt(3)
	r.skip(1)
	if count2Raw != 0 {
		idxPS := bytes.Index(r.data[r.pos:], endPlayerStuff)
		if idxPS == -1 {
			return nil, errors.New("END_PLAYER_STUFF anchor not found")
		}
		r.pos = r.pos + idxPS - 4
	}

	// C++: readMap(//3A); count*4; readArray(4,0)*3
	r.readMapDiscard() // 3A
	count = int(r.readU32())
	r.skip(count * 4)
	r.readArrayDiscard(4, 0) // 11
	r.readArrayDiscard(4, 0) // 25
	r.readArrayDiscard(4, 0) // 25

	// compute player totals
	for ci := range ps.Cities {
		c := &ps.Cities[ci]
		ps.Science += c.Science
		ps.Culture += c.Culture
		ps.Food += c.Food
		ps.Production += c.Production
		ps.Tourism += c.Tourism
	}

	return ps, nil
}

// parseCity reads one city block within a player section.
func parseCity(r *reader, city *CityState) error {
	city.ID = int(r.readU32())
	if err := r.assertU32(0x33); err != nil {
		return errors.New("expected 0x33 in city")
	}
	r.skip(4)
	r.skip(4)
	city.X = int(r.readU32())
	city.Y = int(r.readU32())

	r.skip(12) // 0
	r.skip(4)  // FF
	r.skip(4)  // 00 00 01 00
	city.Population = int(r.readU32())
	r.skip(2) // 01 01
	r.skip(4) // random int
	r.skip(8) // FF
	r.skip(4) // 1F / FF

	r.skip(64 * 4)
	r.skip(1)
	r.skip(18)

	for i := 0; i < 13; i++ {
		if err := r.assertU32(6); err != nil {
			return errors.New("expected 6 in city readMap loop")
		}
		r.readMapDiscard()
	}

	count := int(r.readU32())
	r.skip(count * 20)

	count = int(r.readU32())
	r.skip(count * 16)

	count = int(r.readU32())
	r.skip(count * 8)

	count = int(r.readU32())
	r.skip(count * 12)

	r.skip(8)
	count = int(r.readU32())
	for i := 0; i < count; i++ {
		if err := r.assertU32(0x33); err != nil {
			return errors.New("expected 0x33 in city district sub-loop")
		}
		r.skip(4)
		r.skip(4)
		r.skip(8)
		r.skip(4)
		r.skip(6)
		r.skip(4)
		r.skip(8)
		r.readString() // LOC_DISTRICT_...
	}

	count = int(r.readU32())
	r.skip(count * 16)

	r.skip(17)
	if err := r.assertU32(6); err != nil {
		return errors.New("expected 6 before city name readMap")
	}
	r.readMapDiscard()
	r.skip(13)
	city.Name = r.readString()
	city.Name = normaliseCityName(city.Name)

	r.skip(8)
	r.skip(4)
	count = int(r.peek32())
	r.skip(4) // other city?
	r.skip(count * 4)

	r.skip(8)
	r.readArrayDiscard(4, 0) // 25
	r.readArrayDiscard(4, 0) // 25
	r.readArrayDiscard(4, 1) // 06
	r.readArrayDiscard(4, 1) // 06
	r.skip(9)
	r.skip(36)

	count = int(r.readU32())
	r.skip(count * 22)
	count = int(r.readU32())
	r.skip(count * 20)

	r.skip(4)
	r.skip(25)

	r.readMapDiscard() // 8c
	r.readMapDiscard() // 85
	r.readMapDiscard() // 22
	r.readMapDiscard() // 11
	r.skip(4)
	r.readMapDiscard() // 85
	r.readMapDiscard() // 8c
	r.skip(5)
	r.readArrayDiscard(4, 0) // 17
	r.skip(33)

	count = int(r.readU32())
	r.skip(count * 8) // jungle etc

	r.skip(4)
	city.Religion = r.readU32()
	count = int(r.readU32())
	city.ReligionFollowers = make(map[uint32]uint32, count)
	for i := 0; i < count; i++ {
		r.skip(4) // marker-like field (often 0x05)
		religionHash := r.readU32()
		r.skip(4)             // unknown
		r.skip(1)             // unknown byte
		amount := r.readU32() // follower-like amount
		city.ReligionFollowers[religionHash] = amount
	}
	r.skip(8)
	r.skip(18)
	r.skip(4)
	r.skip(16)

	r.readMapDiscard() // 06
	r.readMapDiscard() // 06
	r.skip(12)
	r.readMapDiscard() // 11
	r.readMapDiscard() // 36

	r.skip(4)
	count = int(r.peek32())
	r.skip(4) // productions
	for i := 0; i < count; i++ {
		if err := r.assertU32(0x2C0F4A46); err != nil {
			return errors.New("expected 0x2C0F4A46 in production")
		}
		r.skip(4) // marker (consumed by assert peek, now advance)
		r.skip(4) // 02
		r.skip(4) // reversed marker
		prodType := int(r.readU32())
		r.skip(4)
		if prodType > 4 {
			return errors.New("unexpected prodType")
		}
		r.readU32()        // currentProd
		if prodType == 0 { // PROD_UNIT
			r.skip(4)
		}
		r.skip(12)
	}
	r.skip(40)

	r.readMapDiscard()       // 46
	r.readMapDiscard()       // 46
	city.Prod = r.readMap(4) // 85 prod
	r.readMapDiscardSV(2)    // 85
	r.readMapDiscard()       // 8C prod
	r.readMapDiscard()       // 22 prod
	r.readMapDiscard()       // 23 prod
	r.readMapDiscardSV(2)    // 22
	r.readMapDiscard()       // 85
	r.readMapDiscard()       // 8C
	r.readMapDiscard()       // 22
	r.readMapDiscard()       // 8C
	r.readArrayDiscard(4, 1) // 46

	r.skip(8)
	count = int(r.readU32())
	r.skip(count * 12)
	r.readMapDiscard() // 8C
	r.readMapDiscard() // 23
	r.skip(9)
	count = int(r.readU32())
	r.skip(count * 12)
	r.skip(4)
	city.Built = r.readMap(2) // 85 built

	// Detect built wonders: in this map 0xFFFF appears to mean "not built".
	for crc, val := range city.Built {
		if val != 0xFFFF {
			if name, ok := WonderCRCs[crc]; ok {
				city.Wonders = append(city.Wonders, name)
			}
		}
	}

	r.readArrayDiscard(4, 1) // 85

	r.readMapDiscardSV(2) // 85
	r.readMapDiscardSV(2) // 85

	count = int(r.readU32())
	r.skip(count * 12) // amphi/museum
	count = int(r.readU32())
	r.skip(count * 12)
	count = int(r.readU32())
	r.skip(count * 12)
	count = int(r.readU32())
	r.skip(count * 8) // YIELD_FAITH
	count = int(r.readU32())
	for i := 0; i < count; i++ {
		r.skip(4)
		count2 := int(r.readU32())
		r.skip(count2 * 12)
	}
	count = int(r.readU32())
	r.skip(count * 12)
	count = int(r.readU32())
	r.skip(count * 16)
	count = int(r.readU32())
	r.skip(count * 4)

	r.skip(4) // 07
	r.skip(37)

	count = int(r.peek32())
	if count != 64 {
		return errors.New("expected 64 in city player bitmap")
	}
	r.skip(4)
	r.skip(64)
	r.skip(4)

	for i := 0; i < 14; i++ {
		if err := r.assertU32(6); err != nil {
			return errors.New("expected 6 in 14-map loop")
		}
		r.readMapDiscard()
	}
	r.skip(4) // 0e
	r.skip(1)
	r.skip(4) // 64
	r.skip(3) // 1D 00 00

	count = int(r.readU32())
	r.skip(count * 13)
	r.skip(12)
	r.skip(49)
	r.skip(12)
	r.skip(4)

	count = int(r.readU32())
	r.skip(count * 24)
	count = int(r.readU32())
	r.skip(count * 20)
	count = int(r.readU32())
	r.skip(count * 16)
	count = int(r.readU32())
	r.skip(count * 32)
	r.skip(4)

	// PLOT_PROPERTIES
	count = int(r.readInt(3))
	r.skip(1)
	for i := 0; i < count; i++ {
		r.skip(8)
		count2 := int(r.readInt(2))
		r.skip(count2)
		r.skip(12)
		count3 := int(r.readInt(3))
		r.skip(1)
		for j := 0; j < count3; j++ {
			count4 := int(r.readU32())
			r.skip(count4 * 22)
		}
	}

	r.skip(8)
	count = int(r.readU32())
	r.skip(count * 4)

	r.readArrayDiscard(4, 0) // 8c
	r.readMapDiscard()       // 06
	return nil
}

// parseDistrict reads one district instance within a player section.
func parseDistrict(r *reader, d *DistrictState) error {
	d.GlobalID = int(r.readU32())
	if err := r.assertU32(0x0f); err != nil {
		return errors.New("expected 0x0f in district")
	}
	r.skip(4)
	d.ID = int(r.readInt(2))
	r.skip(2)
	d.X = int(r.readU32())
	d.Y = int(r.readU32())
	d.CityID = int(r.readInt(2))
	r.skip(2)
	d.Type = r.readU32()
	d.Damage = int(r.readU32())
	_ = int(r.readU32()) // wallDamage
	r.skip(4)            // 9C FF FF FF
	d.Wall = int(r.readU32())
	_ = int(r.readU32()) // cost
	d.Built = int(r.readU32())
	r.skip(1)
	for k := 0; k < 3; k++ {
		if err := r.assertU32(6); err != nil {
			return errors.New("expected 6 in district map loop")
		}
		r.readMapDiscard()
	}
	r.skip(12)
	if err := r.assertU32(6); err != nil {
		return errors.New("expected 6 in district map 4")
	}
	r.readMapDiscard()
	if err := r.assertU32(6); err != nil {
		return errors.New("expected 6 in district map 5")
	}
	r.readMapDiscard()
	r.skip(4)
	count2 := int(r.readU32())
	r.skip(count2 * 16)
	r.readMapDiscard() // 0A
	r.skip(17)
	d.Pillage = int(r.readInt(1))
	return nil
}

// parseDetailedUnit skips over one entry of the detailed per-player unit list.
// Nothing is currently extracted from it.
func parseDetailedUnit(r *reader) error {
	r.readU32() // id
	r.readU32() // count?
	r.skip(4)
	r.readU32() // type
	r.skip(4)
	r.skip(4) // FF FF FF FF
	r.skip(4) // F1 D8 FF FF
	r.skip(4)
	r.readU32() // x
	r.readU32() // y
	r.skip(17)
	r.readU32() // army
	r.skip(8)
	r.readU32() // damage
	r.skip(58)
	r.readU32() // fortified
	r.skip(11)
	r.skip(199)
	r.skip(4 * 4)
	r.skip(4) // 00 with maj
	r.skip(1) // new magic firaxis byte
	count := int(r.readU32())
	r.skip(count * 8)
	r.readMapDiscard()       // C9
	r.readArrayDiscard(4, 1) // 11
	r.readArrayDiscard(4, 1) // 11
	r.readArrayDiscard(4, 1) // 46
	r.readMapDiscard()       // 0A
	r.readMapDiscard()       // 0A
	count = int(r.readU32())
	r.skip(count * 5) // promotions 1
	count = int(r.readU32())
	r.skip(count * 5) // promotions 2
	r.skip(6)
	r.readU32() // xp
	r.readU32() // level
	r.skip(12)
	r.readArrayDiscard(4, 1) // 91
	r.readString()
	r.skip(77)
	count = int(r.readU32())
	r.skip(count * 12)
	r.skip(54)
	if err := r.assertU32(6); err != nil {
		return errors.New("expected 6 in unit map")
	}
	r.readMapDiscard() // 06
	count = int(r.readU32())
	r.skip(count * 4)
	r.readMapDiscard() // 22
	r.readMapDiscard() // 3A
	r.readMapDiscard() // 22
	r.readMapDiscard() // 3A
	r.skip(66)
	count = int(r.readU32())
	r.skip(count * 13)
	count = int(r.readU32())
	r.skip(count * 12)
	r.skip(70)
	// unit operations - variable size per type
	count = int(r.readU32())
	for i := 0; i < count; i++ {
		op := r.peek32()
		switch op {
		case 0x580F2F68, 0xb2cca377, 0x9C0B44C6, 0x09d0292a,
			0x886ffcd1, 0x7fa205d1, 0xcfb9b561, 0xc8ce5dfb,
			0x1f633b1e, 0x852ce4df:
			r.skip(44)
		case 0x8374D954:
			r.skip(56)
		case 0x1d60e778, 0x4885d724, 0x8ca367f, 0x6e68aef:
			r.skip(40)
		case 0x98eca9ea:
			r.skip(48)
		default:
			r.skip(32)
		}
	}
	r.readArrayDiscard(4, 1) // 3F
	r.skip(16)
	// plot properties
	count = int(r.readInt(3))
	r.skip(1)
	for i := 0; i < count; i++ {
		count2 := int(r.readU32())
		for k := 0; k < count2; k++ {
			r.skip(4)
			r.readStringN(2)
			r.skip(16)
		}
	}
	return nil
}

// parseCityYields reads the per-city yield block and assigns the yields
// to the matching cities on ps.
func parseCityYields(r *reader, ps *PlayerState, trueiPlayer int) error {
	cityYieldCount := int(r.readU32())
	for i := 0; i < cityYieldCount; i++ {
		cityID := int(r.readU32())
		if err := r.assertU32(0x11); err != nil {
			return fmt.Errorf("player %d city %d: expected 0x11 at pos %d, got %08x", trueiPlayer, i, r.pos, r.peek32())
		}
		r.skip(4 + 4 + 4 + 4 + 4 + 4) // skip 0x11 + 5 more
		if err := r.assertU32(6); err != nil {
			return fmt.Errorf("player %d city %d: expected 6 before yields at pos %d, got %d", trueiPlayer, i, r.pos, r.peek32())
		}
		yields := r.readMapFloat()
		if err := r.assertU32(6); err != nil {
			return fmt.Errorf("player %d city %d: expected 6 before yields2 at pos %d, got %d", trueiPlayer, i, r.pos, r.peek32())
		}
		r.readMapFloat()   // yields2
		r.readMapDiscard() // 35
		r.skip(8 + 9)
		count2 := int(r.readU32())
		r.skip(count2 * 16) // 25
		count2 = int(r.readU32())
		r.skip(count2 * 4) // 6
		r.skip(12)
		count2 = int(r.readU32())
		r.skip(count2 * 12) // 0 or 3
		r.skip(1 + 16)
		count2 = int(r.readU32())
		r.skip(count2 * 4)
		count2 = int(r.readU32())
		r.skip(4)
		r.skip(count2 * 8)
		// outer 04/64 loop
		count2 = int(r.readU32())
		for j := 0; j < count2; j++ {
			if err := r.assertU32(0x04); err != nil {
				return errors.New("expected 0x04 in yield sub-loop")
			}
			r.skip(4 + 4)
			if err := r.assertU32(0x64); err != nil {
				return errors.New("expected 0x64 in yield sub-loop")
			}
			r.skip(4 + 35)
			if err := r.assertU32(0x40); err != nil {
				return errors.New("expected 0x40 in yield sub-loop")
			}
			count3 := int(r.readU32())
			r.skip(count3 * 4)
			r.skip(12)
		}
		r.skip(20)
		count2 = int(r.readU32())
		r.skip(count2 * 17)
		count2 = int(r.readU32())
		r.skip(count2 * 8)
		r.skip(12 + 16)
		// c2*69 loop: assert1 + 4 + 4 + 61
		count2 = int(r.readU32())
		for j := 0; j < count2; j++ {
			if err := r.assertU32(1); err != nil {
				return errors.New("expected 1 in c2*69 loop")
			}
			r.skip(4 + 4 + 61)
		}
		// c2 * 3maps
		count2 = int(r.readU32())
		for j := 0; j < count2; j++ {
			r.readMapDiscard()
			r.readMapDiscard()
			r.readMapDiscard()
		}
		r.readMapDiscard()
		r.readMapDiscard()
		r.readMapDiscard()
		r.skip(8)
		r.readMapDiscard()

		// assign yields to the city
		for ci := range ps.Cities {
			if ps.Cities[ci].ID == cityID {
				c := &ps.Cities[ci]
				c.Food = yields[YieldFood]
				c.Production = yields[YieldProduction]
				c.Gold = yields[YieldGold]
				c.Science = yields[YieldScience]
				c.Culture = yields[YieldCulture]
				c.Faith = yields[YieldFaith]
				c.Tourism = yields[YieldTourism]
			}
		}
	}
	return nil
}

func parseGreatPeopleRecruitedFromTail(data []byte, start int, playerCount int) (recruited []GreatPersonRecruit, err error) {
	defer func() {
		if rec := recover(); rec != nil {
			err = fmt.Errorf("tail: panic while parsing recruited GP: %v", rec)
			recruited = nil
		}
	}()

	r := newReader(data, start)

	_ = r.readU32() // livePlayerCount (not needed)
	for i := 0; i < playerCount; i++ {
		if err := r.assertU32(0x04); err != nil {
			return nil, errors.New("tail: expected 0x04 in live-player block")
		}
		r.skip(4) // 0x04
		r.skip(4) // player id
		if err := r.assertU32(0x02); err != nil {
			return nil, errors.New("tail: expected 0x02 in live-player block")
		}
		r.skip(4) // 0x02
		count := int(r.readU32())
		for j := 0; j < count; j++ {
			if err := r.assertU32(0x03); err != nil {
				return nil, errors.New("tail: expected 0x03 in live-player sub-block")
			}
			r.skip(64)
		}
		r.skip(61)
	}

	r.skip(4)
	count2 := int(r.readU32())
	for i := 0; i < count2; i++ {
		r.skip(14)
		count := int(r.readU32())
		for j := 0; j < count; j++ {
			r.skip(4 + 8 + 4 + 9)
		}
		count = int(r.readU32())
		for j := 0; j < count; j++ {
			r.skip(12)
		}
		r.skip(13)
		count = int(r.readU32())
		for j := 0; j < count; j++ {
			r.skip(41)
			count3 := int(r.readU32())
			r.skip(count3 * 4)
		}
		r.readMapDiscard()
	}

	r.skip(24)
	if err := r.assertU32(0x06); err != nil {
		return nil, errors.New("tail: expected 0x06")
	}
	r.readMapDiscard()
	r.readMapDiscard()
	r.skip(28)
	r.skip(1)
	if err := r.assertU32(0x40); err != nil {
		return nil, errors.New("tail: expected 0x40")
	}
	count := int(r.readU32())
	r.skip(count * 4)

	count = int(r.readU32())
	for i := 0; i < count; i++ {
		r.skip(16)
		r.skip(65)
		count2 = int(r.readU32())
		for j := 0; j < count2; j++ {
			r.readString()
		}
		r.skip(12)
		count2 = int(r.readU32())
		r.skip(count2 * 8)
		count2 = int(r.readU32())
		r.skip(count2 * 8)
		r.skip(8)
		count2 = int(r.readU32())
		r.skip(count2 * 20)
		count2 = int(r.readU32())
		r.skip(count2 * 4)
		r.skip(8)
		if err := r.assertU32(0x40); err != nil {
			return nil, errors.New("tail: expected 0x40 a")
		}
		count2 = int(r.readU32())
		r.skip(count2 * 4)
		if err := r.assertU32(0x40); err != nil {
			return nil, errors.New("tail: expected 0x40 b")
		}
		count2 = int(r.readU32())
		for j := 0; j < count2; j++ {
			r.skip(24)
			count3 := int(r.readU32())
			r.skip(count3 * 4)
			r.skip(20)
		}
		count2 = int(r.readU32())
		r.skip(count2 * 8)
	}

	count = int(r.readU32())
	r.skip(count * 8)
	r.skip(4)
	count2 = int(r.readU32())
	r.skip(count2 * 16)
	r.skip(4)
	r.skip(12)
	count2 = int(r.readU32())
	for i := 0; i < count2; i++ {
		r.skip(37)
		r.readArrayDiscard(4, 1)
		r.readArrayDiscard(4, 1)
	}
	r.skip(4)
	r.skip(5)

	r.skip(21)
	count2 = int(r.readU32())
	for i := 0; i < count2; i++ {
		r.skip(16)
		count = int(r.readU32())
		for j := 0; j < count; j++ {
			r.skip(28)
			r.skip(20)
			r.skip(20)
			if err := r.assertU32(0x06); err != nil {
				return nil, errors.New("tail: expected 0x06 in deep gp prelude")
			}
			r.readMapDiscard()
			r.readMapDiscard()
			r.skip(8)
			r.skip(1)
			count3 := int(r.readU32())
			r.skip(count3 * 4)
			r.skip(37)
		}
	}
	r.skip(8)
	r.skip(1)

	count2 = int(r.readU32())
	recruited = make([]GreatPersonRecruit, 0, count2)
	for i := 0; i < count2; i++ {
		if err := r.assertU32(0x05); err != nil {
			return nil, errors.New("tail: expected 0x05 at recruited gp entry")
		}
		r.skip(4)
		recruited = append(recruited, GreatPersonRecruit{
			Name:   r.readU32(),
			Class:  r.readU32(),
			Era:    r.readU32(),
			Cost:   r.readU32(),
			Player: int(r.readU32()),
		})
		r.skip(4)
	}
	return recruited, nil
}

// ── utilities ─────────────────────────────────────────────────────────────────

func normaliseCityName(s string) string {
	isLoc := false
	if strings.HasPrefix(s, "LOC_CITY_NAME_") {
		s = s[14:]
		isLoc = true
	} else if strings.HasPrefix(s, "LOC_CITY_") {
		s = s[9:]
		isLoc = true
	} else if strings.Contains(s, "_") {
		// Old parsed games might have stored "Rio_de_janeiro".
		isLoc = true
	}

	if !isLoc {
		return s
	}

	words := strings.Split(s, "_")
	for i, w := range words {
		if len(w) == 0 {
			continue
		}
		w = strings.ToLower(w)
		words[i] = strings.ToUpper(w[:1]) + w[1:]
	}
	return strings.Join(words, " ")
}
