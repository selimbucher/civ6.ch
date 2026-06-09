package achievement

func init() {
	RegisterGame(12, func(g G) bool {
		return g.Leader == "Gandhi" && g.Winner && g.VictoryType != "Religious"
	})
}
