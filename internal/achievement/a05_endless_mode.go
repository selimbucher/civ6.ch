package achievement

func init() {
	RegisterGame(5, func(g G) bool {
		return g.Winner && g.VictoryType == "Score"
	})
}
