package ics

type Event struct {
}

type Calendar struct {
	name         string
	description  string
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
func (c *Calendar) SetDesc(desc string) *Calendar {
	c.description = desc
	return c
}
