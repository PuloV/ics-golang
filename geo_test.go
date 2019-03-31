package ics

import (
	"testing"
)

func TestLongitudeReturnsError(t *testing.T) {
	geo := NewGeo("1", "badLong")
	_, err := geo.Longitude()

	if err == nil {
		t.Error("Expected error when getting Longitude but err was nil.")
	}
}

func TestLongitudeReturnsValue(t *testing.T) {
	geo := NewGeo("1", "123")
	long, err := geo.Longitude()

	if err != nil {
		t.Error("Error when getting longitude.")
	}

	if long != 123 {
		t.Errorf("Expected longitude value 123, but received %v", long)
	}
}

func TestLatitudeReturnsError(t *testing.T) {
	geo := NewGeo("badlat", "1")
	_, err := geo.Latitude()

	if err == nil {
		t.Error("Expected error when getting Latitude but err was nil.")
	}
}

func TestLatitudeReturnsValue(t *testing.T) {
	geo := NewGeo("321", "1")
	lat, err := geo.Latitude()

	if err != nil {
		t.Error("Error when getting latitude.")
	}

	if lat != 321 {
		t.Errorf("Expected latitude value 321, but received %v", lat)
	}
}
