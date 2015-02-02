package ics

import (
	"crypto/md5"
	"fmt"
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
	id            string
	sequence      int
	attendees     []string
	wholeDayEvent bool
	alarmCallback func()
}

func NewEvent() *Event {
	return new(Event)
}

func (e *Event) SetStart(start time.Time) *Event {
	e.start = start
	return e
}

func (e *Event) GetStart() time.Time {
	return e.start
}

func (e *Event) SetEnd(end time.Time) *Event {
	e.end = end
	return e
}

func (e *Event) GetEnd() time.Time {
	return e.end
}

func (e *Event) SetID(id string) *Event {
	e.id = id
	return e
}

func (e *Event) GetID() string {
	return e.id
}

func (e *Event) SetImportedID(id string) *Event {
	e.importedId = id
	return e
}

func (e *Event) GetImportedID() string {
	return e.importedId
}

func (e *Event) SetAttendee(email string) *Event {
	return e
}

func (e *Event) GetAttendees() string {
	return ""
}

func (e *Event) SetClass(class string) *Event {
	e.class = class
	return e
}

func (e *Event) GetClass() string {
	return e.class
}

func (e *Event) SetCreated(created time.Time) *Event {
	e.created = created
	return e
}

func (e *Event) GetCreated() time.Time {
	return e.created
}

func (e *Event) SetLastModified(modified time.Time) *Event {
	e.modified = modified
	return e
}

func (e *Event) GetLastModified() time.Time {
	return e.modified
}

func (e *Event) SetSequence(sq int) *Event {
	e.sequence = sq
	return e
}

func (e *Event) GetSequence() int {
	return e.sequence
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
	e.rrule = rrule
	return e
}

func (e *Event) GetRRule() string {
	return e.rrule
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

func (e *Event) SetWholeDayEvent(wholeDay bool) *Event {
	e.wholeDayEvent = wholeDay
	return e
}

func (e *Event) GetWholeDayEvent() bool {
	return e.wholeDayEvent
}

func (e *Event) IsWholeDay() bool {
	return e.wholeDayEvent
}

//  generates an unique id for the event
func (e *Event) GenerateEventId() string {
	if e.GetImportedID() != "" {
		toBeHashed := fmt.Sprintf("%s%s%s%s", e.GetStart(), e.GetEnd(), e.GetImportedID())
		return fmt.Sprintf("%x", md5.Sum(stringToByte(toBeHashed)))
	} else {
		toBeHashed := fmt.Sprintf("%s%s%s%s", e.GetStart(), e.GetEnd(), e.GetSummary(), e.GetDescription())
		return fmt.Sprintf("%x", md5.Sum(stringToByte(toBeHashed)))
	}

}
