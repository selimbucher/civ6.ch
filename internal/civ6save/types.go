package civ6save

// CityState is the parsed state for one city at the current turn.
type CityState struct {
	ID         int
	X, Y       int
	Population int
	Name       string
	Religion   uint32

	// OriginalOwner is the player index that founded the city. It survives
	// capture (the city keeps it even when owned by another player), so it
	// identifies which civ a captured capital originally belonged to.
	OriginalOwner int
	// IsCapital is true for the player's current capital (the city holding the
	// palace). Note this moves on capture, so it marks the current — not
	// necessarily the original — capital.
	IsCapital bool

	// per-city yields (gross)
	Food       float32
	Production float32
	Gold       float32
	Science    float32
	Culture    float32
	Faith      float32
	Tourism    float32

	// construction/progression maps
	Built   map[uint32]uint32
	Prod    map[uint32]uint32
	Wonders []string // display names of built wonders

	// religion follower presence records from city religion block.
	// key: internal religion hash (not necessarily Symbol), value: follower-like amount.
	ReligionFollowers map[uint32]uint32
}

// DistrictState is the parsed state for one district instance.
type DistrictState struct {
	GlobalID int
	ID       int
	X, Y     int
	CityID   int
	Type     uint32
	Damage   int
	Wall     int
	Built    int
	Pillage  int
}

// MiningTechCRC is the Civ6 CRC (^crc32.IEEE) of "TECH_MINING".
const MiningTechCRC uint32 = 0x88915fe6

// HasTech reports whether the player has researched the tech with the given CRC32.
func (ps *PlayerState) HasTech(crc uint32) bool {
	return ps.TechsResearched[crc]
}

// PlayerState is the parsed state for one player slot.
type PlayerState struct {
	IPlayer int
	// StartX, StartY are the player's settler start plot — the anchor for
	// locating this civ's original capital (see DetectConquestWinner).
	StartX, StartY int
	Cities         []CityState
	Districts      []DistrictState
	Gold           int
	Faith          int
	Government     uint32
	DiploFavor     int

	// totals computed from city yields
	Science    float32
	Culture    float32
	Food       float32
	Production float32
	Tourism    float32

	// tech and civic trees — maps of CRC(name) → researched
	TechsResearched  map[uint32]bool
	CivicsResearched map[uint32]bool

	// great people
	GreatPeopleRecruited int
	GreatPeopleCurrent   float32 // sum of current GP point progress buckets
	GreatPeoplePerTurn   float32 // sum of GP point generation buckets

	ReligionFounded                bool
	ForeignCitiesFollowingReligion int
	EraScore                       int
}

// ScoreBreakdown is a transparent point-by-point score decomposition.
type ScoreBreakdown struct {
	Cities           int
	Population       int
	Districts        int
	Buildings        int
	Wonders          int
	Techs            int
	Civics           int
	GreatPeople      int
	ReligionFounded  int
	ForeignFollowers int
	EraScore         int
	Total            int
}

// ReligionState represents one founded religion definition in the save.
type ReligionState struct {
	FounderPlayer int
	Symbol        uint32
	Name          string
	Beliefs       []uint32
	Buildings     []uint32
	Units         []uint32
}

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

// GameState is the full parsed turn state.
type GameState struct {
	Players              [64]*PlayerState
	Religions            []ReligionState
	RecruitedGreatPeople []GreatPersonRecruit
}
