package domain

type Location struct {
	ID       int
	Lat      float64
	Long     float64
	Category string
}

type Query struct {
	Lat      float64
	Long     float64
	Category string
	Radius   float64
}

type Scored struct {
	Location Location
	Distance float64
}
