package ics

import (
	"fmt"
)

type Attendee struct {
	name   string
	email  string
	status string
	role   string
	cutype string
}

func NewAttendee() *Attendee {
	a := new(Attendee)
	return a
}

func (a *Attendee) SetName(n string) *Attendee {
	a.name = n
	return a
}

func (a *Attendee) GetName() string {
	return a.name
}

func (a *Attendee) SetEmail(e string) *Attendee {
	a.email = e
	return a
}

func (a *Attendee) GetEmail() string {
	return a.email
}

func (a *Attendee) SetStatus(s string) *Attendee {
	a.status = s
	return a
}

func (a *Attendee) GetStatus() string {
	return a.status
}

func (a *Attendee) SetRole(r string) *Attendee {
	a.role = r
	return a
}

func (a *Attendee) GetRole() string {
	return a.role
}

func (a *Attendee) SetType(ct string) *Attendee {
	a.cutype = ct
	return a
}

func (a *Attendee) GetType() string {
	return a.cutype
}

func (a *Attendee) String() string {

	return fmt.Sprintf("%s with email %s", a.name, a.email)
}
