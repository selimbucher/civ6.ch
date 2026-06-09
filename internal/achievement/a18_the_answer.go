package achievement

func init() {
	// The game ends in a Religious victory exactly on turn 42.
	// Player does not need to be the winner.
	RegisterGame(18, func(g G) bool {
		return g.VictoryType == "Religious" && g.Turns == 42
	})
}
