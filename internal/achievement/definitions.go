package achievement

func init() {
	// ── Stats-based ───────────────────────────────────────────────────────────

	// 7: Reach first place on the leaderboard with a settled RD (≤ 85).
	RegisterStats(7, func(s S) bool {
		return s.Rank == 1 && s.RD <= 85
	})

	// ── Game-based ────────────────────────────────────────────────────────────

	// 3: Reach 1600 overall rating.
	RegisterGame(3, func(g G) bool {
		return g.PostRatingOverall >= 1600
	})

	// 5: Win a Score victory.
	RegisterGame(5, func(g G) bool {
		return g.Winner && g.VictoryType == "Score"
	})

	// 9: Win a Territorial victory.
	RegisterGame(9, func(g G) bool {
		return g.Winner && g.VictoryType == "Territorial"
	})

	// 10: Win any game.
	RegisterGame(10, func(g G) bool {
		return g.Winner
	})

	// 11: Play as Tokugawa.
	RegisterGame(11, func(g G) bool {
		return g.Leader == "Tokugawa"
	})

	// 12: Win as Gandhi in a non-Religious victory.
	RegisterGame(12, func(g G) bool {
		return g.Leader == "Gandhi" && g.Winner && g.VictoryType != "Religious"
	})

	// 13: Win in 30 turns or fewer (non-Capitulation).
	RegisterGame(13, func(g G) bool {
		return g.Winner && g.VictoryType != "Capitulation" && g.Turns > 0 && g.Turns <= 30
	})

	// 14: Reach a score of 2000 or more.
	RegisterGame(14, func(g G) bool {
		return g.Score >= 2000
	})

	// 15: Enter as the weakest player and finish with the highest rating (4+ players).
	RegisterGame(15, func(g G) bool {
		return g.PlayerCount >= 4 &&
			g.PreRatingRank == g.PlayerCount &&
			g.PostRatingRank == 1
	})

	// 16: Win with the lowest score in a 4+ team game.
	RegisterGame(16, func(g G) bool {
		return g.Winner && g.TeamCount >= 4 && g.ScoreRank == g.PlayerCount
	})

	// 17: Play as Basil II.
	RegisterGame(17, func(g G) bool {
		return g.Leader == "Basil II"
	})

	// 18: Be in a game that ends in a Religious victory on turn 42.
	RegisterGame(18, func(g G) bool {
		return g.VictoryType == "Religious" && g.Turns == 42
	})

	// 19: Lose a Capitulation before turn 30.
	RegisterGame(19, func(g G) bool {
		return !g.Winner && g.VictoryType == "Capitulation" && g.Turns > 0 && g.Turns <= 29
	})

	// 22: Win a non-Diplomatic victory after 5+ consecutive losses.
	RegisterGame(22, func(g G) bool {
		return g.Winner && g.VictoryType != "Diplomatic" && g.LosingStreakBefore >= 5
	})

	// 23: Play as Tokugawa against Gandhi.
	RegisterGame(23, func(g G) bool {
		return g.Leader == "Tokugawa" && g.HasEnemy("Gandhi")
	})

	// 24: Reach 1750 overall rating with RD ≤ 100.
	RegisterGame(24, func(g G) bool {
		return g.PostRatingOverall >= 1750 && g.PostRDOverall <= 100
	})

	// 26: End a game with 2000+ diplomatic favor.
	RegisterGame(26, func(g G) bool {
		return g.Favor >= 2000
	})

	// 27: End a game with 2000+ science and 2000+ culture per turn.
	RegisterGame(27, func(g G) bool {
		return g.Science >= 2000 && g.Culture >= 2000
	})

	// 28: Achieve a 10-game winning streak.
	RegisterGame(28, func(g G) bool {
		return g.WinStreakAfter >= 10
	})

	// 29: Win a non-Capitulation game without having researched Mining.
	// nil = data unavailable (pre-tracking game), skip.
	RegisterGame(29, func(g G) bool {
		return g.Winner &&
			g.VictoryType != "Capitulation" &&
			g.MiningResearched != nil && !*g.MiningResearched
	})

	// 30–35: Win each victory type.
	RegisterGame(30, func(g G) bool { return g.Winner && g.VictoryType == "Domination" })
	RegisterGame(31, func(g G) bool { return g.Winner && g.VictoryType == "Religious" })
	RegisterGame(32, func(g G) bool { return g.Winner && g.VictoryType == "Science" })
	RegisterGame(33, func(g G) bool { return g.Winner && g.VictoryType == "Culture" })
	RegisterGame(34, func(g G) bool { return g.Winner && g.VictoryType == "Diplomatic" })
	RegisterGame(35, func(g G) bool { return g.Winner && g.VictoryType == "Capitulation" })
}
