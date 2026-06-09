package achievement

func init() {
	// Win a non-diplomatic victory after at least 5 consecutive losses.
	RegisterGame(22, func(g G) bool {
		return g.Winner &&
			g.VictoryType != "Diplomatic" &&
			g.LosingStreakBefore >= 5
	})
}
