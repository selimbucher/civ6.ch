package achievement

func init() {
	RegisterGame(24, func(g G) bool {
		return g.PostRatingOverall >= 1750 && g.PostRDOverall <= 100
	})
}
