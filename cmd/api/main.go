package main

import (
	"flag"
	"log"
	"net/http"

	"proximity/internal/index"
	"proximity/internal/loader"
	"proximity/internal/service"
	"proximity/internal/transport"
)

func main() {
	csvPath := flag.String("csv", "locations.csv", "path to locations csv")
	addr := flag.String("addr", ":8080", "listen address")
	flag.Parse()

	locations, err := loader.NewCSVLoader(*csvPath).Load()
	if err != nil {
		log.Fatalf("load locations: %v", err)
	}
	log.Printf("loaded %d locations from %s", len(locations), *csvPath)

	rec := service.NewRecommender(index.NewGridIndex(locations))
	handler := transport.NewHandler(rec)

	mux := http.NewServeMux()
	mux.Handle("/search/", handler)
	mux.Handle("/search", handler)

	log.Printf("listening on %s", *addr)
	if err := http.ListenAndServe(*addr, mux); err != nil {
		log.Fatalf("server: %v", err)
	}
}
