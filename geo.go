package ics

import (
	"strconv"
)

// Geo has latitude and longitude from the the GEO property of an event
type Geo struct {
	latStr string
	lat    *float64

	longStr string
	long    *float64
}

// NewGeo creates a new Geo object
func NewGeo(lat string, long string) *Geo {
	return &Geo{
		latStr:  lat,
		longStr: long,
	}
}

// Latitude returns the latitude value from Geo
func (g *Geo) Latitude() (float64, error) {
	if g.lat != nil {
		return *g.lat, nil
	}

	latVal, err := strconv.ParseFloat(g.latStr, 64)
	if err != nil {
		return 0, err
	}

	g.lat = &latVal
	return latVal, nil
}

// Longitude returns the longitude value from Geo
func (g *Geo) Longitude() (float64, error) {
	if g.long != nil {
		return *g.long, nil
	}

	longVal, err := strconv.ParseFloat(g.longStr, 64)
	if err != nil {
		return 0, err
	}

	g.long = &longVal
	return longVal, nil
}
