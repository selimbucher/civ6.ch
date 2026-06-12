package civ6save

// districtNoScore are district types that don't contribute to score.
var districtNoScore = map[uint32]bool{
	0x8a32007d: true, // DISTRICT_CITY_CENTER
	0x62f1b509: true, // DISTRICT_WONDER
	0x226247b8: true, // DISTRICT_CITY_OF_LIGHTS
	0xd2a343c7: true, // DISTRICT_BARBARIAN_OUTPOST
	0x4ef3d333: true, // DISTRICT_TRIBAL_VILLAGE
}

// IsScoreExemptDistrictCRC reports whether the district type does not
// contribute to score (city centers, wonder plots, barbarian camps, …).
func IsScoreExemptDistrictCRC(crc uint32) bool {
	return districtNoScore[crc]
}

// ScoreBreakdown computes score using BBG scoring rules:
// - city 2, pop 2, district 2 (unique 4), building 1, wonder 4,
// - tech 3, civic 4, great person 3,
// - religion founded 10, foreign followers 1,
// - era score removed (0 points).
func (ps *PlayerState) ScoreBreakdown() ScoreBreakdown {
	var b ScoreBreakdown

	b.Cities = len(ps.Cities) * 2
	for _, c := range ps.Cities {
		b.Population += c.Population * 2
	}

	for _, d := range ps.Districts {
		if d.Built != 1 || districtNoScore[d.Type] {
			continue
		}
		b.Districts += DistrictScorePoints(d.Type)
	}

	// Wonders are deduped globally per player to avoid accidental double counting.
	seenWonders := map[string]bool{}
	for _, c := range ps.Cities {
		for _, w := range c.Wonders {
			if !seenWonders[w] {
				seenWonders[w] = true
				b.Wonders += 4
			}
		}
	}

	// Buildings: in city.Built, 0xFFFF appears to mean "not built".
	for _, c := range ps.Cities {
		for crc, v := range c.Built {
			if v == 0xFFFF {
				continue
			}
			if _, isWonder := WonderCRCs[crc]; isWonder {
				continue
			}
			if IsKnownBuildingCRC(crc) {
				b.Buildings++
			}
		}
	}

	b.Techs = len(ps.TechsResearched) * 3
	b.Civics = len(ps.CivicsResearched) * 4
	b.GreatPeople = ps.GreatPeopleRecruited * 3
	if ps.ReligionFounded {
		b.ReligionFounded = 10
	}
	b.ForeignFollowers = ps.ForeignCitiesFollowingReligion
	b.EraScore = 0 // BBG removes era score points

	b.Total = b.Cities + b.Population + b.Districts + b.Buildings + b.Wonders +
		b.Techs + b.Civics + b.GreatPeople + b.ReligionFounded + b.ForeignFollowers + b.EraScore
	return b
}

// Score returns the total from ScoreBreakdown.
func (ps *PlayerState) Score() int {
	return ps.ScoreBreakdown().Total
}
