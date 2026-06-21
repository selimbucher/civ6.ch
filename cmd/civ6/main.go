package main

import (
	"context"
	"fmt"
	"image/png"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"

	"github.com/selimbucher/civ6.ch/internal/civ6save"
	"github.com/selimbucher/civ6.ch/internal/db"
	"github.com/selimbucher/civ6.ch/internal/rating"
)

func main() {
	ctx := context.Background()

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "parsesave":
			if len(os.Args) < 3 {
				log.Fatal("usage: civ6 parsesave <file>")
			}
			parseSave(os.Args[2], false)
			return
		case "parsesave-debug":
			if len(os.Args) < 3 {
				log.Fatal("usage: civ6 parsesave-debug <file>")
			}
			parseSave(os.Args[2], true)
			return
		case "detectvictory":
			if len(os.Args) < 3 {
				log.Fatal("usage: civ6 detectvictory <file>")
			}
			detectVictory(os.Args[2])
			return
		case "scorebreakdown":
			if len(os.Args) < 3 {
				log.Fatal("usage: civ6 scorebreakdown <file>")
			}
			printScoreBreakdown(os.Args[2])
			return
		case "scoreaudit":
			if len(os.Args) < 3 {
				log.Fatal("usage: civ6 scoreaudit <file> [playerIndex]")
			}
			idx := 0
			if len(os.Args) >= 4 {
				fmt.Sscanf(os.Args[3], "%d", &idx)
			}
			printScoreAudit(os.Args[2], idx)
			return
		case "buildaudit":
			if len(os.Args) < 3 {
				log.Fatal("usage: civ6 buildaudit <file> [playerIndex]")
			}
			idx := 0
			if len(os.Args) >= 4 {
				fmt.Sscanf(os.Args[3], "%d", &idx)
			}
			printBuildAudit(os.Args[2], idx)
			return
		case "regenerate":
			if len(os.Args) < 3 {
				log.Fatal("usage: civ6 regenerate <gameID>")
			}
			id, err := strconv.Atoi(os.Args[2])
			if err != nil {
				log.Fatalf("invalid game id %q", os.Args[2])
			}
			regenerateGame(ctx, id)
			return
		case "recalculate":
			pool, err := db.Connect(ctx)
			if err != nil {
				log.Fatal(err)
			}
			defer pool.Close()
			if err := rating.RecalculateAll(ctx, pool); err != nil {
				log.Fatal(err)
			}
			log.Println("recalculation complete")
			return
		default:
			log.Fatalf("unknown command: %s", os.Args[1])
		}
	}

	pool, err := db.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// mustLoadState reads, decompresses, and parses the save at path,
// exiting the process on any failure.
func mustLoadState(path string) (decompressed []byte, gs *civ6save.GameState) {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	decompressed, err = civ6save.Decompress(data)
	if err != nil {
		log.Fatal(err)
	}
	gs, err = civ6save.ParseState(decompressed)
	if err != nil {
		log.Fatal(err)
	}
	return decompressed, gs
}

// mustLoadPlayer is mustLoadState plus a bounds-checked player lookup.
// It also prints the common audit header line.
func mustLoadPlayer(path string, playerIdx int) (*civ6save.GameState, *civ6save.PlayerState) {
	decompressed, gs := mustLoadState(path)
	if playerIdx < 0 || playerIdx >= len(gs.Players) || gs.Players[playerIdx] == nil {
		log.Fatalf("player[%d] not found", playerIdx)
	}
	fmt.Printf("file=%s turn=%d player=%d\n", path, civ6save.ParseTurn(decompressed), playerIdx)
	return gs, gs.Players[playerIdx]
}

func printScoreAudit(path string, playerIdx int) {
	_, p := mustLoadPlayer(path, playerIdx)

	for _, c := range p.Cities {
		cityPts := 2
		popPts := c.Population * 2
		distPts := 0
		bldPts := 0
		wndPts := 0

		for _, d := range p.Districts {
			if d.Built != 1 {
				continue
			}
			if d.CityID != c.ID {
				continue
			}
			if civ6save.IsScoreExemptDistrictCRC(d.Type) {
				continue
			}
			distPts += civ6save.DistrictScorePoints(d.Type)
		}

		for crc, v := range c.Built {
			if v == 0xFFFF {
				continue
			}
			if _, isWonder := civ6save.WonderCRCs[crc]; isWonder {
				wndPts += 4
				continue
			}
			if civ6save.IsKnownBuildingCRC(crc) {
				bldPts++
			}
		}

		total := cityPts + popPts + distPts + bldPts
		fmt.Printf("%s: city=%d pop=%d dist=%d bld=%d => empire_city_total=%d (wonders=%d)\n",
			c.Name, cityPts, popPts, distPts, bldPts, total, wndPts)
	}

	b := p.ScoreBreakdown()
	totalEmpire := b.Cities + b.Population + b.Districts + b.Buildings
	fmt.Printf("\nsummary: empire=%d (city=%d pop=%d dist=%d bld=%d) religion=%d (found=%d fol=%d) wonders=%d tech=%d civics=%d gp=%d total=%d\n",
		totalEmpire, b.Cities, b.Population, b.Districts, b.Buildings,
		b.ReligionFounded+b.ForeignFollowers, b.ReligionFounded, b.ForeignFollowers,
		b.Wonders, b.Techs, b.Civics, b.GreatPeople, b.Total)
}

func printBuildAudit(path string, playerIdx int) {
	_, p := mustLoadPlayer(path, playerIdx)

	type entry struct {
		crc  uint32
		v    uint32
		kind string
		name string
	}

	for _, c := range p.Cities {
		entries := make([]entry, 0, len(c.Built))
		notBuilt := 0
		known := 0
		wonders := 0
		unknown := 0
		for crc, v := range c.Built {
			if v == 0xFFFF {
				notBuilt++
				continue
			}
			e := entry{crc: crc, v: v, kind: "unknown", name: "?"}
			if wn, ok := civ6save.WonderCRCs[crc]; ok {
				e.kind = "wonder"
				e.name = wn
				wonders++
			} else if bn, ok := civ6save.BuildingNameByCRC(crc); ok {
				e.kind = "building"
				e.name = bn
				known++
			} else {
				unknown++
			}
			entries = append(entries, e)
		}
		sort.Slice(entries, func(i, j int) bool {
			if entries[i].kind != entries[j].kind {
				return entries[i].kind < entries[j].kind
			}
			if entries[i].name != entries[j].name {
				return entries[i].name < entries[j].name
			}
			return entries[i].crc < entries[j].crc
		})

		fmt.Printf("\n%s: built_entries=%d known_buildings=%d wonders=%d unknown=%d not_built=%d map_entries=%d\n",
			c.Name, len(entries), known, wonders, unknown, notBuilt, len(c.Built))
		for _, e := range entries {
			fmt.Printf("  %s crc=%08x v=%04x name=%s\n", e.kind, e.crc, e.v, e.name)
		}
	}
}

func detectVictory(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	players := civ6save.ParsePlayers(data)
	decompressed, err := civ6save.Decompress(data)
	if err != nil {
		log.Fatal(err)
	}
	gs, err := civ6save.ParseState(decompressed)
	if err != nil {
		log.Fatal(err)
	}

	pseudo := make(map[int]string, len(players))
	for _, p := range players {
		pseudo[p.Index] = p.Pseudo
	}

	fmt.Printf("file=%s turn=%d\n", path, civ6save.ParseTurn(decompressed))
	fmt.Println("original capital control:")
	for _, cc := range civ6save.CapitalControls(players, gs) {
		ownerName := "(unknown)"
		if cc.Owner >= 0 {
			ownerName = pseudo[cc.Owner]
		}
		fmt.Printf("  civ %-2d %-18s capital %-22q held by player %2d (%s) team %d\n",
			cc.Civ, pseudo[cc.Civ], cc.CityName, cc.Owner, ownerName, cc.OwnerTeam)
	}

	winner := civ6save.DetectConquestWinner(players, gs)
	if winner < 0 {
		fmt.Println("=> no conquest victory detected")
	} else {
		fmt.Printf("=> CONQUEST victory: team %d controls every original capital\n", winner)
		for _, p := range players {
			if p.Team == winner {
				fmt.Printf("     winner: player %d (%s)\n", p.Index, p.Pseudo)
			}
		}
	}

	// ── Religion ─────────────────────────────────────────────────────────────
	fmt.Println("religious predominance (rival civs converted / required):")
	teamOf := make(map[int]int, len(players))
	for _, p := range players {
		teamOf[p.Index] = p.Team
	}
	for _, rel := range gs.Religions {
		team, ok := teamOf[rel.FounderPlayer]
		if !ok {
			continue
		}
		name := rel.Name
		if name == "" {
			name = fmt.Sprintf("religion %08x", rel.Symbol)
		}
		standings := civ6save.ReligionStandings(players, gs, rel.Symbol, team)
		converted := 0
		for _, s := range standings {
			if s.Predominant {
				converted++
			}
		}
		fmt.Printf("  %-22s (founder %d, team %d): %d/%d rival civs\n",
			name, rel.FounderPlayer, team, converted, len(standings))
		for _, s := range standings {
			mark := " "
			if s.Predominant {
				mark = "✓"
			}
			fmt.Printf("      %s civ %-2d %d/%d religious cities (%d total)\n",
				mark, s.Civ, s.Following, s.ReligiousCities, s.Cities)
		}
	}

	if rw := civ6save.DetectReligiousWinner(players, gs); rw < 0 {
		fmt.Println("=> no religious victory detected")
	} else {
		fmt.Printf("=> RELIGIOUS victory: team %d\n", rw)
		for _, p := range players {
			if p.Team == rw {
				fmt.Printf("     winner: player %d (%s)\n", p.Index, p.Pseudo)
			}
		}
	}
}

func printScoreBreakdown(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	players := civ6save.ParsePlayers(data)
	meta := make(map[int]civ6save.Player, len(players))
	for _, p := range players {
		meta[p.Index] = p
	}

	decompressed, err := civ6save.Decompress(data)
	if err != nil {
		log.Fatal(err)
	}
	turn := civ6save.ParseTurn(decompressed)
	gs, err := civ6save.ParseState(decompressed)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("file=%s turn=%d\n", path, turn)

	seen := map[int]bool{}
	idxs := make([]int, 0, len(meta)+4)
	for i := range meta {
		idxs = append(idxs, i)
		seen[i] = true
	}
	// Also include likely major non-header players (e.g. AI) and skip city-states.
	for i, p := range gs.Players {
		if p == nil || seen[i] {
			continue
		}
		if len(p.Cities) >= 3 {
			idxs = append(idxs, i)
		}
	}
	sort.Ints(idxs)

	for _, i := range idxs {
		p := gs.Players[i]
		if p == nil {
			continue
		}
		m := meta[i]
		name := m.Pseudo
		if name == "" {
			name = "(unknown)"
		}
		b := p.ScoreBreakdown()
		fmt.Printf("p[%d] %-20s score=%4d | cities=%d pop=%d dist=%d bld=%d wnd=%d tech=%d civic=%d gp=%d era=%d rel=%d fol=%d\n",
			i, name, b.Total,
			b.Cities, b.Population, b.Districts, b.Buildings, b.Wonders,
			b.Techs, b.Civics, b.GreatPeople, b.EraScore, b.ReligionFounded, b.ForeignFollowers)
	}
}

func parseSave(path string, debug bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	// ── players (from header packets) ────────────────────────────────────────
	players := civ6save.ParsePlayers(data)
	for _, p := range players {
		log.Printf("player[%d] team=%d elim=%v steam=%s iColor=%d pseudo='%s' leader=%s color=%v",
			p.Index, p.Team, p.Eliminated, p.SteamID, p.IColor, p.Pseudo, p.Leader,
			civ6save.PlayerColor(p.Leader, p.IColor))
	}
	playerColors := civ6save.BuildPlayerColors(players)

	// ── decompress ───────────────────────────────────────────────────────────
	decompressed, err := civ6save.Decompress(data)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("decompressed: %d bytes", len(decompressed))
	log.Printf("turn: %d", civ6save.ParseTurn(decompressed))

	// ── map ──────────────────────────────────────────────────────────────────
	m, err := civ6save.ParseMap(decompressed)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("map: %d tiles, width: %d, height: %d",
		len(m.Tiles), m.Width, len(m.Tiles)/m.Width)

	ownerCount := make(map[uint16]int)
	for _, t := range m.Tiles {
		if t.Owner < 64 {
			ownerCount[t.Owner]++
		}
	}
	for owner, count := range ownerCount {
		if pc, ok := playerColors[int(owner)]; ok {
			log.Printf("owner[%d]: %d tiles → rgb(%d,%d,%d)",
				owner, count, pc.R, pc.G, pc.B)
		}
	}

	img := civ6save.RenderMap(m, playerColors)
	f, err := os.Create("/tmp/map.png")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	png.Encode(f, img)
	log.Println("wrote /tmp/map.png")

	// ── game state (cities, yields, etc.) ────────────────────────────────────
	gs, err := civ6save.ParseState(decompressed)
	if err != nil {
		log.Printf("ParseState error (work in progress): %v", err)
		return
	}
	if debug {
		if len(gs.Religions) == 0 {
			log.Printf("debug: religions=0")
		} else {
			log.Printf("debug: religions=%d", len(gs.Religions))
			for _, rel := range gs.Religions {
				log.Printf("  religion symbol=%08x founder=%d name='%s' beliefs=%d buildings=%d units=%d",
					rel.Symbol, rel.FounderPlayer, rel.Name, len(rel.Beliefs), len(rel.Buildings), len(rel.Units))
			}
		}

		byPlayer := map[int]int{}
		total := 0
		for i, p := range gs.Players {
			if p == nil || p.GreatPeopleRecruited == 0 {
				continue
			}
			byPlayer[i] = p.GreatPeopleRecruited
			total += p.GreatPeopleRecruited
		}
		keys := make([]int, 0, len(byPlayer))
		for p := range byPlayer {
			keys = append(keys, p)
		}
		sort.Ints(keys)
		log.Printf("debug: recruited great people total=%d players=%d entries=%d", total, len(byPlayer), len(gs.RecruitedGreatPeople))
		for _, p := range keys {
			log.Printf("  gp recruited player[%d]=%d", p, byPlayer[p])
			for _, gp := range gs.RecruitedGreatPeople {
				if gp.Player == p {
					log.Printf("    - class=%08x name=%08x era=%d cost=%d", gp.Class, gp.Name, gp.Era, gp.Cost)
				}
			}
		}
	}

	for i, p := range gs.Players {
		if p == nil {
			continue
		}

		b := p.ScoreBreakdown()
		log.Printf("player[%d] cities=%d gold=%d faith=%d science=%.1f culture=%.1f tourism=%.1f score=%d",
			i, len(p.Cities), p.Gold, p.Faith, p.Science, p.Culture, p.Tourism, b.Total)
		log.Printf("  score breakdown: cities=%d pop=%d districts=%d buildings=%d wonders=%d techs=%d civics=%d great_people=%d era=%d religion=%d followers=%d",
			b.Cities, b.Population, b.Districts, b.Buildings, b.Wonders, b.Techs, b.Civics, b.GreatPeople, b.EraScore, b.ReligionFounded, b.ForeignFollowers)
		if debug {
			log.Printf("  gp debug: recruited=%d current_points=%.1f per_turn=%.1f", p.GreatPeopleRecruited, p.GreatPeopleCurrent, p.GreatPeoplePerTurn)
		}

		if !debug {
			continue
		}

		districtBuilt := 0
		districtByType := map[uint32]int{}
		standardDistricts := 0
		uniqueDistricts := 0
		unknownDistricts := 0
		for _, d := range p.Districts {
			if d.Built != 1 {
				continue
			}
			districtBuilt++
			districtByType[d.Type]++
			switch {
			case civ6save.IsStandardDistrictCRC(d.Type):
				standardDistricts++
			case civ6save.IsUniqueDistrictCRC(d.Type):
				uniqueDistricts++
			default:
				unknownDistricts++
			}
		}
		log.Printf("  debug: districts total=%d built=%d unique_types=%d standard=%d unique=%d unknown=%d",
			len(p.Districts), districtBuilt, len(districtByType), standardDistricts, uniqueDistricts, unknownDistricts)

		for _, c := range p.Cities {
			religionLabel := "none"
			if c.Religion != 0 && c.Religion != 0xFFFFFFFF {
				religionLabel = fmt.Sprintf("%08x", c.Religion)
				if rel := gs.ReligionBySymbol(c.Religion); rel != nil {
					religionLabel = fmt.Sprintf("%s (%08x)", rel.Name, rel.Symbol)
				}
			}

			builtEntries := 0
			knownBuildings := 0
			wonders := 0
			unknownBuilt := 0
			notBuilt := 0
			for crc, v := range c.Built {
				if v == 0xFFFF {
					notBuilt++
					continue
				}
				builtEntries++
				if _, ok := civ6save.WonderCRCs[crc]; ok {
					wonders++
					continue
				}
				if civ6save.IsKnownBuildingCRC(crc) {
					knownBuildings++
				} else {
					unknownBuilt++
				}
			}
			log.Printf("  city '%s' (%d,%d) rel=%s pop=%d food=%.1f prod=%.1f gold=%.1f sci=%.1f cul=%.1f faith=%.1f tour=%.1f built(built_entries=%d known_buildings=%d wonders=%d unknown_built=%d not_built=%d map_entries=%d)",
				c.Name, c.X, c.Y, religionLabel, c.Population,
				c.Food, c.Production, c.Gold,
				c.Science, c.Culture, c.Faith, c.Tourism,
				builtEntries, knownBuildings, wonders, unknownBuilt, notBuilt, len(c.Built))
		}
	}
}
