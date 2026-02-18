package generate

import "emoji-roguelike/internal/gamemap"

// carveCorridor digs an L-shaped tunnel between (x1,y1) and (x2,y2).
func carveCorridor(gmap *gamemap.GameMap, x1, y1, x2, y2 int, cfg *Config) {
	switch cfg.CorridorStyle {
	case CorridorZShaped:
		carveZShaped(gmap, x1, y1, x2, y2)
	case CorridorStraight:
		carveH(gmap, x1, x2, y1)
		carveV(gmap, y1, y2, x2)
	default: // LShaped
		if cfg.Rand.Intn(2) == 0 {
			carveH(gmap, x1, x2, y1)
			carveV(gmap, y1, y2, x2)
		} else {
			carveV(gmap, y1, y2, x1)
			carveH(gmap, x1, x2, y2)
		}
	}
}

func carveH(gmap *gamemap.GameMap, x1, x2, y int) {
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	for x := x1; x <= x2; x++ {
		if gmap.InBounds(x, y) {
			gmap.Set(x, y, gamemap.MakeFloor())
		}
	}
}

func carveV(gmap *gamemap.GameMap, y1, y2, x int) {
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	for y := y1; y <= y2; y++ {
		if gmap.InBounds(x, y) {
			gmap.Set(x, y, gamemap.MakeFloor())
		}
	}
}

func carveZShaped(gmap *gamemap.GameMap, x1, y1, x2, y2 int) {
	midY := (y1 + y2) / 2
	carveV(gmap, y1, midY, x1)
	carveH(gmap, x1, x2, midY)
	carveV(gmap, midY, y2, x2)
}
