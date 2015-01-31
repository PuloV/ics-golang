package ics

import (
	"time"
)

type Calendar struct {
	name         string
	description  string
	version      float64
	timezone     time.Location
	events       []Event
	eventsByDate map[string][]Event
}

func NewCalendar() *Calendar {
	c := new(Calendar)
	// c.events = make([]Event)
	c.eventsByDate = make(map[string][]Event)
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

func (c *Calendar) SetEvent(events Event) (*Calendar, error) {
	return c, nil
}

func (c *Calendar) GetEvent() (*Event, error) {
	return new(Event), nil
}
