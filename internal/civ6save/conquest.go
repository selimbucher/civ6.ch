package civ6save

// Conquest (domination) victory detection.
//
// A conquest victory is achieved when a single team controls every major
// civ's original capital. Capitals cannot be razed, so each civ's original
// capital always exists as a city somewhere — owned by the founding civ if
// never lost, or by its conqueror otherwise. We locate it via the founding
// civ's start plot (PlayerState.StartX/Y) and the per-city OriginalOwner,
// which survives capture.

// CapitalControl reports who currently controls one major civ's original capital.
type CapitalControl struct {
	Civ       int    // founding (original-owner) player index
	CityName  string // current name of that capital city
	Owner     int    // player index currently holding it, or -1 if not found
	OwnerTeam int    // team currently holding it, or -1 if not found
}

// CapitalControls returns, for every major civ in players, who currently
// controls that civ's original capital. Order follows players.
func CapitalControls(players []Player, gs *GameState) []CapitalControl {
	teamOf := make(map[int]int, len(players))
	for _, p := range players {
		teamOf[p.Index] = p.Team
	}

	out := make([]CapitalControl, 0, len(players))
	for _, p := range players {
		owner, name := originalCapitalOwner(p.Index, gs)
		cc := CapitalControl{Civ: p.Index, CityName: name, Owner: owner, OwnerTeam: -1}
		if t, ok := teamOf[owner]; ok {
			cc.OwnerTeam = t
		}
		out = append(out, cc)
	}
	return out
}

// DetectConquestWinner returns the team that controls every major civ's
// original capital, or -1 if no single team does (no conquest victory).
//
// It is deliberately conservative: if any capital cannot be attributed to a
// known team, it returns -1 rather than guessing.
func DetectConquestWinner(players []Player, gs *GameState) int {
	if gs == nil || len(players) == 0 {
		return -1
	}

	winner := -1
	for _, cc := range CapitalControls(players, gs) {
		if cc.OwnerTeam < 0 {
			return -1 // unattributable capital — don't claim a victory
		}
		if winner == -1 {
			winner = cc.OwnerTeam
		} else if winner != cc.OwnerTeam {
			return -1 // capitals split across teams — game not decided
		}
	}
	return winner
}

// originalCapitalOwner finds the city founded by civ that lies closest to
// civ's start plot, and returns its current owner (the index of the player
// whose Cities list contains it) and current name. Returns -1 if none found.
//
// The start plot is the settler's origin; the founded capital sits on or
// within a tile of it, so nearest-to-start reliably picks the original
// capital even when the civ has multiple captured cities.
func originalCapitalOwner(civ int, gs *GameState) (owner int, name string) {
	sp := gs.Players[civ]
	if sp == nil {
		return -1, ""
	}
	owner, name = -1, ""
	best := 1 << 30
	for idx, ps := range gs.Players {
		if ps == nil {
			continue
		}
		for _, c := range ps.Cities {
			if c.OriginalOwner != civ {
				continue
			}
			d := abs(c.X-sp.StartX) + abs(c.Y-sp.StartY)
			if d < best {
				best, owner, name = d, idx, c.Name
			}
		}
	}
	return owner, name
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
