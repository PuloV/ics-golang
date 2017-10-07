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
	Attendees     []*Attendee
	Organizer     *Attendee
	WholeDayEvent bool
	InCalendar    *Calendar
	AlarmCallback func(*Event)
}

func NewEvent() *Event {
	e := new(Event)
	e.Attendees = []*Attendee{}
	return e
}

func (e *Event) Clone() *Event {
	newE := *e
	return &newE
}

func (e *Event) SetAlarm(alarmAfter time.Duration, callback func(*Event)) *Event {
	e.AlarmCallback = callback
	e.AlarmTime = alarmAfter
	go func() {
		select {
		case <-time.After(alarmAfter):
			callback(e)
		}
	}()
	return e
}

//  generates an unique id for the event
func (e *Event) GenerateEventId() string {
	if e.ImportedID != "" {
		toBeHashed := fmt.Sprintf("%s%s%s%s", e.Start, e.End, e.ImportedID)
		return fmt.Sprintf("%x", md5.Sum(stringToByte(toBeHashed)))
	} else {
		toBeHashed := fmt.Sprintf("%s%s%s%s", e.Start, e.End, e.Summary, e.Description)
		return fmt.Sprintf("%x", md5.Sum(stringToByte(toBeHashed)))
	}

}

func (e *Event) String() string {
	from := e.Start.Format(YmdHis)
	to := e.End.Format(YmdHis)
	summ := e.Summary
	status := e.Status
	attendeeCount := len(e.Attendees)
	return fmt.Sprintf("Event(%s) from %s to %s about %s . %d people are invited to it", status, from, to, summ, attendeeCount)
}
