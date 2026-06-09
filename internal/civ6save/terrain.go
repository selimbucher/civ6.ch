package civ6save

import "image/color"

var TerrainColors = map[uint32]color.RGBA{
	2213004848: {103, 125, 48, 255},  // grassland
	1855786096: {103, 125, 48, 255},  // grassland hills
	1602466867: {132, 133, 134, 255}, // grassland mountains
	4226188894: {159, 159, 53, 255},  // plains
	3872285854: {159, 159, 53, 255},  // plains hills
	2746853616: {132, 133, 134, 255}, // plains mountains
	3852995116: {236, 196, 111, 255}, // desert
	3108058291: {236, 196, 111, 255}, // desert hills
	1418772217: {132, 133, 134, 255}, // desert mountains
	1223859883: {171, 170, 139, 255}, // tundra
	3949113590: {171, 170, 139, 255}, // tundra hills
	3746160061: {132, 133, 134, 255}, // tundra mountains
	1743422479: {204, 223, 243, 255}, // snow
	3842183808: {204, 223, 243, 255}, // snow hills
	699483892:  {132, 133, 134, 255}, // snow mountains
	1204357597: {45, 49, 86, 255},    // ocean
	1248885265: {45, 89, 120, 255},   // coast
}

var FeatureColors = map[uint32]color.RGBA{
	3910227001: {34, 85, 34, 255},    // rainforest
	1542194068: {171, 188, 219, 255}, // ice
	3727362748: {183, 20, 20, 255},   // barb camp
	1434118760: {241, 209, 100, 255}, // goody hut
}

const opacity = 200

var defaultColor = color.RGBA{30, 30, 30, 255}

func TileColor(terrain, feature uint32) color.RGBA {
	if c, ok := FeatureColors[feature]; ok {
		return c
	}
	if c, ok := TerrainColors[terrain]; ok {
		return c
	}
	return defaultColor
}

func blendColor(base, overlay color.RGBA) color.RGBA {
	a := float64(overlay.A) / 255.0
	return color.RGBA{
		R: uint8(float64(base.R)*(1-a) + float64(overlay.R)*a),
		G: uint8(float64(base.G)*(1-a) + float64(overlay.G)*a),
		B: uint8(float64(base.B)*(1-a) + float64(overlay.B)*a),
		A: 255,
	}
}
