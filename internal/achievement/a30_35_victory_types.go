package achievement

func init() {
	RegisterGame(30, func(g G) bool { return g.Winner && g.VictoryType == "Domination" })
	RegisterGame(31, func(g G) bool { return g.Winner && g.VictoryType == "Religious" })
	RegisterGame(32, func(g G) bool { return g.Winner && g.VictoryType == "Science" })
	RegisterGame(33, func(g G) bool { return g.Winner && g.VictoryType == "Culture" })
	RegisterGame(34, func(g G) bool { return g.Winner && g.VictoryType == "Diplomatic" })
	RegisterGame(35, func(g G) bool { return g.Winner && g.VictoryType == "Capitulation" })
}
