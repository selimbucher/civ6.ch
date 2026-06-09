package achievement

func init() {
	RegisterGame(27, func(g G) bool {
		return g.Science >= 2000 && g.Culture >= 2000
	})
}
