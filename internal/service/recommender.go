package service

import (
	"math"
	"sort"

	"proximity/internal/domain"
	"proximity/internal/index"
	"proximity/internal/roadnet"
)

type Recommender struct {
	grid  domain.Grid
	idx   index.Index
	roads roadnet.Source
}

func NewRecommender(g domain.Grid, idx index.Index, roads roadnet.Source) *Recommender {
	return &Recommender{grid: g, idx: idx, roads: roads}
}

func (r *Recommender) Recommend(q domain.Query, limit int) ([]int, error) {
	router, err := r.roads.Resolve(q.Link)
	if err != nil {
		return nil, err
	}

	dist := router.Distances(r.grid.Node(q.Lat, q.Long))
	candidates := r.idx.Candidates(q)

	scored := make([]domain.Scored, 0, len(candidates))
	for _, loc := range candidates {
		if !inside(q, loc) {
			continue
		}
		d, ok := dist[r.grid.Node(loc.Lat, loc.Long)]
		if !ok {
			d = math.Inf(1)
		}
		scored = append(scored, domain.Scored{Location: loc, Distance: d})
	}

	sort.Slice(scored, func(i, j int) bool {
		if scored[i].Distance != scored[j].Distance {
			return scored[i].Distance < scored[j].Distance
		}
		return scored[i].Location.ID < scored[j].Location.ID
	})

	if len(scored) > limit {
		scored = scored[:limit]
	}

	ids := make([]int, len(scored))
	for i, s := range scored {
		ids[i] = s.Location.ID
	}
	return ids, nil
}

func inside(q domain.Query, loc domain.Location) bool {
	dx := loc.Lat - q.Lat
	dy := loc.Long - q.Long
	return dx*dx+dy*dy <= q.Radius*q.Radius
}
