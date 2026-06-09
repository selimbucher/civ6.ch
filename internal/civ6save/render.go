package civ6save

import (
	"image"
	"image/color"
	"math"
)

const hexSize = 12

func hexCorner(cx, cy float64, i int) (float64, float64) {
	angle := math.Pi / 180 * float64(60*i-30)
	return cx + hexSize*math.Cos(angle), cy + hexSize*math.Sin(angle)
}

func drawHex(img *image.RGBA, cx, cy float64, c color.RGBA) {
	corners := make([][2]float64, 6)
	for i := range corners {
		x, y := hexCorner(cx, cy, i)
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

func RenderMap(m *Map, playerColors map[int]color.RGBA) *image.RGBA {
	height := len(m.Tiles) / m.Width

	w := int(float64(m.Width)*hexSize*math.Sqrt(3)) + hexSize*2
	h := int(float64(height)*hexSize*1.5) + hexSize*2

	img := image.NewRGBA(image.Rect(0, 0, w, h))

	for i, tile := range m.Tiles {
		col := i % m.Width
		row := i / m.Width

		cx := float64(hexSize) + float64(col)*hexSize*math.Sqrt(3)
		if row%2 == 0 {
			cx += hexSize * math.Sqrt(3) / 2
		}
		cy := float64(hexSize) + float64(height-1-row)*hexSize*1.5

		c := TileColor(tile.Terrain, tile.Feature)

		if tile.Owner < 64 {
			if pc, ok := playerColors[int(tile.Owner)]; ok {
				c = blendColor(c, pc)
			}
		}

		drawHex(img, cx, cy, c)
	}

	return img
}
