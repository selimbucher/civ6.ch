package achievement

func init() {
	RegisterGame(10, func(g G) bool { return g.Winner })
}
