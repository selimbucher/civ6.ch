package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"slices"

	"github.com/chai2010/webp"
	"github.com/selimbucher/civ6.ch/internal/civ6save"
	"github.com/selimbucher/civ6.ch/internal/db"
	"github.com/selimbucher/civ6.ch/internal/storage"
)

// regenerateGame re-parses the stored save of a game and rewrites all
// parse-derived data: games settings columns, game_players stats, city rows
// and the rendered map image. Manually curated fields (victory_type, winner,
// player_id, ratings, tmp) are left untouched.
func regenerateGame(ctx context.Context, gameID int) {
	pool, err := db.Connect(ctx)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer pool.Close()

	store, err := storage.New(os.Getenv("STORAGE_BACKEND"), os.Getenv("STORAGE_PATH"))
	if err != nil {
		log.Fatalf("storage: %v", err)
	}

	// ── load stored save ──────────────────────────────────────────────────
	saveKey := fmt.Sprintf("saves/%d.Civ6Save.gz", gameID)
	compressed, err := store.Get(ctx, saveKey)
	if err != nil {
		log.Fatalf("load %s: %v", saveKey, err)
	}
	gr, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		log.Fatalf("gunzip %s: %v", saveKey, err)
	}
	data, err := io.ReadAll(gr)
	gr.Close()
	if err != nil {
		log.Fatalf("gunzip %s: %v", saveKey, err)
	}
	log.Printf("loaded save %s (%d B)", saveKey, len(data))

	// ── re-parse ──────────────────────────────────────────────────────────
	settings := civ6save.ParseSettings(data)
	players := civ6save.ParsePlayers(data)

	decompressed, err := civ6save.Decompress(data)
	if err != nil {
		log.Fatalf("decompress: %v", err)
	}
	turn := civ6save.ParseTurn(decompressed)

	state, err := civ6save.ParseState(decompressed)
	if err != nil {
		log.Printf("parse state: %v (continuing without per-player stats)", err)
	}

	// ── update DB ─────────────────────────────────────────────────────────
	tx, err := pool.Begin(ctx)
	if err != nil {
		log.Fatalf("db begin: %v", err)
	}
	defer tx.Rollback(ctx)

	mapSize := stripPrefix(settings.MapSize, "MAPSIZE_")
	gameSpeed := stripPrefix(settings.GameSpeed, "GAMESPEED_")
	era := stripPrefix(settings.CurrentEra, "ERA_")
	ruleset := stripPrefix(settings.Ruleset, "RULESET_")
	difficulty := stripPrefix(settings.Difficulty, "DIFFICULTY_")

	tag, err := tx.Exec(ctx, `
		UPDATE games SET
			turns=$1, map=$2, map_size=$3, game_speed=$4,
			shuffle_techs=$5, allow_conquest=$6, allow_score=$7, allow_science=$8,
			allow_religious=$9, allow_culture=$10, allow_diplomatic=$11,
			secret_societies=$12, heroes_and_legends=$13, apocalypse_mode=$14, monopolies=$15,
			barbarian_clans=$16, zombie_defense=$17,
			era=$18, ruleset=$19, difficulty=$20
		WHERE id=$21`,
		int16(turn),
		nullStr(settings.Map), nullStr(mapSize), nullStr(gameSpeed),
		slices.Contains(settings.Modes, "TREE_RANDOMIZER"),
		slices.Contains(settings.EnabledVictories, "VICTORY_CONQUEST"),
		slices.Contains(settings.EnabledVictories, "VICTORY_SCORE"),
		slices.Contains(settings.EnabledVictories, "VICTORY_TECHNOLOGY"),
		slices.Contains(settings.EnabledVictories, "VICTORY_RELIGIOUS"),
		slices.Contains(settings.EnabledVictories, "VICTORY_CULTURE"),
		slices.Contains(settings.EnabledVictories, "VICTORY_DIPLOMATIC"),
		slices.Contains(settings.Modes, "SECRETSOCIETIES"),
		slices.Contains(settings.Modes, "HEROES_AND_LEGENDS"),
		slices.Contains(settings.Modes, "APOCALYPSE"),
		slices.Contains(settings.Modes, "MONOPOLIES"),
		slices.Contains(settings.Modes, "BARBARIAN_CLANS"),
		slices.Contains(settings.Modes, "ZOMBIE_DEFENSE"),
		nullStr(era), nullStr(ruleset), nullStr(difficulty),
		gameID,
	)
	if err != nil {
		log.Fatalf("update games: %v", err)
	}
	if tag.RowsAffected() == 0 {
		log.Fatalf("game %d not found", gameID)
	}

	// Existing player rows, matched by team (= save player index).
	rows, err := tx.Query(ctx,
		`SELECT id, team FROM game_players WHERE game_id=$1`, gameID)
	if err != nil {
		log.Fatalf("query game_players: %v", err)
	}
	gpByTeam := make(map[int]int)
	for rows.Next() {
		var id int
		var team int16
		if err := rows.Scan(&id, &team); err != nil {
			log.Fatalf("scan game_players: %v", err)
		}
		gpByTeam[int(team)] = id
	}
	rows.Close()

	for _, p := range players {
		gpID, ok := gpByTeam[p.Index]
		if !ok {
			log.Printf("warning: save player %d (%s) has no game_players row, skipping", p.Index, p.Leader)
			continue
		}
		delete(gpByTeam, p.Index)

		leader := civ6save.LeaderSlug(p.Leader)

		var score *int
		var population, science, culture, food, production, gold, faith, tourism, favor *int
		var miningResearched *bool
		if state != nil && state.Players[p.Index] != nil {
			ps := state.Players[p.Index]
			v := ps.HasTech(civ6save.MiningTechCRC)
			miningResearched = &v
			score = intPtr(ps.Score())
			population = intPtr(totalPopulation(ps))
			science = intPtr(roundToInt(ps.Science))
			culture = intPtr(roundToInt(ps.Culture))
			food = intPtr(roundToInt(ps.Food))
			production = intPtr(roundToInt(ps.Production))
			gold = intPtr(ps.Gold)
			faith = intPtr(ps.Faith)
			tourism = intPtr(roundToInt(ps.Tourism))
			favor = intPtr(ps.DiploFavor)
		}

		_, err = tx.Exec(ctx, `
			UPDATE game_players SET
				leader=$1, pseudo_name=$2, score=$3,
				population=$4, science=$5, culture=$6, food=$7, production=$8,
				gold=$9, faith=$10, tourism=$11, favor=$12,
				mining_researched=$13
			WHERE id=$14`,
			leader, nullStr(p.Pseudo), score,
			population, science, culture, food, production,
			gold, faith, tourism, favor,
			miningResearched, gpID,
		)
		if err != nil {
			log.Fatalf("update game_players %d: %v", gpID, err)
		}
		log.Printf("updated player team=%d leader=%s (game_players id=%d)", p.Index, leader, gpID)

		if state == nil || state.Players[p.Index] == nil {
			continue
		}
		ps := state.Players[p.Index]

		if _, err = tx.Exec(ctx,
			`DELETE FROM game_player_cities WHERE game_player_id=$1`, gpID); err != nil {
			log.Fatalf("delete cities for game_players %d: %v", gpID, err)
		}
		for _, c := range ps.Cities {
			var relName *string
			if c.Religion != 0 && c.Religion != 0xFFFFFFFF {
				if rel := state.ReligionBySymbol(c.Religion); rel != nil {
					name := rel.Name
					relName = &name
				}
			}
			wonders := c.Wonders
			if wonders == nil {
				wonders = []string{}
			}
			_, err = tx.Exec(ctx, `
				INSERT INTO game_player_cities (
					game_player_id, name, population, religion, wonders,
					food, production, gold, science, culture, faith
				) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
				gpID, c.Name, c.Population, relName, wonders,
				c.Food, c.Production, c.Gold, c.Science, c.Culture, c.Faith,
			)
			if err != nil {
				log.Fatalf("insert city %q: %v", c.Name, err)
			}
		}
	}
	for team, gpID := range gpByTeam {
		log.Printf("warning: game_players row id=%d (team=%d) has no matching player in save", gpID, team)
	}

	if err := tx.Commit(ctx); err != nil {
		log.Fatalf("db commit: %v", err)
	}

	// ── re-render map ─────────────────────────────────────────────────────
	m, err := civ6save.ParseMap(decompressed)
	if err != nil {
		log.Fatalf("parse map: %v", err)
	}
	img := civ6save.RenderMap(m, civ6save.BuildPlayerColors(players))

	var mapBuf bytes.Buffer
	if err := webp.Encode(&mapBuf, img, &webp.Options{Lossless: true}); err != nil {
		log.Fatalf("encode map: %v", err)
	}
	mapKey := fmt.Sprintf("maps/%d.webp", gameID)
	if err := store.Put(ctx, mapKey, mapBuf.Bytes(), "image/webp"); err != nil {
		log.Fatalf("store map: %v", err)
	}
	pool.Exec(ctx, `UPDATE games SET has_map = true WHERE id = $1`, gameID)
	log.Printf("stored map %s (%d B)", mapKey, mapBuf.Len())

	log.Printf("game %d regenerated (turn %d, %d players)", gameID, turn, len(players))
}

// ── helpers (mirrors of cmd/civ6server) ───────────────────────────────────────

func totalPopulation(ps *civ6save.PlayerState) int {
	total := 0
	for _, c := range ps.Cities {
		total += c.Population
	}
	return total
}

func stripPrefix(s, prefix string) string {
	if len(s) > len(prefix) && s[:len(prefix)] == prefix {
		return s[len(prefix):]
	}
	return s
}

func nullStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func intPtr(v int) *int { return &v }

func roundToInt(f float32) int { return int(math.Round(float64(f))) }
