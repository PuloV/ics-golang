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
	summary       string
	rrule         string
	class         string
	id            int
	sequence      int
	attendees     []string
	alarmCallback func()
}

func NewEvent() *Event {
}

func (e *Event) SetStart(start string) *Event {
}

func (e *Event) GetStart() string {
}

func (e *Event) SetEnd(end string) *Event {
}

func (e *Event) GetEnd() string {
}

func (e *Event) SetID(id string) *Event {
}

func (e *Event) GetID() string {
}

func (e *Event) SetAttendee(email string) *Event {
}

func (e *Event) GetAttendees() string {
}

func (e *Event) SetClass(class string) *Event {
}

func (e *Event) GetClass() string {
}

func (e *Event) SetCreated(created string) *Event {
}

func (e *Event) GetCreated() string {
}

func (e *Event) SetLastModified(modified string) *Event {
}

func (e *Event) GetLastModified() string {
}

func (e *Event) SetSequence(sq string) *Event {
}

func (e *Event) GetSequence() string {
}

func (e *Event) SetStatus(status string) *Event {
}

func (e *Event) GetStatus() string {
}

func (e *Event) SetSummary(summary string) *Event {
}

func (e *Event) GetSummary() string {
}

func (e *Event) SetRRule(rrule string) *Event {
}

func (e *Event) GetRRule() string {
}

func (e *Event) Clone(string) *Event {
}

func (e *Event) SetAlarm(time string, callback func()) *Event {
}

func (e *Event) GetAlarm() string {
}
