package achievement

func init() {
	RegisterGame(14, func(g G) bool { return g.Score >= 2000 })
}
