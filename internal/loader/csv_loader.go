package loader

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"

	"proximity/internal/domain"
)

type CSVLoader struct {
	path string
}

func NewCSVLoader(path string) *CSVLoader {
	return &CSVLoader{path: path}
}

func (l *CSVLoader) Load() ([]domain.Location, error) {
	f, err := os.Open(l.path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", l.path, err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.FieldsPerRecord = -1

	rows, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", l.path, err)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("read %s: empty file", l.path)
	}

	out := make([]domain.Location, 0, len(rows)-1)
	for i, row := range rows {
		if i == 0 {
			continue
		}
		loc, err := parseRow(row, i+1)
		if err != nil {
			return nil, err
		}
		out = append(out, loc)
	}
	return out, nil
}

func parseRow(row []string, line int) (domain.Location, error) {
	if len(row) < 4 {
		return domain.Location{}, fmt.Errorf("line %d: want 4 columns, got %d", line, len(row))
	}

	id, err := strconv.Atoi(strings.TrimSpace(row[0]))
	if err != nil {
		return domain.Location{}, fmt.Errorf("line %d: bad id %q: %w", line, row[0], err)
	}

	lat, err := strconv.ParseFloat(strings.TrimSpace(row[1]), 64)
	if err != nil {
		return domain.Location{}, fmt.Errorf("line %d: bad lat %q: %w", line, row[1], err)
	}

	long, err := strconv.ParseFloat(strings.TrimSpace(row[2]), 64)
	if err != nil {
		return domain.Location{}, fmt.Errorf("line %d: bad long %q: %w", line, row[2], err)
	}

	cat := strings.TrimSpace(row[3])
	if cat == "" {
		return domain.Location{}, fmt.Errorf("line %d: empty category", line)
	}

	return domain.Location{ID: id, Lat: lat, Long: long, Category: cat}, nil
}
