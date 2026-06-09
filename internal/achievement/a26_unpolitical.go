package achievement

func init() {
	RegisterGame(26, func(g G) bool { return g.Favor >= 2000 })
}
