package achievement

func init() {
	RegisterGame(13, func(g G) bool {
		return g.Winner && g.VictoryType != "Capitulation" && g.Turns > 0 && g.Turns <= 30
	})
}
