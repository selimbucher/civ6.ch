package civ6save

import (
	"image"
	"image/color"
	"math"
)

// Final hex radius (px) and the supersample factor we render at before
// box-downscaling — this is what gives the map smooth, anti-aliased edges.
const (
	hexSize = 14
	ss      = 2
)

// mapBackground frames the map (visible at the rounded corners / map edge).
var mapBackground = color.RGBA{13, 16, 27, 255}

// cityStateFill darkens a minor civ's territory to near-black; its border colour
// carries the meaning instead.
var cityStateFill = color.RGBA{10, 11, 16, 255}

func hexCornerAt(cx, cy, size float64, i int) (float64, float64) {
	angle := math.Pi / 180 * float64(60*i-30)
	return cx + size*math.Cos(angle), cy + size*math.Sin(angle)
}

func drawHex(img *image.RGBA, cx, cy, size float64, c color.RGBA) {
	corners := make([][2]float64, 6)
	for i := range corners {
		x, y := hexCornerAt(cx, cy, size, i)
		corners[i] = [2]float64{x, y}
	}
	fillPolygon(img, corners, c)
}

func fillPolygon(img *image.RGBA, corners [][2]float64, c color.RGBA) {
	bounds := img.Bounds()
	minY, maxY := corners[0][1], corners[0][1]
	for _, p := range corners {
		if p[1] < minY {
			minY = p[1]
		}
		if p[1] > maxY {
			maxY = p[1]
		}
	}
	for y := int(minY); y <= int(maxY); y++ {
		if y < bounds.Min.Y || y >= bounds.Max.Y {
			continue
		}
		var xs []float64
		n := len(corners)
		for i := 0; i < n; i++ {
			x1, y1 := corners[i][0], corners[i][1]
			x2, y2 := corners[(i+1)%n][0], corners[(i+1)%n][1]
			if (y1 <= float64(y) && y2 > float64(y)) || (y2 <= float64(y) && y1 > float64(y)) {
				x := x1 + (float64(y)-y1)/(y2-y1)*(x2-x1)
				xs = append(xs, x)
			}
		}
		for i := 0; i+1 < len(xs); i += 2 {
			x0, x1 := int(xs[i]), int(xs[i+1])
			if x0 > x1 {
				x0, x1 = x1, x0
			}
			for x := x0; x <= x1; x++ {
				if x >= bounds.Min.X && x < bounds.Max.X {
					img.SetRGBA(x, y, c)
				}
			}
		}
	}
}

func setPix(img *image.RGBA, x, y int, c color.RGBA) {
	b := img.Bounds()
	if x < b.Min.X || x >= b.Max.X || y < b.Min.Y || y >= b.Max.Y {
		return
	}
	img.SetRGBA(x, y, c)
}

// drawThickLine stamps a rounded line of the given half-thickness.
func drawThickLine(img *image.RGBA, x0, y0, x1, y1, thick float64, c color.RGBA) {
	dx, dy := x1-x0, y1-y0
	length := math.Hypot(dx, dy)
	steps := int(length) + 1
	r := int(math.Round(thick))
	for s := 0; s <= steps; s++ {
		t := float64(s) / float64(steps)
		px, py := x0+dx*t, y0+dy*t
		for oy := -r; oy <= r; oy++ {
			for ox := -r; ox <= r; ox++ {
				if ox*ox+oy*oy > r*r {
					continue
				}
				setPix(img, int(px)+ox, int(py)+oy, c)
			}
		}
	}
}

// downscale box-averages src by an integer factor for anti-aliasing.
func downscale(src *image.RGBA, factor int) *image.RGBA {
	if factor <= 1 {
		return src
	}
	b := src.Bounds()
	w, h := b.Dx()/factor, b.Dy()/factor
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	f2 := factor * factor
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			var r, g, bl, a int
			for dy := 0; dy < factor; dy++ {
				for dx := 0; dx < factor; dx++ {
					o := src.PixOffset(x*factor+dx, y*factor+dy)
					r += int(src.Pix[o])
					g += int(src.Pix[o+1])
					bl += int(src.Pix[o+2])
					a += int(src.Pix[o+3])
				}
			}
			o := dst.PixOffset(x, y)
			dst.Pix[o] = uint8(r / f2)
			dst.Pix[o+1] = uint8(g / f2)
			dst.Pix[o+2] = uint8(bl / f2)
			dst.Pix[o+3] = uint8(a / f2)
		}
	}
	return dst
}

// borderColor returns the outline colour for a tile's owner: the major player's
// secondary jersey, a city-state's type colour, or a neutral fallback.
func borderColor(owner uint16, playerColors map[int]ColorPair, cityStates map[int]color.RGBA) color.RGBA {
	if pc, ok := playerColors[int(owner)]; ok {
		return pc.Secondary
	}
	if cs, ok := cityStates[int(owner)]; ok {
		return cs
	}
	return color.RGBA{150, 150, 160, 255}
}

// RenderMap draws the territory map: terrain tinted by each owner's primary
// colour, empires outlined in their secondary colour, and city-states drawn dark
// with a type-coloured border (cityStates maps a minor-civ owner index to its
// colour; nil renders them with a neutral border).
func RenderMap(m *Map, playerColors map[int]ColorPair, cityStates map[int]color.RGBA) *image.RGBA {
	height := len(m.Tiles) / m.Width
	hs := float64(hexSize * ss)

	w := int(float64(m.Width)*hs*math.Sqrt(3)) + int(hs*2)
	h := int(float64(height)*hs*1.5) + int(hs*2)
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i] = mapBackground.R
		img.Pix[i+1] = mapBackground.G
		img.Pix[i+2] = mapBackground.B
		img.Pix[i+3] = 255
	}

	cx := make([]float64, len(m.Tiles))
	cy := make([]float64, len(m.Tiles))
	at := make(map[[2]int]int, len(m.Tiles))
	for i := range m.Tiles {
		col := i % m.Width
		row := i / m.Width
		x := hs + float64(col)*hs*math.Sqrt(3)
		if row%2 == 0 {
			x += hs * math.Sqrt(3) / 2
		}
		y := hs + float64(height-1-row)*hs*1.5
		cx[i], cy[i] = x, y
		at[[2]int{int(math.Round(x)), int(math.Round(y))}] = i
	}
	lookup := func(x, y float64) (int, bool) {
		rx, ry := int(math.Round(x)), int(math.Round(y))
		for oy := -1; oy <= 1; oy++ {
			for ox := -1; ox <= 1; ox++ {
				if idx, ok := at[[2]int{rx + ox, ry + oy}]; ok {
					return idx, true
				}
			}
		}
		return 0, false
	}

	// 1) Terrain, tinted by owner.
	for i, tile := range m.Tiles {
		c := TileColor(tile.Terrain, tile.Feature)
		if tile.Owner < 64 {
			if pc, ok := playerColors[int(tile.Owner)]; ok {
				c = blendColor(c, color.RGBA{pc.Primary.R, pc.Primary.G, pc.Primary.B, opacity})
			} else {
				c = blendColor(c, color.RGBA{cityStateFill.R, cityStateFill.G, cityStateFill.B, 225})
			}
		}
		drawHex(img, cx[i], cy[i], hs, c)
	}

	// 2) Territory outlines, inset just inside each owned tile so neighbouring
	//    empires show two parallel coloured lines at their shared border.
	const inset = 0.82
	thick := float64(ss) * 1.3
	for i, tile := range m.Tiles {
		if tile.Owner >= 64 {
			continue
		}
		bc := borderColor(tile.Owner, playerColors, cityStates)
		for k := 0; k < 6; k++ {
			ang := math.Pi / 180 * float64(60*k)
			nIdx, ok := lookup(cx[i]+hs*math.Sqrt(3)*math.Cos(ang), cy[i]+hs*math.Sqrt(3)*math.Sin(ang))
			if ok && m.Tiles[nIdx].Owner == tile.Owner {
				continue // shared interior edge
			}
			x0, y0 := hexCornerAt(cx[i], cy[i], hs, k)
			x1, y1 := hexCornerAt(cx[i], cy[i], hs, k+1)
			x0, y0 = cx[i]+(x0-cx[i])*inset, cy[i]+(y0-cy[i])*inset
			x1, y1 = cx[i]+(x1-cx[i])*inset, cy[i]+(y1-cy[i])*inset
			drawThickLine(img, x0, y0, x1, y1, thick, bc)
		}
	}

	return downscale(img, ss)
}
