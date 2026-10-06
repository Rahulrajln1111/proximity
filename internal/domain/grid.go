package domain

import (
	"math"
	"sort"
)

type Node struct {
	I int
	J int
}

type Grid struct {
	MinLat  float64
	MinLong float64
	Step    float64
}

func NewGrid(locations []Location) Grid {
	g := Grid{Step: 1}
	if len(locations) == 0 {
		return g
	}

	lats := make([]float64, 0, len(locations))
	longs := make([]float64, 0, len(locations))
	minLat := math.Inf(1)
	minLong := math.Inf(1)
	for _, l := range locations {
		if l.Lat < minLat {
			minLat = l.Lat
		}
		if l.Long < minLong {
			minLong = l.Long
		}
		lats = append(lats, l.Lat)
		longs = append(longs, l.Long)
	}

	g.MinLat = minLat
	g.MinLong = minLong

	step := math.Inf(1)
	if d := smallestGap(lats); d > 0 && d < step {
		step = d
	}
	if d := smallestGap(longs); d > 0 && d < step {
		step = d
	}
	if !math.IsInf(step, 1) {
		g.Step = step
	}
	return g
}

func (g Grid) Node(lat, long float64) Node {
	return Node{
		I: int(math.Round((lat - g.MinLat) / g.Step)),
		J: int(math.Round((long - g.MinLong) / g.Step)),
	}
}

func (g Grid) LatLon(n Node) (float64, float64) {
	return g.MinLat + float64(n.I)*g.Step, g.MinLong + float64(n.J)*g.Step
}

func smallestGap(vals []float64) float64 {
	sort.Float64s(vals)
	best := math.Inf(1)
	for i := 1; i < len(vals); i++ {
		if d := vals[i] - vals[i-1]; d > 1e-12 && d < best {
			best = d
		}
	}
	if math.IsInf(best, 1) {
		return 0
	}
	return best
}
