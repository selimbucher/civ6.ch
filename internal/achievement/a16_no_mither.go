package achievement

func init() {
	// Win a game with at least 4 teams while having the lowest score in the lobby.
	RegisterGame(16, func(g G) bool {
		return g.Winner && g.TeamCount >= 4 && g.ScoreRank == g.PlayerCount
	})
}
