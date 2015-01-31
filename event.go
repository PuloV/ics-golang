package ics

import (
	"time"
)

type Event struct {
	start         time.Time
	end           time.Time
	created       time.Time
	modified      time.Time
	alarmTime     time.Time
	importedId    string
	status        string
	description   string
	summary       string
	rrule         string
	class         string
	id            int
	sequence      int
	attendees     []string
	alarmCallback func()
}

func NewEvent() *Event {
	return new(Event)
}

func (e *Event) SetStart(start string) *Event {
	return e
}

func (e *Event) GetStart() string {
	return ""
}

func (e *Event) SetEnd(end string) *Event {
	return e
}

func (e *Event) GetEnd() string {
	return ""
}

func (e *Event) SetID(id string) *Event {
	return e
}

func (e *Event) GetID() string {
	return ""
}

func (e *Event) SetAttendee(email string) *Event {
	return e
}

func (e *Event) GetAttendees() string {
	return ""
}

func (e *Event) SetClass(class string) *Event {
	return e
}

func (e *Event) GetClass() string {
	return ""
}

func (e *Event) SetCreated(created string) *Event {
	return e
}

func (e *Event) GetCreated() string {
	return ""
}

func (e *Event) SetLastModified(modified string) *Event {
	return e
}

func (e *Event) GetLastModified() string {
	return ""
}

func (e *Event) SetSequence(sq string) *Event {
	return e
}

func (e *Event) GetSequence() string {
	return ""
}

func (e *Event) SetStatus(status string) *Event {
	e.status = status
	return e
}

func (e *Event) GetStatus() string {
	return e.status
}

func (e *Event) SetSummary(summary string) *Event {
	e.summary = summary
	return e
}

func (e *Event) GetSummary() string {
	return e.summary
}

func (e *Event) SetDescription(description string) *Event {
	e.description = description
	return e
}

func (e *Event) GetDescription() string {
	return e.description
}

func (e *Event) SetRRule(rrule string) *Event {
	return e
}

func (e *Event) GetRRule() string {
	return ""
}

func (e *Event) Clone(string) *Event {
	return e
}

func (e *Event) SetAlarm(time string, callback func()) *Event {
	return e
}

func (e *Event) GetAlarm() string {
	return ""
}
