package index

import (
	"math"
	"sort"

	"proximity/internal/domain"
)

type gridIndex struct {
	cell    float64
	minLat  float64
	minLong float64
	byCats  map[string]map[int64][]domain.Location
}

var _ Index = (*gridIndex)(nil)

func NewGridIndex(locations []domain.Location) *gridIndex {
	g := &gridIndex{
		cell:    cellSize(locations),
		minLat:  math.Inf(1),
		minLong: math.Inf(1),
		byCats:  make(map[string]map[int64][]domain.Location),
	}

	for _, loc := range locations {
		if loc.Lat < g.minLat {
			g.minLat = loc.Lat
		}
		if loc.Long < g.minLong {
			g.minLong = loc.Long
		}
	}
	if math.IsInf(g.minLat, 1) {
		g.minLat = 0
	}
	if math.IsInf(g.minLong, 1) {
		g.minLong = 0
	}

	for _, loc := range locations {
		cells, ok := g.byCats[loc.Category]
		if !ok {
			cells = make(map[int64][]domain.Location)
			g.byCats[loc.Category] = cells
		}
		i := g.cellOf(loc.Lat, g.minLat)
		j := g.cellOf(loc.Long, g.minLong)
		k := key(i, j)
		cells[k] = append(cells[k], loc)
	}
	return g
}

func (g *gridIndex) Candidates(q domain.Query) []domain.Location {
	cells, ok := g.byCats[q.Category]
	if !ok {
		return nil
	}

	ci := g.cellOf(q.Lat, g.minLat)
	cj := g.cellOf(q.Long, g.minLong)
	span := int(math.Ceil(q.Radius/g.cell)) + 1

	var out []domain.Location
	for i := ci - span; i <= ci+span; i++ {
		for j := cj - span; j <= cj+span; j++ {
			if c, ok := cells[key(i, j)]; ok {
				out = append(out, c...)
			}
		}
	}
	return out
}

func (g *gridIndex) cellOf(v, min float64) int {
	return int(math.Floor((v - min) / g.cell))
}

func key(i, j int) int64 {
	return int64(uint64(uint32(i))<<32 | uint64(uint32(j)))
}

func cellSize(locations []domain.Location) float64 {
	if len(locations) < 2 {
		return 1
	}

	lats := make([]float64, 0, len(locations))
	longs := make([]float64, 0, len(locations))
	for _, l := range locations {
		lats = append(lats, l.Lat)
		longs = append(longs, l.Long)
	}

	best := math.Inf(1)
	if d := gap(lats); d > 0 && d < best {
		best = d
	}
	if d := gap(longs); d > 0 && d < best {
		best = d
	}
	if math.IsInf(best, 1) {
		return 1
	}
	return best
}

func gap(vals []float64) float64 {
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
