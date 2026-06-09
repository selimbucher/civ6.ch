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
