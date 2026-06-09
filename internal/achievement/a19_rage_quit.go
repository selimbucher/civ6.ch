package achievement

func init() {
	// Be on the losing side of a Capitulation before turn 30.
	RegisterGame(19, func(g G) bool {
		return !g.Winner && g.VictoryType == "Capitulation" && g.Turns > 0 && g.Turns <= 29
	})
}
