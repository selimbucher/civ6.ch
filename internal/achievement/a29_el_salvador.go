package achievement

func init() {
	// Win a non-capitulation game without having researched Mining.
	// mining_researched is nil for games uploaded before tracking was added;
	// those games are skipped (nil = data not available, cannot confirm nor deny).
	RegisterGame(29, func(g G) bool {
		return g.Winner &&
			g.VictoryType != "Capitulation" &&
			g.MiningResearched != nil && !*g.MiningResearched
	})
}
