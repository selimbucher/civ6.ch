package civ6save

var uniqueDistrictCRCs = func() map[uint32]bool {
	names := []string{
		"DISTRICT_ACROPOLIS",
		"DISTRICT_LAVRA",
		"DISTRICT_HANSA",
		"DISTRICT_SEOWON",
		"DISTRICT_MBANZA",
		"DISTRICT_STREET_CARNIVAL",
		"DISTRICT_WATER_STREET_CARNIVAL",
		"DISTRICT_COTHON",
		"DISTRICT_ROYAL_NAVY_DOCKYARD",
		"DISTRICT_IKANDA",
		"DISTRICT_THANH",
		"DISTRICT_OBSERVATORY",
		"DISTRICT_OPPIDUM",
		"DISTRICT_SUGUBA",
		"DISTRICT_HIPPODROME",
		"DISTRICT_BATH",
		// common modded uniques seen in local save corpus
		"DISTRICT_SUK_DZONG",
		"DISTRICT_LIME_TEO_TOLLAN",
	}
	m := make(map[uint32]bool, len(names))
	for _, n := range names {
		m[districtCRC(n)] = true
	}
	return m
}()

func IsUniqueDistrictCRC(crc uint32) bool {
	return uniqueDistrictCRCs[crc]
}
