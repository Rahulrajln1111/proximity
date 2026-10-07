package transport

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"proximity/internal/domain"
	"proximity/internal/service"
)

const topN = 10

type Handler struct {
	rec *service.Recommender
}

func NewHandler(rec *service.Recommender) *Handler {
	return &Handler{rec: rec}
}

type result struct {
	IDs []int `json:"recommended_ids"`
}

type failure struct {
	Error string `json:"error"`
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "use GET")
		return
	}

	q, err := readQuery(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	ids, err := h.rec.Recommend(q, topN)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	_ = json.NewEncoder(w).Encode(result{IDs: ids})
}

func readQuery(r *http.Request) (domain.Query, error) {
	v := r.URL.Query()

	lat, err := number(v.Get("lat"), "lat")
	if err != nil {
		return domain.Query{}, err
	}

	long, err := number(v.Get("long"), "long")
	if err != nil {
		return domain.Query{}, err
	}

	rad, err := number(v.Get("rad"), "rad")
	if err != nil {
		return domain.Query{}, err
	}
	if rad < 0 {
		return domain.Query{}, fmt.Errorf("parameter %q must not be negative", "rad")
	}

	cat := v.Get("cat")
	if cat == "" {
		return domain.Query{}, fmt.Errorf("missing parameter %q", "cat")
	}

	return domain.Query{
		Lat:      lat,
		Long:     long,
		Category: cat,
		Radius:   rad,
		Link:     v.Get("link"),
	}, nil
}

func number(raw, name string) (float64, error) {
	if raw == "" {
		return 0, fmt.Errorf("missing parameter %q", name)
	}
	n, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %q: %q is not a number", name, raw)
	}
	return n, nil
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(failure{Error: msg})
}
