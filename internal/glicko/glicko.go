// Package glicko implements Glicko-Civ, an extended Glicko-2 rating system
// adapted for Civilization 6 multiplayer with support for team and FFA scenarios.
package glicko

import (
	"math"
)

const scale = 173.7178

func toGlicko(r float64) float64 {
	return (r - 1500) / scale
}

func fromGlicko(μ float64) float64 {
	return scale*μ + 1500
}

func toGlickoRD(rd float64) float64 {
	return rd / scale
}

func fromGlickoRD(φ float64) float64 {
	return scale * φ
}

func g(φ float64) float64 {
	return 1 / math.Sqrt(1+3*φ*φ/math.Pi/math.Pi)
}

func expected(μ, μj, φj float64) float64 {
	return 1 / (1 + math.Exp(-g(φj)*(μ-μj)))
}

// Result computes the Glicko-2 result scalar from scores.
// Returns 1 for hard win, 0 for hard loss, otherwise cubic-weighted ratio.
func Result(team1Score, team2Score int, team1Won bool, team2Won bool) float64 {
	if team1Won || team2Score == 0 {
		return 1
	}
	if team2Won || team1Score == 0 {
		return 0
	}
	if team1Score == team2Score {
		return 0.5
	}
	a := math.Pow(float64(team1Score), 3)
	b := math.Pow(float64(team2Score), 3)
	ratio := a / (a + b)
	winner := 0.0
	if team1Score > team2Score {
		winner = 1.0
	}
	return winner*0.4 + ratio*0.6
}

// DecayRD increases RD based on days of inactivity, capped at maxRD.
func DecayRD(rd, volatility float64, days float64, maxRD float64) float64 {
	φ := toGlickoRD(rd)
	φ = math.Sqrt(φ*φ + volatility*volatility*days/365)
	return math.Min(fromGlickoRD(φ), maxRD)
}

// Player holds the Glicko-2 state for one player.
type Player struct {
	Rating     float64
	RD         float64
	Volatility float64
}

// Opponent represents an aggregated opposing side.
// In other words, the average opposing team.
type Opponent struct {
	Rating float64
	RD     float64
}

// Update runs one Glicko-2 update for a player against aggregated opponents.
// teamRating and teamRD are the player's team aggregates (pass player values
// for solo). teamRD is currently unused by the algorithm but kept so callers
// supply the full team aggregate.
// results is the slice of result scalars against each opposing team.
// opponents is the slice of opposing team aggregates.
// teamSize is used to split the rating gain for team games.
func Update(p Player, teamRating, teamRD float64, opponents []Opponent, results []float64, teamSize int, τ float64) Player {
	μ := toGlicko(p.Rating)
	φ := toGlickoRD(p.RD)
	σ := p.Volatility

	μt := toGlicko(teamRating)

	μj := make([]float64, len(opponents))
	φj := make([]float64, len(opponents))
	for i, o := range opponents {
		μj[i] = toGlicko(o.Rating)
		φj[i] = toGlickoRD(o.RD)
	}

	// Step 3: compute v
	v := 0.0
	for i := range opponents {
		gv := g(φj[i])
		e := expected(μt, μj[i], φj[i])
		v += gv * gv * e * (1 - e)
	}
	v = 1 / v

	// Step 4: compute Δ
	delta := 0.0
	for i := range opponents {
		e := expected(μt, μj[i], φj[i])
		delta += g(φj[i]) * (results[i] - e)
	}
	delta *= v

	// Step 5: update volatility
	a := math.Log(σ * σ)
	f := func(x float64) float64 {
		ex := math.Exp(x)
		num := ex * (delta*delta - φ*φ - v - ex)
		den := 2 * math.Pow(φ*φ+v+ex, 2)
		return num/den - (x-a)/(τ*τ)
	}

	A := a
	B := 0.0
	if delta*delta > φ*φ+v {
		B = math.Log(delta*delta - φ*φ - v)
	} else {
		k := 1.0
		for f(a-k*τ) < 0 {
			k++
		}
		B = a - k*τ
	}

	fA, fB := f(A), f(B)
	ε := 0.000001
	for math.Abs(B-A) > ε {
		C := A + (A-B)*fA/(fB-fA)
		fC := f(C)
		if fC*fB <= 0 {
			A = B
			fA = fB
		} else {
			fA /= 2
		}
		B = C
		fB = fC
	}
	σ2 := math.Exp(A / 2)

	// Step 6: φ*
	φs := math.Sqrt(φ*φ + σ2*σ2)

	// Step 7: new φ′ and μ′
	φ2 := 1 / math.Sqrt(1/φs/φs+1/v)
	μ2 := μ
	for i := range opponents {
		e := expected(μt, μj[i], φj[i])
		μ2 += φ2 * φ2 * g(φj[i]) * (results[i] - e)
	}

	// team size split
	r2 := (fromGlicko(μ2)-p.Rating)/float64(teamSize) + p.Rating

	return Player{
		Rating:     r2,
		RD:         fromGlickoRD(φ2),
		Volatility: σ2,
	}
}
