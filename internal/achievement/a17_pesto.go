package achievement

func init() {
	RegisterGame(17, func(g G) bool { return g.Leader == "Basil II" })
}
