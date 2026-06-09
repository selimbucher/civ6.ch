package achievement

func init() {
	RegisterGame(11, func(g G) bool { return g.Leader == "Tokugawa" })
}
