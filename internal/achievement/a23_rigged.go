package achievement

func init() {
	RegisterGame(23, func(g G) bool {
		return g.Leader == "Tokugawa" && g.HasEnemy("Gandhi")
	})
}
