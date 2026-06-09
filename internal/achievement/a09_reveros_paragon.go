package achievement

func init() {
	RegisterGame(9, func(g G) bool {
		return g.Winner && g.VictoryType == "Territorial"
	})
}
