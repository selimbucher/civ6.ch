package achievement

func init() {
	// Enter a game with the lowest rating in the lobby (pre-rating rank = last)
	// and finish with the highest rating overall (post-rating rank = 1).
	// Requires at least 4 players.
	RegisterGame(15, func(g G) bool {
		return g.PlayerCount >= 4 &&
			g.PreRatingRank == g.PlayerCount && // weakest entering
			g.PostRatingRank == 1 // strongest after
	})
}
