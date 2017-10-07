package ics

import (
	"crypto/md5"
	"fmt"
	"time"
)

type Event struct {
	Start         time.Time     `json:"start"`
	End           time.Time     `json:"end"`
	Created       time.Time     `json:"created"`
	Modified      time.Time     `json:"modified"`
	AlarmTime     time.Duration `json:"alarmTime"`
	ImportedID    string        `json:"importedID"`
	Status        string        `json:"status"`
	Description   string        `json:"description"`
	Location      string        `json:"location"`
	Summary       string        `json:"summary"`
	RRule         string        `json:"rrule"`
	Class         string        `json:"class"`
	ID            string        `json:"id"`
	Sequence      int           `json:"sequence"`
	WholeDayEvent bool          `json:"wholeDayEvent"`
	Attendees     []*Attendee   `json:"-"`
	Organizer     *Attendee     `json:"-"`
	InCalendar    *Calendar     `json:"-"`
	AlarmCallback func(*Event)  `json:"-"`
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
