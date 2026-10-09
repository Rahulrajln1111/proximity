# Proximity API

A lightweight HTTP service that recommends the ten nearest locations to a query
point, ranked by **shortest-path road distance** over a supplied road network
rather than straight-line distance. Written in Go using only the standard library.

## Overview

Given a coordinate, a category, a search radius, and an optional road graph, the
API filters candidate locations by a Euclidean radius and then ranks the survivors
by their actual traversal distance on the road network. Locations that cannot be
reached are ranked last.

- **Radius** is a circular (Euclidean) filter — which locations are *eligible*.
- **Distance** is the road-network traversal cost (BFS over the graph) — which
  eligible locations are *closest*.

Because some direct links between grid points may be missing (as in a real road
network), straight-line distance is not a reliable proximity measure, so ranking
is done on the graph.

## API

### Endpoint

```
GET /search/
Content-Type: application/json
```

### Parameters

| Parameter | Required | Type   | Description                                              |
|-----------|----------|--------|----------------------------------------------------------|
| `lat`     | Yes      | float  | Query latitude                                           |
| `long`    | Yes      | float  | Query longitude                                          |
| `cat`     | Yes      | string | Category to search (e.g. `bank`, `cafe`)                 |
| `rad`     | Yes      | float  | Search radius (circular, non-negative)                   |
| `link`    | No       | string | Road graph source: HTTP(S) URL, inline text, or file path |

### Response

```json
{"recommended_ids":[4850,4753,5155,5349,5452,4948,5156,5348,4648,4654]}
```

The array holds the top-10 location IDs ranked 1st → 10th by shortest-path road
distance.

### Behavior

- **Radius filter:** a location is kept when `dx² + dy² <= rad²`.
- **Ranking:** BFS shortest path over the road graph given in `link`.
- **Unreachable locations:** assigned distance = infinity, ranked last; ties break by ascending ID.
- **No `link` supplied:** falls back to a complete 4-neighbour grid network (Manhattan-style ordering).
- **`link` auto-detection:**
  - `http(s)://…` → downloaded over HTTP (cached)
  - contains newlines → parsed inline from the query string
  - otherwise → treated as a filesystem path
- **Snapping:** a query point not exactly on a lattice node is snapped to the nearest node before BFS.

### Errors

| Condition                              | Status | Body                              |
|----------------------------------------|--------|-----------------------------------|
| Missing required parameter             | 400    | `{"error":"missing parameter ..."}` |
| Invalid number / negative radius       | 400    | `{"error":"invalid ..."}`         |
| Bad `link` value                       | 400    | `{"error":"..."}`                 |
| Non-GET method                         | 405    | `{"error":"use GET"}`             |

## Getting started

### Build

```bash
go build -o proximity-api ./cmd/api
```

### Run

```bash
./proximity-api -csv locations.csv -addr :8080
```

| Flag   | Default          | Description              |
|--------|------------------|--------------------------|
| `-csv` | `locations.csv`  | Path to the dataset CSV  |
| `-addr`| `:8080`          | Listen address           |

### Example requests

```bash
# With a road-graph file
curl -s "http://localhost:8080/search/?lat=0.5&long=0.5&cat=bank&rad=0.3&link=link.txt"

# Full grid (no link supplied)
curl -s "http://localhost:8080/search/?lat=0.5&long=0.5&cat=bank&rad=0.3"
```

### Dataset format

`locations.csv` has the columns `ID, Latitude, Longitude, Category`. Locations lie
on a 1 × 1 lattice; the grid origin and step are inferred from the data.

The `link` road file lists one edge per line as space-separated coordinate pairs
describing an undirected connection between two grid points.

## Architecture

```
cmd/api/main.go
  ├─ flags: -csv, -addr
  ├─ loader.CSVLoader        → []domain.Location
  ├─ domain.NewGrid          → lattice grid (origin + step)
  ├─ roadnet.CompleteNetwork → full-grid baseline network
  └─ service.Recommender
        ├─ index.NewGridIndex → cell-based candidate filter
        └─ roadnet.NewCache   → link resolution + BFS routers
                 │
          transport.Handler  → GET /search/
```

| Package              | Role                                                                 |
|----------------------|----------------------------------------------------------------------|
| `internal/domain`    | Core types: `Location`, `Query`, `Scored`, `Grid`, `Node`            |
| `internal/loader`    | `CSVLoader` — parses and validates `locations.csv`                   |
| `internal/index`     | `GridIndex` — cell-keyed candidate filter on the lattice             |
| `internal/roadnet`   | `Network` (BFS), `ParseNetwork`, `CompleteNetwork`, `Cache`          |
| `internal/service`   | `Recommender` — radius filter + BFS ranking                          |
| `internal/transport` | `Handler` — HTTP parsing, response/error encoding                    |

## How it works

- **Grid inference:** the lattice origin and step are derived from the dataset —
  the minimum coordinate and the smallest positive gap between sorted coordinates —
  so no step size is hard-coded.
- **Candidate filtering:** locations are bucketed per category into grid cells. A
  query scans only the cells within `ceil(rad/step)+1` of the query node, falling
  back to a full category scan when the window would cover the whole category anyway.
- **Routing:** each distinct `link` produces a `Network` whose BFS from the query
  node yields hop-count distances to every reachable node. Parsed graphs are cached
  per link value, so repeated queries reuse them.
- **Ranking:** survivors of the radius filter are sorted by BFS distance, with
  unreachable nodes placed last and ID as a deterministic tie-breaker.

## Testing

```bash
go vet ./...
go test ./...
go build ./...
```

The service test suite covers grid-index-vs-full-scan equivalence, missing-edge
rerouting, full-grid fallback, unreachable ranking, and radius filtering, plus
benchmarks. At 10,000 locations the measured cost is ~1.9 ms/query.

## License

MIT
