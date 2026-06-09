package civ6save

import "hash/crc32"

func districtCRC(s string) uint32 {
	return ^crc32.ChecksumIEEE([]byte(s))
}

// standardDistrictCRCs are non-unique district types that should score 2 points when built.
var standardDistrictCRCs = func() map[uint32]bool {
	names := []string{
		"DISTRICT_CAMPUS",
		"DISTRICT_HOLY_SITE",
		"DISTRICT_ENCAMPMENT",
		"DISTRICT_COMMERCIAL_HUB",
		"DISTRICT_HARBOR",
		"DISTRICT_INDUSTRIAL_ZONE",
		"DISTRICT_THEATER",
		"DISTRICT_ENTERTAINMENT_COMPLEX",
		"DISTRICT_WATER_ENTERTAINMENT_COMPLEX",
		"DISTRICT_AERODROME",
		"DISTRICT_NEIGHBORHOOD",
		"DISTRICT_SPACEPORT",
		"DISTRICT_GOVERNMENT",
		"DISTRICT_DIPLOMATIC_QUARTER",
		"DISTRICT_AQUEDUCT",
		"DISTRICT_CANAL",
		"DISTRICT_DAM",
		"DISTRICT_PRESERVE",
	}
	m := make(map[uint32]bool, len(names))
	for _, n := range names {
		m[districtCRC(n)] = true
	}
	return m
}()

// DistrictScorePoints returns the score contribution of a built, score-eligible district.
// Standard districts are 2; unique/modded non-standard districts default to 4.
func IsStandardDistrictCRC(crc uint32) bool {
	return standardDistrictCRCs[crc]
}

func DistrictScorePoints(districtType uint32) int {
	if IsStandardDistrictCRC(districtType) {
		return 2
	}
	if IsUniqueDistrictCRC(districtType) {
		return 4
	}
	// Unknown/non-standard districts are treated as unique by default.
	return 4
}
