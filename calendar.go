package ics

type Calendar struct {
	name         string
	description  string
	version      string
	timezone     string
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

func (c *Calendar) SetVersion(ver string) *Calendar {
	return c
}

func (c *Calendar) GetVersion() string {
	return ""
}

func (c *Calendar) SetTimezone(tz string) *Calendar {
	return c
}

func (c *Calendar) GetTimezone() string {
	return ""
}

func (c *Calendar) SetEvent(events Event) (*Calendar, error) {
	return c, nil
}

func (c *Calendar) GetEvent() (*Event, error) {
	return new(Event), nil
}
