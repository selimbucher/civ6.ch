package achievement

func init() {
	RegisterGame(3, func(g G) bool {
		return g.PostRatingOverall >= 1600
	})
}
