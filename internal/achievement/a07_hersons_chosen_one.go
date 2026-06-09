package achievement

func init() {
	// Reach first place on the leaderboard with a settled RD (≤ 85).
	// Once earned it is never revoked even if the player later drops.
	RegisterStats(7, func(s S) bool {
		return s.Rank == 1 && s.RD <= 85
	})
}
