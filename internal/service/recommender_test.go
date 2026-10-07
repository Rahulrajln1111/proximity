package service_test

import (
	"math/rand"
	"strings"
	"testing"

	"proximity/internal/domain"
	"proximity/internal/index"
	"proximity/internal/roadnet"
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

func lattice() []domain.Location {
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
	return locs
}

func TestGridMatchesScan(t *testing.T) {
	locs := lattice()
	grid := domain.NewGrid(locs)
	src := roadnet.NewCache(grid, roadnet.CompleteNetwork(grid, locs))

	a := service.NewRecommender(grid, index.NewGridIndex(grid, locs), src)
	b := service.NewRecommender(grid, scanIndex{locs}, src)

	cats := []string{"bank", "cafe", "hospital", "park", "pharmacy", "restaurant", "school", "store"}
	rng := rand.New(rand.NewSource(42))

	for n := 0; n < 500; n++ {
		q := domain.Query{
			Lat:      rng.Float64(),
			Long:     rng.Float64(),
			Category: cats[rng.Intn(len(cats))],
			Radius:   0.02 + rng.Float64()*0.4,
		}

		got, err := a.Recommend(q, 10)
		if err != nil {
			t.Fatal(err)
		}
		want, err := b.Recommend(q, 10)
		if err != nil {
			t.Fatal(err)
		}

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

func corners() []domain.Location {
	return []domain.Location{
		{ID: 1, Lat: 0, Long: 0, Category: "bank"},
		{ID: 2, Lat: 0, Long: 1, Category: "bank"},
		{ID: 3, Lat: 1, Long: 0, Category: "bank"},
		{ID: 4, Lat: 1, Long: 1, Category: "bank"},
	}
}

func TestMissingEdgeLengthensRoute(t *testing.T) {
	locs := corners()
	grid := domain.NewGrid(locs)

	// loop over the long way: the direct (0,0)-(1,0) road is missing
	link := strings.Join([]string{
		"0 0 1 0",
		"1 0 1 1",
		"1 1 0 1",
	}, "\n")

	src := roadnet.NewCache(grid, roadnet.CompleteNetwork(grid, locs))
	rec := service.NewRecommender(grid, index.NewGridIndex(grid, locs), src)

	got, err := rec.Recommend(domain.Query{
		Lat: 0, Long: 0, Category: "bank", Radius: 5, Link: link,
	}, 10)
	if err != nil {
		t.Fatal(err)
	}

	want := []int{1, 2, 4, 3}
	if len(got) != len(want) {
		t.Fatalf("got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v want %v", got, want)
		}
	}
}

func TestFullGridIsManhattan(t *testing.T) {
	locs := corners()
	grid := domain.NewGrid(locs)

	src := roadnet.NewCache(grid, roadnet.CompleteNetwork(grid, locs))
	rec := service.NewRecommender(grid, index.NewGridIndex(grid, locs), src)

	got, err := rec.Recommend(domain.Query{
		Lat: 0, Long: 0, Category: "bank", Radius: 5,
	}, 10)
	if err != nil {
		t.Fatal(err)
	}

	want := []int{1, 2, 3, 4}
	if len(got) != len(want) {
		t.Fatalf("got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v want %v", got, want)
		}
	}
}

func TestUnreachableRankedLast(t *testing.T) {
	locs := corners()
	grid := domain.NewGrid(locs)

	// id 2 sits at (0,1) which no road reaches
	link := strings.Join([]string{
		"0 0 0 1",
		"0 1 1 1",
	}, "\n")

	src := roadnet.NewCache(grid, roadnet.CompleteNetwork(grid, locs))
	rec := service.NewRecommender(grid, index.NewGridIndex(grid, locs), src)

	got, err := rec.Recommend(domain.Query{
		Lat: 0, Long: 0, Category: "bank", Radius: 5, Link: link,
	}, 10)
	if err != nil {
		t.Fatal(err)
	}

	want := []int{1, 3, 4, 2}
	if len(got) != len(want) {
		t.Fatalf("got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v want %v", got, want)
		}
	}
}

func BenchmarkRecommendFullGrid(b *testing.B) {
	locs := lattice()
	grid := domain.NewGrid(locs)
	src := roadnet.NewCache(grid, roadnet.CompleteNetwork(grid, locs))
	rec := service.NewRecommender(grid, index.NewGridIndex(grid, locs), src)

	q := domain.Query{Lat: 0.25, Long: 0.25, Category: "bank", Radius: 0.1}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := rec.Recommend(q, 10); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRecommendWideRadius(b *testing.B) {
	locs := lattice()
	grid := domain.NewGrid(locs)
	src := roadnet.NewCache(grid, roadnet.CompleteNetwork(grid, locs))
	rec := service.NewRecommender(grid, index.NewGridIndex(grid, locs), src)

	q := domain.Query{Lat: 0.5, Long: 0.5, Category: "bank", Radius: 1.5}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := rec.Recommend(q, 10); err != nil {
			b.Fatal(err)
		}
	}
}

func TestRadiusIsStraightLine(t *testing.T) {
	locs := []domain.Location{
		{ID: 1, Lat: 0, Long: 0.09, Category: "bank"},
		{ID: 2, Lat: 0, Long: 0.11, Category: "bank"},
	}
	grid := domain.NewGrid(locs)
	src := roadnet.NewCache(grid, roadnet.CompleteNetwork(grid, locs))
	rec := service.NewRecommender(grid, index.NewGridIndex(grid, locs), src)

	got, err := rec.Recommend(domain.Query{
		Lat: 0, Long: 0, Category: "bank", Radius: 0.1,
	}, 10)
	if err != nil {
		t.Fatal(err)
	}

	if len(got) != 1 || got[0] != 1 {
		t.Fatalf("radius filter should keep only id 1, got %v", got)
	}
}
