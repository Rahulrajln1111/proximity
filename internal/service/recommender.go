package service

import (
	"math"
	"sort"

	"proximity/internal/domain"
	"proximity/internal/index"
)

type Recommender struct {
	idx index.Index
}

func NewRecommender(idx index.Index) *Recommender {
	return &Recommender{idx: idx}
}

func (r *Recommender) Recommend(q domain.Query, limit int) []int {
	candidates := r.idx.Candidates(q)

	scored := make([]domain.Scored, 0, len(candidates))
	for _, loc := range candidates {
		if !inside(q, loc) {
			continue
		}
		scored = append(scored, domain.Scored{
			Location: loc,
			Distance: walk(q, loc),
		})
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
	return ids
}

func inside(q domain.Query, loc domain.Location) bool {
	dx := loc.Lat - q.Lat
	dy := loc.Long - q.Long
	return dx*dx+dy*dy <= q.Radius*q.Radius
}

func walk(q domain.Query, loc domain.Location) float64 {
	return math.Abs(loc.Lat-q.Lat) + math.Abs(loc.Long-q.Long)
}
