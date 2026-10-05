package index

import "proximity/internal/domain"

type Index interface {
	Candidates(q domain.Query) []domain.Location
}
