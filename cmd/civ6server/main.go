package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"slices"
	"strconv"

	"github.com/chai2010/webp"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/selimbucher/civ6.ch/internal/civ6save"
	"github.com/selimbucher/civ6.ch/internal/db"
	"github.com/selimbucher/civ6.ch/internal/rating"
	"github.com/selimbucher/civ6.ch/internal/storage"
)

type server struct {
	pool  *pgxpool.Pool
	store storage.Backend
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	pool, err := db.Connect(context.Background())
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer pool.Close()

	store, err := storage.New(
		os.Getenv("STORAGE_BACKEND"),
		os.Getenv("STORAGE_PATH"),
	)
	if err != nil {
		log.Fatalf("storage: %v", err)
	}

	s := &server{pool: pool, store: store}

	// Dynamic REST routes using Go 1.22+ parameter matching
	http.HandleFunc("GET /files/saves/{id}", s.handleGetSave)
	http.HandleFunc("GET /files/maps/{id}", s.handleGetMap)
	http.HandleFunc("GET /files/avatars/{id}", s.handleGetAvatar)
	http.HandleFunc("POST /players/{id}/avatar", s.handleUploadAvatar)
	http.HandleFunc("DELETE /games/{id}", s.handleDeleteGameFiles)
	http.HandleFunc("POST /games/{id}/update", s.handleUpdateSave)

	// Clean fallback matching for root endpoints to prevent internal 404 routing conflicts
	http.HandleFunc("/parse", s.handleParse)
	http.HandleFunc("/recalculate", s.handleRecalculate)

	log.Printf("listening on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

// ── Parse ─────────────────────────────────────────────────────────────────────

type parseResponse struct {
	GameID int `json:"game_id"`
}

func (s *server) handleParse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil || len(data) == 0 {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	// Deduplicate by content hash — same save file must not produce two rows.
	sum := sha256.Sum256(data)
	hashStr := hex.EncodeToString(sum[:])

	var existingID int
	if err := s.pool.QueryRow(r.Context(),
		`SELECT id FROM games WHERE save_hash = $1`, hashStr,
	).Scan(&existingID); err == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(parseResponse{GameID: existingID})
		return
	}

	filename := r.Header.Get("X-Filename")

	settings := civ6save.ParseSettings(data)
	players := civ6save.ParsePlayers(data)

	decompressed, err := civ6save.Decompress(data)
	if err != nil {
		log.Printf("decompress: %v", err)
	}

	var turn int
	var state *civ6save.GameState
	if decompressed != nil {
		turn = civ6save.ParseTurn(decompressed)
		if state, err = civ6save.ParseState(decompressed); err != nil {
			log.Printf("parse state: %v", err)
		}
	}

	gameID, err := insertGame(r.Context(), s.pool, settings, turn, players, state, hashStr, filename)
	if err != nil {
		log.Printf("insert: %v", err)
		http.Error(w, fmt.Sprintf("database error: %v", err), http.StatusInternalServerError)
		return
	}

	// Store the save file and rendered map in the background so the HTTP
	// response is not delayed by encoding/IO.
	go s.storeFiles(gameID, data, decompressed, players)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(parseResponse{GameID: gameID})
}

func (s *server) storeFiles(gameID int, raw, decompressed []byte, players []civ6save.Player) {
	ctx := context.Background()

	// ── Save file (gzip) ──────────────────────────────────────────────────
	var saveBuf bytes.Buffer
	gz, _ := gzip.NewWriterLevel(&saveBuf, gzip.BestCompression)
	gz.Write(raw)
	gz.Close()

	saveKey := fmt.Sprintf("saves/%d.Civ6Save.gz", gameID)
	if err := s.store.Put(ctx, saveKey, saveBuf.Bytes(), "application/gzip"); err != nil {
		log.Printf("store save %d: %v", gameID, err)
	} else {
		log.Printf("stored save %d (%d B → %d B)", gameID, len(raw), saveBuf.Len())
		s.pool.Exec(ctx, `UPDATE games SET has_save = true WHERE id = $1`, gameID)
	}

	// ── Map image (WebP lossless) ─────────────────────────────────────────
	if decompressed == nil {
		return
	}
	m, err := civ6save.ParseMap(decompressed)
	if err != nil {
		log.Printf("parse map %d: %v", gameID, err)
		return
	}
	colors := civ6save.BuildPlayerColors(players)
	img := civ6save.RenderMap(m, colors)

	var mapBuf bytes.Buffer
	if err := webp.Encode(&mapBuf, img, &webp.Options{Lossless: true}); err != nil {
		log.Printf("encode map %d: %v", gameID, err)
		return
	}
	mapKey := fmt.Sprintf("maps/%d.webp", gameID)
	if err := s.store.Put(ctx, mapKey, mapBuf.Bytes(), "image/webp"); err != nil {
		log.Printf("store map %d: %v", gameID, err)
	} else {
		log.Printf("stored map %d (%d B)", gameID, mapBuf.Len())
		s.pool.Exec(ctx, `UPDATE games SET has_map = true WHERE id = $1`, gameID)
	}
}

// ── File serving ──────────────────────────────────────────────────────────────

func (s *server) handleGetSave(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	compressed, err := s.store.Get(r.Context(), fmt.Sprintf("saves/%d.Civ6Save.gz", id))
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	gr, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		http.Error(w, "decompress error", http.StatusInternalServerError)
		return
	}
	defer gr.Close()

	var turns int16
	var saveFilename *string
	s.pool.QueryRow(r.Context(),
		`SELECT COALESCE(turns, 0), save_filename FROM games WHERE id = $1`, id,
	).Scan(&turns, &saveFilename)

	filename := fmt.Sprintf("AutoSave_%04d.Civ6Save", turns)
	if saveFilename != nil && *saveFilename != "" {
		filename = *saveFilename
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename=%q`, filename))
	io.Copy(w, gr)
}

// ── Avatars ────────────────────────────────────────────────────────────────────

func (s *server) handleUploadAvatar(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	// Cap at 6 MB; the web layer already validates type/size.
	data, err := io.ReadAll(io.LimitReader(r.Body, 6<<20))
	if err != nil || len(data) == 0 {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	ct := r.Header.Get("Content-Type")
	if ct == "" {
		ct = http.DetectContentType(data)
	}
	if err := s.store.Put(r.Context(), fmt.Sprintf("avatars/%d", id), data, ct); err != nil {
		http.Error(w, fmt.Sprintf("store error: %v", err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *server) handleGetAvatar(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	data, err := s.store.Get(r.Context(), fmt.Sprintf("avatars/%d", id))
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	// Avatars change in place, so revalidate rather than cache hard.
	w.Header().Set("Content-Type", http.DetectContentType(data))
	w.Header().Set("Cache-Control", "no-cache")
	w.Write(data)
}

func (s *server) handleGetMap(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	data, err := s.store.Get(r.Context(), fmt.Sprintf("maps/%d.webp", id))
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	// Maps can be regenerated under the same URL, so no immutable caching.
	w.Header().Set("Content-Type", "image/webp")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.Write(data)
}

// ── Update save ──────────────────────────────────────────────────────────────

type updateResponse struct {
	OK      bool   `json:"ok"`
	Turns   int    `json:"turns,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

func (s *server) handleUpdateSave(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil || len(data) == 0 {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	reply := func(r updateResponse) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(r)
	}

	var existingMap string
	var existingTurns int16
	err = s.pool.QueryRow(r.Context(),
		`SELECT COALESCE(map,''), COALESCE(turns,0) FROM games WHERE id=$1`, id,
	).Scan(&existingMap, &existingTurns)
	if err != nil {
		http.Error(w, "game not found", http.StatusNotFound)
		return
	}

	rows, err := s.pool.Query(r.Context(),
		`SELECT id, COALESCE(leader,'') FROM game_players WHERE game_id=$1 ORDER BY id`, id)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type gpRow struct {
		id     int
		leader string
	}
	var existing []gpRow
	for rows.Next() {
		var g gpRow
		rows.Scan(&g.id, &g.leader)
		existing = append(existing, g)
	}

	settings := civ6save.ParseSettings(data)
	newPlayers := civ6save.ParsePlayers(data)

	decompressed, err := civ6save.Decompress(data)
	if err != nil {
		reply(updateResponse{Error: "failed to decompress save"})
		return
	}
	newTurn := civ6save.ParseTurn(decompressed)

	if existingMap != "" && settings.Map != "" && settings.Map != existingMap {
		reply(updateResponse{Error: fmt.Sprintf("wrong map: save has %q, game has %q", settings.Map, existingMap)})
		return
	}
	if newTurn <= int(existingTurns) {
		reply(updateResponse{Error: fmt.Sprintf("save must be newer (game is at turn %d, new save is turn %d)", existingTurns, newTurn)})
		return
	}
	if len(newPlayers) != len(existing) {
		reply(updateResponse{Error: fmt.Sprintf("player count changed (%d → %d)", len(existing), len(newPlayers))})
		return
	}

	existingByLeader := make(map[string]int)
	for _, g := range existing {
		existingByLeader[g.leader] = g.id
	}
	for _, np := range newPlayers {
		slug := civ6save.LeaderSlug(np.Leader)
		if _, ok := existingByLeader[slug]; !ok {
			reply(updateResponse{Error: fmt.Sprintf("leader %q not in original game", slug)})
			return
		}
	}

	var state *civ6save.GameState
	if decompressed != nil {
		state, _ = civ6save.ParseState(decompressed)
	}

	tx, err := s.pool.Begin(r.Context())
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())

	sum := sha256.Sum256(data)
	tx.Exec(r.Context(), `UPDATE games SET turns=$1, save_hash=$2 WHERE id=$3`,
		int16(newTurn), hex.EncodeToString(sum[:]), id)

	for _, np := range newPlayers {
		slug := civ6save.LeaderSlug(np.Leader)
		gpID := existingByLeader[slug]
		if state == nil || state.Players[np.Index] == nil {
			continue
		}
		ps := state.Players[np.Index]
		tx.Exec(r.Context(), `
			UPDATE game_players SET
				score=$1, population=$2, science=$3, culture=$4,
				food=$5, production=$6, gold=$7, faith=$8, tourism=$9, favor=$10
			WHERE id=$11`,
			ps.Score(), totalPopulation(ps),
			roundToInt(ps.Science), roundToInt(ps.Culture),
			roundToInt(ps.Food), roundToInt(ps.Production),
			ps.Gold, ps.Faith, roundToInt(ps.Tourism), ps.DiploFavor,
			gpID,
		)
	}

	if err = tx.Commit(r.Context()); err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	go s.storeFiles(id, data, decompressed, newPlayers)
	reply(updateResponse{OK: true, Turns: newTurn, Message: fmt.Sprintf("Updated to turn %d", newTurn)})
}

func (s *server) handleDeleteGameFiles(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	for _, key := range []string{
		fmt.Sprintf("saves/%d.Civ6Save.gz", id),
		fmt.Sprintf("maps/%d.webp", id),
	} {
		if err := s.store.Delete(ctx, key); err != nil {
			log.Printf("delete file %s: %v", key, err)
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

// ── Recalculate ───────────────────────────────────────────────────────────────

func (s *server) handleRecalculate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := rating.RecalculateAll(r.Context(), s.pool); err != nil {
		log.Printf("recalculate: %v", err)
		http.Error(w, fmt.Sprintf("recalculate error: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}

// ── DB insert ─────────────────────────────────────────────────────────────────

func insertGame(ctx context.Context, pool *pgxpool.Pool, settings civ6save.GameSettings, turn int, players []civ6save.Player, state *civ6save.GameState, saveHash, saveFilename string) (int, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	allowConquest := slices.Contains(settings.EnabledVictories, "VICTORY_CONQUEST")
	allowScore := slices.Contains(settings.EnabledVictories, "VICTORY_SCORE")
	allowScience := slices.Contains(settings.EnabledVictories, "VICTORY_TECHNOLOGY")
	allowReligious := slices.Contains(settings.EnabledVictories, "VICTORY_RELIGIOUS")
	allowCulture := slices.Contains(settings.EnabledVictories, "VICTORY_CULTURE")
	allowDiplomatic := slices.Contains(settings.EnabledVictories, "VICTORY_DIPLOMATIC")
	shuffleTechs := slices.Contains(settings.Modes, "TREE_RANDOMIZER")
	secretSocieties := slices.Contains(settings.Modes, "SECRETSOCIETIES")
	heroesAndLegends := slices.Contains(settings.Modes, "HEROES_AND_LEGENDS")
	apocalypseMode := slices.Contains(settings.Modes, "APOCALYPSE")
	monopolies := slices.Contains(settings.Modes, "MONOPOLIES")
	barbarianClans := slices.Contains(settings.Modes, "BARBARIAN_CLANS")
	zombieDefense := slices.Contains(settings.Modes, "ZOMBIE_DEFENSE")

	mapSize := stripPrefix(settings.MapSize, "MAPSIZE_")
	gameSpeed := stripPrefix(settings.GameSpeed, "GAMESPEED_")
	era := stripPrefix(settings.CurrentEra, "ERA_")
	ruleset := stripPrefix(settings.Ruleset, "RULESET_")
	difficulty := stripPrefix(settings.Difficulty, "DIFFICULTY_")

	var gameID int
	err = tx.QueryRow(ctx, `
		INSERT INTO games (
			victory_type, turns, map, map_size, game_speed,
			shuffle_techs, allow_conquest, allow_score, allow_science,
			allow_religious, allow_culture, allow_diplomatic,
			secret_societies, heroes_and_legends, apocalypse_mode, monopolies,
			barbarian_clans, zombie_defense,
			era, ruleset, difficulty,
			tmp, save_hash, save_filename
		) VALUES (
			'Unknown', $1, $2, $3, $4,
			$5, $6, $7, $8, $9, $10, $11,
			$12, $13, $14, $15,
			$16, $17,
			$18, $19, $20,
			true, $21, $22
		) RETURNING id`,
		int16(turn),
		nullStr(settings.Map), nullStr(mapSize), nullStr(gameSpeed),
		shuffleTechs, allowConquest, allowScore, allowScience,
		allowReligious, allowCulture, allowDiplomatic,
		secretSocieties, heroesAndLegends, apocalypseMode, monopolies,
		barbarianClans, zombieDefense,
		nullStr(era), nullStr(ruleset), nullStr(difficulty),
		nullStr(saveHash), nullStr(saveFilename),
	).Scan(&gameID)
	if err != nil {
		return 0, err
	}

	for _, p := range players {
		leader := civ6save.LeaderSlug(p.Leader)

		var score *int
		var population, science, culture, food, production, gold, faith, tourism, favor *int
		var miningResearched *bool
		if state != nil && state.Players[p.Index] != nil {
			v := state.Players[p.Index].HasTech(civ6save.MiningTechCRC)
			miningResearched = &v
			ps := state.Players[p.Index]
			s := ps.Score()
			score = intPtr(s)
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

		// Eliminated players are recorded as participants but with score 0.
		if p.Eliminated {
			score = intPtr(0)
		}

		var gpID int
		err = tx.QueryRow(ctx, `
			INSERT INTO game_players (
				game_id, team, player_index, leader, pseudo_name, score,
				population, science, culture, food, production, gold, faith, tourism, favor,
				mining_researched, eliminated, steam_id
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
			RETURNING id`,
			gameID, int16(p.Team), int16(p.Index), leader, nullStr(p.Pseudo), score,
			population, science, culture, food, production, gold, faith, tourism, favor,
			miningResearched, p.Eliminated, nullStr(p.SteamID),
		).Scan(&gpID)
		if err != nil {
			return 0, err
		}

		if state != nil && state.Players[p.Index] != nil {
			ps := state.Players[p.Index]
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
					return 0, err
				}
			}
		}
	}

	return gameID, tx.Commit(ctx)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

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

// roundToInt rounds a float32 yield to the nearest int for DB storage.
func roundToInt(f float32) int { return int(math.Round(float64(f))) }
