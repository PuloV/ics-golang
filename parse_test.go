package ics

import (
	"testing"
	"time"
)

func TestCalendarInfo(t *testing.T) {
	calendar, err := ParseCalendar("testCalendars/2eventsCal.ics")
	if err != nil {
		t.Errorf("Failed to parse the calendar ( %s ) \n", err.Error())
	}

	if calendar.Name != "2 Events Cal" {
		t.Errorf("Expected name '%s' calendar , got '%s' calendars \n", "2 Events Cal", calendar.Name)
	}

	if calendar.Description != "The cal has 2 events(1st with attendees and second without)" {
		t.Errorf("Expected description '%s' calendar , got '%s' calendars \n", "The cal has 2 events(1st with attendees and second without)", calendar.Description)
	}

	if calendar.Version != 2.0 {
		t.Errorf("Expected version %s calendar , got %s calendars \n", 2.0, calendar.Version)
	}

	events := calendar.Events
	if len(events) != 2 {
		t.Errorf("Expected  %s events in calendar , got %s events \n", 2, len(events))
	}
}

func TestCalendarEvents(t *testing.T) {
	calendar, err := ParseCalendar("testCalendars/2eventsCal.ics")
	if err != nil {
		t.Errorf("Failed to parse the calendar ( %s ) \n", err.Error())
	}
	event := calendar.Events[0]
	start, _ := time.Parse(IcsFormat, "20140714T100000Z")
	end, _ := time.Parse(IcsFormat, "20140714T110000Z")
	created, _ := time.Parse(IcsFormat, "20140515T075711Z")
	modified, _ := time.Parse(IcsFormat, "20141125T074253Z")
	location := "In The Office"
	desc := "1. Report on previous weekly tasks. \\n2. Plan of the present weekly tasks."
	seq := 1
	status := "CONFIRMED"
	summary := "General Operative Meeting"
	rrule := ""
	attendeesCount := 3

	if event.Start != start {
		t.Errorf("Expected start %s , found %s  \n", start, event.Start)
	}

	if event.End != end {
		t.Errorf("Expected end %s , found %s  \n", end, event.End)
	}

	if event.Created != created {
		t.Errorf("Expected created %s , found %s  \n", created, event.Created)
	}

	if event.Modified != modified {
		t.Errorf("Expected modified %s , found %s  \n", modified, event.Modified)
	}

	if event.Location != location {
		t.Errorf("Expected location %s , found %s  \n", location, event.Location)
	}

	if event.Description != desc {
		t.Errorf("Expected description %s , found %s  \n", desc, event.Description)
	}

	if event.Sequence != seq {
		t.Errorf("Expected sequence %s , found %s  \n", seq, event.Sequence)
	}

	if event.Status != status {
		t.Errorf("Expected status %s , found %s  \n", status, event.Status)
	}

	if event.Summary != summary {
		t.Errorf("Expected status %s , found %s  \n", summary, event.Summary)
	}

	if event.RRule != rrule {
		t.Errorf("Expected rrule %s , found %s  \n", rrule, event.RRule)
	}

	if len(event.Attendees) != attendeesCount {
		t.Errorf("Expected attendeesCount %s , found %s  \n", attendeesCount, len(event.Attendees))
	}
}
