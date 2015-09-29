package ics

import "time"

type Calendar struct {
	Name        string
	Description string
	URL         string
	Version     float64
	Timezone    *time.Location
	Events      []Event
}

func NewCalendar() Calendar {
	return Calendar{
		Events: []Event{},
	}
}
