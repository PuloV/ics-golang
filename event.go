package ics

import (
	"crypto/md5"
	"fmt"
	"time"
)

type Event struct {
	Start         time.Time
	End           time.Time
	Created       time.Time
	Modified      time.Time
	AlarmTime     time.Duration
	ImportedID    string
	Status        string
	Description   string
	Location      string
	Summary       string
	RRule         string
	Class         string
	ID            string
	Sequence      int
	Attendees     []Attendee
	Organizer     Attendee
	WholeDayEvent bool
}

func NewEvent() *Event {
	return &Event{
		Attendees: []Attendee{},
	}
}

func (e *Event) Clone() *Event {
	newEvent := *e
	return &newEvent
}

func (e *Event) GenerateEventId() string {
	if e.ImportedID != "" {
		toBeHashed := fmt.Sprintf("%s%s%s%s", e.Start, e.End, e.ImportedID)
		return fmt.Sprintf("%x", md5.Sum([]byte(toBeHashed)))
	} else {
		toBeHashed := fmt.Sprintf("%s%s%s%s", e.Start, e.End, e.Summary, e.Description)
		return fmt.Sprintf("%x", md5.Sum([]byte(toBeHashed)))
	}
}
