package achievement

func init() {
	RegisterGame(28, func(g G) bool {
		return g.WinStreakAfter >= 10
	})
}
