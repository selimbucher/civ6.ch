package civ6save

// Religious (conversion) victory detection.
//
// A religious victory is achieved when one religion is "predominant" in every
// living major civ outside its founder's team. A religion is predominant in a
// civ when it is the majority religion of more than half of that civ's cities
// that follow a religion at all — cities with no majority religion (0 /
// 0xFFFFFFFF) are excluded from the count, matching the game's own check.
// City-states and eliminated civs (no cities) do not count, and teammates
// share the victory rather than needing conversion.

// ReligionStanding reports how one religion stands in one major civ.
type ReligionStanding struct {
	Civ             int  // major civ being measured
	Cities          int  // that civ's total city count
	ReligiousCities int  // cities that have a majority religion at all
	Following       int  // cities whose majority religion is this one
	Predominant     bool // Following*2 > ReligiousCities (strict majority of religious cities)
}

const noReligion = 0xFFFFFFFF

// ReligionStandings returns, for the religion with the given symbol, its
// standing in every living major civ that is not on excludeTeam. Civs with no
// cities are skipped. Order follows players.
func ReligionStandings(players []Player, gs *GameState, symbol uint32, excludeTeam int) []ReligionStanding {
	out := make([]ReligionStanding, 0, len(players))
	for _, p := range players {
		if p.Team == excludeTeam {
			continue
		}
		ps := gs.Players[p.Index]
		if ps == nil || len(ps.Cities) == 0 {
			continue // eliminated / no cities — not part of the requirement
		}
		following, religious := 0, 0
		for _, c := range ps.Cities {
			// Cities with no majority religion (0 / 0xFFFFFFFF) are excluded from
			// the denominator — the game only weighs cities that actually follow a
			// religion when deciding predominance.
			if c.Religion == 0 || c.Religion == noReligion {
				continue
			}
			religious++
			if c.Religion == symbol {
				following++
			}
		}
		out = append(out, ReligionStanding{
			Civ:             p.Index,
			Cities:          len(ps.Cities),
			ReligiousCities: religious,
			Following:       following,
			Predominant:     religious > 0 && following*2 > religious,
		})
	}
	return out
}

// DetectReligiousWinner returns the team that has achieved a religious victory,
// or -1 if none has. A team wins when a religion founded by one of its members
// is predominant in every living major civ outside the team.
func DetectReligiousWinner(players []Player, gs *GameState) int {
	if gs == nil || len(players) == 0 {
		return -1
	}

	teamOf := make(map[int]int, len(players))
	for _, p := range players {
		teamOf[p.Index] = p.Team
	}

	for _, rel := range gs.Religions {
		if rel.Symbol == 0 || rel.Symbol == noReligion {
			continue
		}
		team, ok := teamOf[rel.FounderPlayer]
		if !ok {
			continue // founded by a non-major (city-state); cannot win
		}

		standings := ReligionStandings(players, gs, rel.Symbol, team)
		if len(standings) == 0 {
			continue // no rival civ to convert — not a religious victory
		}
		allPredominant := true
		for _, s := range standings {
			if !s.Predominant {
				allPredominant = false
				break
			}
		}
		if allPredominant {
			return team
		}
	}
	return -1
}
