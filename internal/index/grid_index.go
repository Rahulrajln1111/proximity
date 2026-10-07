package index

import (
	"math"

	"proximity/internal/domain"
)

type gridIndex struct {
	grid  domain.Grid
	byCat map[string]map[int64][]domain.Location
}

var _ Index = (*gridIndex)(nil)

func NewGridIndex(g domain.Grid, locations []domain.Location) *gridIndex {
	idx := &gridIndex{
		grid:  g,
		byCat: make(map[string]map[int64][]domain.Location),
	}

	for _, loc := range locations {
		cells, ok := idx.byCat[loc.Category]
		if !ok {
			cells = make(map[int64][]domain.Location)
			idx.byCat[loc.Category] = cells
		}
		n := g.Node(loc.Lat, loc.Long)
		k := key(n.I, n.J)
		cells[k] = append(cells[k], loc)
	}
	return idx
}

func (g *gridIndex) Candidates(q domain.Query) []domain.Location {
	cells, ok := g.byCat[q.Category]
	if !ok {
		return nil
	}

	n := g.grid.Node(q.Lat, q.Long)
	span := int(math.Ceil(q.Radius/g.grid.Step)) + 1

	if int64(2*span+1)*int64(2*span+1) >= int64(len(cells)) {
		out := make([]domain.Location, 0, len(cells))
		for _, c := range cells {
			out = append(out, c...)
		}
		return out
	}

	var out []domain.Location
	for i := n.I - span; i <= n.I+span; i++ {
		for j := n.J - span; j <= n.J+span; j++ {
			if c, ok := cells[key(i, j)]; ok {
				out = append(out, c...)
			}
		}
	}
	return out
}

func key(i, j int) int64 {
	return int64(uint64(uint32(i))<<32 | uint64(uint32(j)))
}
