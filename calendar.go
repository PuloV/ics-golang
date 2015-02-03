package ics

import (
	"errors"
	"fmt"
	"time"
)

type Calendar struct {
	name         string
	description  string
	version      float64
	timezone     time.Location
	events       []Event
	eventsByDate map[string][]*Event
	eventByID    map[string]*Event
}

func NewCalendar() *Calendar {
	c := new(Calendar)
	// c.events = make([]Event)
	c.eventsByDate = make(map[string][]*Event)
	c.eventByID = make(map[string]*Event)
	return c
}

func (c *Calendar) SetName(n string) *Calendar {
	c.name = n
	return c
}

func (c *Calendar) GetName() string {
	return ""
}

func (c *Calendar) SetDesc(desc string) *Calendar {
	c.description = desc
	return c
}

func (c *Calendar) GetDesc() string {
	return ""
}

func (c *Calendar) SetVersion(ver float64) *Calendar {
	c.version = ver
	return c
}

func (c *Calendar) GetVersion() float64 {
	return c.version
}

func (c *Calendar) SetTimezone(tz time.Location) *Calendar {
	c.timezone = tz
	return c
}

func (c *Calendar) GetTimezone() time.Location {
	return c.timezone
}

//  add event to the calendar
func (c *Calendar) SetEvent(event Event) (*Calendar, error) {
	//  lock so that the events array doesn't change its size from other goruote
	mutex.Lock()

	// reference to the calendar
	if event.GetCalendar() == nil || event.GetCalendar() != c {
		event.SetCalendar(c)
	}
	// add the event to the main array with events
	c.events = append(c.events, event)

	// pointer to the added event in the main array
	eventPtr := &c.events[len(c.events)-1]

	// calculate the start day of the event
	eventStartTime := event.GetStart()
	tz := c.GetTimezone()
	eventDate := time.Date(eventStartTime.Year(), eventStartTime.Month(), eventStartTime.Day(), 0, 0, 0, 0, &tz)

	// faster search by date
	c.eventsByDate[eventDate.Format(YmdHis)] = append(c.eventsByDate[eventDate.Format(YmdHis)], eventPtr)
	// faster search by id
	c.eventByID[event.GetID()] = eventPtr

	mutex.Unlock()
	return c, nil
}

//  get event by id
func (c *Calendar) GetEventByID(eventID string) (*Event, error) {
	event, ok := c.eventByID[eventID]
	if ok {
		return event, nil
	}
	return nil, errors.New(fmt.Sprintf("There is no event with id %s", eventID))
}

//  get all events in the calendar
func (c *Calendar) GetEvents() []Event {
	return c.events
}

//  get all events in the calendar ordered by date
func (c *Calendar) GetEventsByDates() map[string][]*Event {
	return c.eventsByDate
}

// get all events for specified date
func (c *Calendar) GetEventsByDate(dateTime time.Time) ([]*Event, error) {
	tz := c.GetTimezone()
	day := time.Date(dateTime.Year(), dateTime.Month(), dateTime.Day(), 0, 0, 0, 0, &tz)
	events, ok := c.eventsByDate[day.Format(YmdHis)]
	if ok {
		return events, nil
	}
	return nil, errors.New(fmt.Sprintf("There are no events for the day %s", day.Format(YmdHis)))
}
