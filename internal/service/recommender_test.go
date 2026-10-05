package service_test

import (
	"math/rand"
	"testing"

	"proximity/internal/domain"
	"proximity/internal/index"
	"proximity/internal/service"
)

type scanIndex struct{ locs []domain.Location }

func (s scanIndex) Candidates(q domain.Query) []domain.Location {
	var out []domain.Location
	for _, l := range s.locs {
		if l.Category == q.Category {
			out = append(out, l)
		}
	}
	return out
}

func TestGridMatchesScan(t *testing.T) {
	cats := []string{
		"bank", "cafe", "hospital", "park",
		"pharmacy", "restaurant", "school", "store",
	}

	var locs []domain.Location
	id := 1
	for i := 0; i < 100; i++ {
		for j := 0; j < 100; j++ {
			locs = append(locs, domain.Location{
				ID:       id,
				Lat:      float64(i) / 99,
				Long:     float64(j) / 99,
				Category: cats[(i*100+j)%len(cats)],
			})
			id++
		}
	}

	grid := service.NewRecommender(index.NewGridIndex(locs))
	scan := service.NewRecommender(scanIndex{locs})

	rng := rand.New(rand.NewSource(42))
	for n := 0; n < 500; n++ {
		q := domain.Query{
			Lat:      rng.Float64(),
			Long:     rng.Float64(),
			Category: cats[rng.Intn(len(cats))],
			Radius:   0.02 + rng.Float64()*0.4,
		}

		got := grid.Recommend(q, 10)
		want := scan.Recommend(q, 10)

		if len(got) != len(want) {
			t.Fatalf("query %+v: got %v want %v", q, got, want)
		}
		for i := range got {
			if got[i] != want[i] {
				t.Fatalf("query %+v: got %v want %v", q, got, want)
			}
		}
	}
}

func TestRankingIsWalkNotStraightLine(t *testing.T) {
	locs := []domain.Location{
		{ID: 1, Lat: 0.1, Long: 0.1, Category: "bank"},
		{ID: 2, Lat: 0.19, Long: 0.0, Category: "bank"},
	}
	rec := service.NewRecommender(scanIndex{locs})

	got := rec.Recommend(domain.Query{Lat: 0, Long: 0, Category: "bank", Radius: 0.5}, 10)

	if len(got) != 2 || got[0] != 2 || got[1] != 1 {
		t.Fatalf("expected [2 1], got %v", got)
	}
}
