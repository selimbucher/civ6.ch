package civ6save

// ReligionBySymbol returns the religion with the given symbol id.
func (gs *GameState) ReligionBySymbol(symbol uint32) *ReligionState {
	for i := range gs.Religions {
		if gs.Religions[i].Symbol == symbol {
			return &gs.Religions[i]
		}
	}
	return nil
}

// FoundedReligion returns the religion founded by the given player index, or
// nil if that player founded none.
func FoundedReligion(gs *GameState, playerIndex int) *ReligionState {
	if gs == nil {
		return nil
	}
	for i := range gs.Religions {
		r := &gs.Religions[i]
		if r.FounderPlayer == playerIndex && r.Symbol != 0 && r.Symbol != noReligion {
			return r
		}
	}
	return nil
}

// FoundedReligionFields returns the founded religion's custom name (may be empty
// for un-renameable religions), its icon key, and its display colour hex, ready
// for storage. All three are empty when the player founded no religion.
func FoundedReligionFields(gs *GameState, playerIndex int) (name, icon, color string) {
	rel := FoundedReligion(gs, playerIndex)
	if rel == nil {
		return "", "", ""
	}
	return rel.Name, ReligionIconKey(rel.Symbol), rel.ColorHex()
}
