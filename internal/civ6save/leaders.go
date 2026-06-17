package civ6save

import "strings"

// leaderSlugs maps the internal LEADER_* identifier (without LEADER_ prefix, uppercase)
// to the canonical icon slug matching web/src/lib/assets/icons/leaders/<slug>_(Civ6).webp.
var leaderSlugs = map[string]string{
	// A
	"ABRAHAM_LINCOLN": "Abraham_Lincoln",
	"AHIRAM":          "Ahiram",
	"ALEXANDER":       "Alexander",
	"ALFONSOEL_SABIO": "Alfonso_XI",
	"ALFONSO_XI":      "Alfonso_XI",
	"AL_HAKAM_II":     "Al-Hakam_II",
	"AL_HASAN":        "Al-Hasan",
	"AMANITORE":       "Amanitore",
	"AMBIORIX":        "Ambiorix",
	"ANACAONA":        "Anacaona",

	// B
	"BASIL": "Basil_II",
	"BEF":   "BEF",

	// C
	"CANUTE":                  "Canute",
	"CATHERINE":               "Catherine_de_Medici",
	"CATHERINE_DE_MEDICI":     "Catherine_de_Medici",
	"CATHERINE_DE_MEDICI_ALT": "Catherine_de_Medici_(Magnificence)",
	"CHANDRAGUPTA":            "Chandragupta",
	"CHARLEMAGNE":             "Charlemagne",
	"CLEOPATRA":               "Cleopatra",
	"CLEOPATRA_PTOLEMAIC":     "Cleopatra_(Ptolemaic)",
	"CYRUS":                   "Cyrus",

	// D
	"DIDO": "Dido",

	// E
	"ELEANOR":              "Eleanor_of_Aquitaine",
	"ELEANOR_OF_AQUITAINE": "Eleanor_of_Aquitaine",
	"ELEANOR_ENGLAND":      "Eleanor_of_Aquitaine_(English)",
	"ELEANOR_FRANCE":       "Eleanor_of_Aquitaine_(French)",
	"ELIZABETH":            "Elizabeth_I",

	// F
	"FREDERICK_BARBAROSSA": "Frederick_Barbarossa",

	// G
	"GANDHI":       "Gandhi",
	"GENGHIS_KHAN": "Genghis_Khan",
	"GILGAMESH":    "Gilgamesh",
	"GITARJA":      "Gitarja",
	"GORGO":        "Gorgo",

	// H
	"HAMMURABI":          "Hammurabi",
	"Harald":             "Harald_Hardrada",
	"HARDRADA":           "Harald_Hardrada",
	"HARDRADA_NORWAY":    "Harald_Hardrada",
	"HARDRADA_VARANGIAN": "Harald_Hardrada_(Varangian)",
	"HERALD_ALT":         "Harald_Hardrada_(Varangian)",
	"HOJO_TOKIMUNE":      "Hojo_Tokimune",

	// I
	"INGOLFUR": "Ingólfur_Arnarson",

	// J
	"JADWIGA":       "Jadwiga",
	"JFD_STANISLAW": "Stanislaw_II",
	"JAYAVARMAN":    "Jayavarman_VII",
	"JOAO_III":      "João_III",
	"JOHN_CURTIN":   "John_Curtin",
	"JULIUS_CAESAR": "Julius_Caesar",

	// K
	"KRISTINA":             "Kristina",
	"KUBLAI_KHAN_CHINA":    "Kublai_Khan_(Chinese)",
	"KUBLAI_KHAN_MONGOLIA": "Kublai_Khan_(Mongolian)",
	"KUPE":                 "Kupe",

	// L
	"LADY_SIX_SKY":   "Lady_Six_Sky",
	"LADY_TRIEU":     "Bà_Triệu",
	"LAUTARO":        "Lautaro",
	"LIME_TEO_OWL":   "Spearthrower_Owl",
	"LL_TEKINICH_II": "Te'_K'inich_II",
	"LUDWIG":         "Ludwig_II",

	// M
	"MANSA_MUSA":        "Mansa_Musa",
	"MATTHIAS_CORVINUS": "Matthias_Corvinus",
	"MENELIK":           "Menelik_II",
	"MER_MARIA_THERESA": "Maria_Theresa",
	"MER_THEODORIC":     "Theodoric",
	"MONTEZUMA":         "Montezuma",
	"MVEMBA":            "Mvemba_a_Nzinga",

	// N
	"NZINGA_MBANDE": "Nzinga_Mbande",

	// P
	"PACHACUTI":  "Pachacuti",
	"PEDRO":      "Pedro_II",
	"PERICLES":   "Pericles",
	"PETER":      "Peter",
	"PHILIP_II":  "Philip_II",
	"POUNDMAKER": "Poundmaker",

	// Q
	"QIN":     "Qin_Shi_Huang",
	"QIN_ALT": "Qin_Shi_Huang_(Unifier)",

	// R
	"RAMSES":       "Ramses_II",
	"ROBERT_BRUCE": "Robert_the_Bruce",

	// S
	"SALADIN":           "Saladin",
	"SALADIN_SULTAN":    "Saladin_(Sultan)",
	"SEJONG":            "Sejong",
	"SEONDEOK":          "Seondeok",
	"SHAKA":             "Shaka",
	"SIMON_BOLIVAR":     "Simón_Bolívar",
	"SULEIMAN":          "Suleiman",
	"SULEIMAN_MUHTEŞEM": "Suleiman_(Muhteşem)",
	"SUNDIATA_KEITA":    "Sundiata_Keita",

	// T
	"TAMAR":                      "Tamar",
	"TEDDY_ROOSEVELT":            "Teddy_Roosevelt",
	"TEDDY_ROOSEVELT_ROUGHRIDER": "Teddy_Roosevelt_(Rough_Rider)",
	"TOKUGAWA":                   "Tokugawa",
	"TOMYRIS":                    "Tomyris",
	"TRAJAN":                     "Trajan",

	// V
	"VICTORIA":     "Victoria",
	"VICTORIA_ALT": "Victoria_(Age_of_Steam)",

	// W
	"WILFRID_LAURIER": "Wilfrid_Laurier",
	"WILHELMINA":      "Wilhelmina",
	"WU_ZETIAN":       "Wu_Zetian",

	// Y
	"YONGLE": "Yongle",
}

// LeaderSlug converts a raw save leader string (e.g. "LEADER_SEONDEOK" or a
// legacy mixed-case/spaced DB value) to the canonical icon slug used for portrait
// filenames (e.g. "Seondeok").  Returns the input unchanged if no mapping is found.
func LeaderSlug(raw string) string {
	// Strip LEADER_ prefix if present.
	s := raw
	if strings.HasPrefix(strings.ToUpper(s), "LEADER_") {
		s = s[7:]
	}
	upper := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(s), " ", "_"))
	if slug, ok := leaderSlugs[upper]; ok {
		return slug
	}
	// Fallback: return the stripped/trimmed original so nothing is lost.
	return strings.TrimSpace(s)
}
