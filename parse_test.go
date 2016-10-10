package ics_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/PuloV/ics-golang"
)

func TestLoadCalendar(t *testing.T) {
	parser := ics.New()
	calBytes, err := ioutil.ReadFile("testCalendars/2eventsCal.ics")
	if err != nil {
		t.Errorf("Failed to read calendar file ( %s )", err)
	}

	parser.Load(string(calBytes))

	parseErrors, err := parser.GetErrors()
	if err != nil {
		t.Errorf("Failed to wait the parse of the calendars ( %s )", err)
	}
	for i, pErr := range parseErrors {
		t.Errorf("Parsing Error №%d: %s", i, pErr)
	}

	calendars, errCal := parser.GetCalendars()
	if errCal != nil {
		t.Errorf("Failed to get calendars ( %s )", errCal)
	}
	if len(calendars) != 1 {
		t.Errorf("Expected 1 calendar, found %d calendars", len(calendars))
	}
}

func TestNewParser(t *testing.T) {
	parser := ics.New()
	rType := fmt.Sprintf("%v", reflect.TypeOf(parser))
	if rType != "*ics.Parser" {
		t.Errorf("Failed to create a Parser !")
	}
}

func TestNewParserChans(t *testing.T) {
	parser := ics.New()
	input := parser.GetInputChan()
	output := parser.GetOutputChan()

	rType := fmt.Sprintf("%v", reflect.TypeOf(input))

	if rType != "chan string" {
		t.Errorf("Failed to create a input chan! Received: Type %s Value %s", rType, input)
	}

	rType = fmt.Sprintf("%v", reflect.TypeOf(output))
	if rType != "chan *ics.Event" {
		t.Errorf("Failed to create a output chan! Received: Type %s Value %s", rType, output)
	}
}

func TestParsing0Calendars(t *testing.T) {
	parser := ics.New()
	parser.Wait()

	parseErrors, err := parser.GetErrors()

	if err != nil {
		t.Errorf("Failed to wait the parse of the calendars ( %s )", err)
	}
	for i, pErr := range parseErrors {
		t.Errorf("Parsing Error №%d: %s", i, pErr)
	}
}

func TestParsing1Calendars(t *testing.T) {
	parser := ics.New()
	input := parser.GetInputChan()
	input <- "testCalendars/2eventsCal.ics"
	parser.Wait()

	parseErrors, err := parser.GetErrors()

	if err != nil {
		t.Errorf("Failed to wait the parse of the calendars ( %s )", err)
	}
	for i, pErr := range parseErrors {
		t.Errorf("Parsing Error №%d: %s", i, pErr)
	}

	calendars, errCal := parser.GetCalendars()

	if errCal != nil {
		t.Errorf("Failed to get calendars ( %s )", errCal)
	}

	if len(calendars) != 1 {
		t.Errorf("Expected 1 calendar, found %d calendars", len(calendars))
	}

}

func TestParsing2Calendars(t *testing.T) {
	parser := ics.New()
	input := parser.GetInputChan()
	input <- "testCalendars/2eventsCal.ics"
	input <- "testCalendars/3eventsNoAttendee.ics"
	parser.Wait()

	parseErrors, err := parser.GetErrors()

	if err != nil {
		t.Errorf("Failed to wait the parse of the calendars ( %s )", err)
	}
	for i, pErr := range parseErrors {
		t.Errorf("Parsing Error №%d: %s", i, pErr)
	}

	calendars, errCal := parser.GetCalendars()

	if errCal != nil {
		t.Errorf("Failed to get calendars ( %s )", errCal)
	}

	if len(calendars) != 2 {
		t.Errorf("Expected 1 calendar, found %d calendars", len(calendars))
	}

}

func TestParsingNotExistingCalendar(t *testing.T) {
	parser := ics.New()
	input := parser.GetInputChan()
	input <- "testCalendars/notFound.ics"
	parser.Wait()

	parseErrors, err := parser.GetErrors()

	if err != nil {
		t.Errorf("Failed to wait the parse of the calendars ( %s )", err)
	}
	if len(parseErrors) != 1 {
		t.Errorf("Expected 1 error, found %d in :\n  %#v", len(parseErrors), parseErrors)
	}

}

func TestParsingNotExistingAndExistingCalendars(t *testing.T) {
	parser := ics.New()
	input := parser.GetInputChan()
	input <- "testCalendars/3eventsNoAttendee.ics"
	input <- "testCalendars/notFound.ics"
	parser.Wait()

	parseErrors, err := parser.GetErrors()

	if err != nil {
		t.Errorf("Failed to wait the parse of the calendars ( %s )", err)
	}
	if len(parseErrors) != 1 {
		t.Errorf("Expected 1 error, found %d in :\n  %#v", len(parseErrors), parseErrors)
	}

	calendars, errCal := parser.GetCalendars()

	if errCal != nil {
		t.Errorf("Failed to get calendars ( %s )", errCal)
	}

	if len(calendars) != 1 {
		t.Errorf("Expected 1 calendar, found %d calendars", len(calendars))
	}

}
func TestParsingWrongCalendarUrls(t *testing.T) {
	parser := ics.New()
	input := parser.GetInputChan()
	input <- "http://localhost/goTestFails"
	parser.Wait()

	parseErrors, err := parser.GetErrors()

	if err != nil {
		t.Errorf("Failed to wait the parse of the calendars ( %s )", err)
	}
	if len(parseErrors) != 1 {
		t.Errorf("Expected 1 error, found %d in :\n  %#v", len(parseErrors), parseErrors)
	}

	calendars, errCal := parser.GetCalendars()

	if errCal != nil {
		t.Errorf("Failed to get calendars ( %s )", errCal)
	}

	if len(calendars) != 0 {
		t.Errorf("Expected 0 calendar, found %d calendars", len(calendars))
	}
}

func TestCreatingTempDir(t *testing.T) {
	ics.FilePath = "testingTempDir/"
	parser := ics.New()
	input := parser.GetInputChan()
	input <- "https://www.google.com/calendar/ical/yordanpulov%40gmail.com/private-81525ac0eb14cdc2e858c15e1b296a1c/basic.ics"
	parser.Wait()
	_, err := os.Stat(ics.FilePath)
	if err != nil {
		t.Errorf("Failed to create %s", ics.FilePath)
	}
	// remove the new dir
	os.Remove(ics.FilePath)
	// return the var to default
	ics.FilePath = "tmp/"
}

func TestCalendarInfo(t *testing.T) {
	parser := ics.New()
	input := parser.GetInputChan()
	input <- "testCalendars/2eventsCal.ics"
	parser.Wait()

	parseErrors, err := parser.GetErrors()

	if err != nil {
		t.Errorf("Failed to wait the parse of the calendars ( %s )", err)
	}
	if len(parseErrors) != 0 {
		t.Errorf("Expected 0 error, found %d in :\n %#v", len(parseErrors), parseErrors)
	}

	calendars, errCal := parser.GetCalendars()

	if errCal != nil {
		t.Errorf("Failed to get calendars ( %s )", errCal)
	}

	if len(calendars) != 1 {
		t.Errorf("Expected 1 calendar, found %d calendars", len(calendars))
		return
	}

	calendar := calendars[0]

	if calendar.GetName() != "2 Events Cal" {
		t.Errorf("Expected name '%s' calendar, got '%s' calendars", "2 Events Cal", calendar.GetName())
	}

	if calendar.GetDesc() != "The cal has 2 events(1st with attendees and second without)" {
		t.Errorf("Expected description '%s' calendar, got '%s' calendars", "The cal has 2 events(1st with attendees and second without)", calendar.GetDesc())
	}

	if calendar.GetVersion() != 2.0 {
		t.Errorf("Expected version %v calendar, got %v calendars", 2.0, calendar.GetVersion())
	}

	events := calendar.GetEvents()
	if len(events) != 2 {
		t.Errorf("Expected %d events in calendar, got %d events", 2, len(events))
	}

	eventsByDates := calendar.GetEventsByDates()
	if len(eventsByDates) != 2 {
		t.Errorf("Expected %d events by date in calendar, got %d events", 2, len(eventsByDates))
	}

	geometryExamIcsFormat, errICS := time.Parse(ics.IcsFormat, "20140616T060000Z")
	if err != nil {
		t.Errorf("(ics time format) Unexpected error %s", errICS)
	}

	geometryExamYmdHis, errYMD := time.Parse(ics.YmdHis, "2014-06-16 06:00:00")
	if err != nil {
		t.Errorf("(YmdHis time format) Unexpected error %s", errYMD)
	}
	eventsByDate, err := calendar.GetEventsByDate(geometryExamIcsFormat)
	if err != nil {
		t.Errorf("(ics time format) Unexpected error %s", err)
	}
	if len(eventsByDate) != 1 {
		t.Errorf("(ics time format) Expected %d events in calendar for the date 2014-06-16, got %d events", 1, len(eventsByDate))
	}

	eventsByDate, err = calendar.GetEventsByDate(geometryExamYmdHis)
	if err != nil {
		t.Errorf("(YmdHis time format) Unexpected error %s", err)
	}
	if len(eventsByDate) != 1 {
		t.Errorf("(YmdHis time format) Expected %d events in calendar for the date 2014-06-16, got %d events", 1, len(eventsByDate))
	}

}

func TestCalendarEvents(t *testing.T) {
	parser := ics.New()
	input := parser.GetInputChan()
	input <- "testCalendars/2eventsCal.ics"
	parser.Wait()

	parseErrors, err := parser.GetErrors()

	if err != nil {
		t.Errorf("Failed to wait the parse of the calendars ( %s )", err)
	}
	if len(parseErrors) != 0 {
		t.Errorf("Expected 0 error, found %d in :\n  %#v", len(parseErrors), parseErrors)
	}

	calendars, errCal := parser.GetCalendars()

	if errCal != nil {
		t.Errorf("Failed to get calendars ( %s )", errCal)
	}

	if len(calendars) != 1 {
		t.Errorf("Expected 1 calendar, found %d calendars", len(calendars))
		return
	}

	calendar := calendars[0]
	event, err := calendar.GetEventByImportedID("btb9tnpcnd4ng9rn31rdo0irn8@google.com")
	if err != nil {
		t.Errorf("Failed to get event by id with error %s", err)
	}

	//  event must have
	start, _ := time.Parse(ics.IcsFormat, "20140714T100000Z")
	end, _ := time.Parse(ics.IcsFormat, "20140714T110000Z")
	created, _ := time.Parse(ics.IcsFormat, "20140515T075711Z")
	modified, _ := time.Parse(ics.IcsFormat, "20141125T074253Z")
	location := "In The Office"
	desc := "1. Report on previous weekly tasks. \\n2. Plan of the present weekly tasks."
	seq := 1
	status := "CONFIRMED"
	summary := "General Operative Meeting"
	rrule := ""
	attendeesCount := 3

	org := new(ics.Attendee)
	org.SetName("r.chupetlovska@gmail.com")
	org.SetEmail("r.chupetlovska@gmail.com")

	if event.GetStart() != start {
		t.Errorf("Expected start %s, found %s", start, event.GetStart())
	}

	if event.GetEnd() != end {
		t.Errorf("Expected end %s, found %s", end, event.GetEnd())
	}

	if event.GetCreated() != created {
		t.Errorf("Expected created %s, found %s", created, event.GetCreated())
	}

	if event.GetLastModified() != modified {
		t.Errorf("Expected modified %s, found %s", modified, event.GetLastModified())
	}

	if event.GetLocation() != location {
		t.Errorf("Expected location %s, found %s", location, event.GetLocation())
	}

	if event.GetDescription() != desc {
		t.Errorf("Expected description %s, found %s", desc, event.GetDescription())
	}

	if event.GetSequence() != seq {
		t.Errorf("Expected sequence %s, found %s", seq, event.GetSequence())
	}

	if event.GetStatus() != status {
		t.Errorf("Expected status %s, found %s", status, event.GetStatus())
	}

	if event.GetSummary() != summary {
		t.Errorf("Expected status %s, found %s", summary, event.GetSummary())
	}

	if event.GetRRule() != rrule {
		t.Errorf("Expected rrule %s, found %s", rrule, event.GetRRule())
	}

	if len(event.GetAttendees()) != attendeesCount {
		t.Errorf("Expected attendeesCount %s, found %s", attendeesCount, len(event.GetAttendees()))
	}

	eventOrg := event.GetOrganizer()
	if *eventOrg != *org {
		t.Errorf("Expected organizer %s, found %s", org, event.GetOrganizer())
	}

	// SECOND EVENT WITHOUT ATTENDEES AND ORGANIZER
	eventNoAttendees, errNoAttendees := calendar.GetEventByImportedID("mhhesb7si5968njvthgbiub7nk@google.com")
	attendeesCount = 0
	org = new(ics.Attendee)

	if errNoAttendees != nil {
		t.Errorf("Failed to get event by id with error %s", errNoAttendees)
	}

	if len(eventNoAttendees.GetAttendees()) != attendeesCount {
		t.Errorf("Expected attendeesCount %s, found %s", attendeesCount, len(event.GetAttendees()))
	}

	if eventNoAttendees.GetOrganizer() != nil {
		t.Errorf("Expected organizer %s, found %s", org, eventNoAttendees.GetOrganizer())
	}
}

func TestCalendarEventAttendees(t *testing.T) {
	parser := ics.New()
	input := parser.GetInputChan()
	input <- "testCalendars/2eventsCal.ics"
	parser.Wait()

	parseErrors, err := parser.GetErrors()

	if err != nil {
		t.Errorf("Failed to wait the parse of the calendars ( %s )", err)
	}
	if len(parseErrors) != 0 {
		t.Errorf("Expected 0 error, found %d in :\n  %#v", len(parseErrors), parseErrors)
	}

	calendars, errCal := parser.GetCalendars()

	if errCal != nil {
		t.Errorf("Failed to get calendars ( %s )", errCal)
	}

	if len(calendars) != 1 {
		t.Errorf("Expected 1 calendar, found %d calendars", len(calendars))
		return
	}

	calendar := calendars[0]
	event, err := calendar.GetEventByImportedID("btb9tnpcnd4ng9rn31rdo0irn8@google.com")
	if err != nil {
		t.Errorf("Failed to get event by id with error %s", err)
	}
	attendees := event.GetAttendees()
	attendeesCount := 3

	if len(attendees) != attendeesCount {
		t.Errorf("Expected attendeesCount %s, found %s", attendeesCount, len(attendees))
		return
	}

	john := attendees[0]
	sue := attendees[1]
	travis := attendees[2]

	// check name
	if john.GetName() != "John Smith" {
		t.Errorf("Expected attendee name %s, found %s", "John Smith", john.GetName())
	}
	if sue.GetName() != "Sue Zimmermann" {
		t.Errorf("Expected attendee name %s, found %s", "Sue Zimmermann", sue.GetName())
	}
	if travis.GetName() != "Travis M. Vollmer" {
		t.Errorf("Expected attendee name %s, found %s", "Travis M. Vollmer", travis.GetName())
	}

	// check email
	if john.GetEmail() != "j.smith@gmail.com" {
		t.Errorf("Expected attendee email %s, found %s", "j.smith@gmail.com", john.GetEmail())
	}
	if sue.GetEmail() != "SueMZimmermann@dayrep.com" {
		t.Errorf("Expected attendee email %s, found %s", "SueMZimmermann@dayrep.com", sue.GetEmail())
	}
	if travis.GetEmail() != "travis@dayrep.com" {
		t.Errorf("Expected attendee email %s, found %s", "travis@dayrep.com", travis.GetEmail())
	}

	// check status
	if john.GetStatus() != "ACCEPTED" {
		t.Errorf("Expected attendee status %s, found %s", "ACCEPTED", john.GetStatus())
	}
	if sue.GetStatus() != "NEEDS-ACTION" {
		t.Errorf("Expected attendee status %s, found %s", "NEEDS-ACTION", sue.GetStatus())
	}
	if travis.GetStatus() != "NEEDS-ACTION" {
		t.Errorf("Expected attendee status %s, found %s", "NEEDS-ACTION", travis.GetStatus())
	}

	// check role
	if john.GetRole() != "REQ-PARTICIPANT" {
		t.Errorf("Expected attendee status %s, found %s", "REQ-PARTICIPANT", john.GetRole())
	}
	if sue.GetRole() != "REQ-PARTICIPANT" {
		t.Errorf("Expected attendee status %s, found %s", "REQ-PARTICIPANT", sue.GetRole())
	}
	if travis.GetRole() != "REQ-PARTICIPANT" {
		t.Errorf("Expected attendee status %s, found %s", "REQ-PARTICIPANT", travis.GetRole())
	}
}

func TestCalendarMultidayEvent(t *testing.T) {
	parser := ics.New()
	input := parser.GetInputChan()
	input <- "testCalendars/multiday.ics"
	parser.Wait()

	parseErrors, err := parser.GetErrors()
	if err != nil {
		t.Errorf("Failed to wait the parse of the calendars ( %s )", err)
	}
	if len(parseErrors) != 0 {
		t.Errorf("Expected 0 error, found %d in :\n  %#v", len(parseErrors), parseErrors)
	}

	calendars, errCal := parser.GetCalendars()
	if errCal != nil {
		t.Errorf("Failed to get calendars ( %s )", errCal)
	}

	if len(calendars) != 1 {
		t.Errorf("Expected 1 calendar, found %d calendars", len(calendars))
		return
	}

	calendar := calendars[0]

	// Test a day before the start day
	events, err := calendar.GetEventsByDate(time.Date(2016, 8, 31, 0, 0, 0, 0, time.UTC))
	if err == nil {
		t.Errorf("Expected no event before the start day, got %d", len(events))
	}

	// Test exact start day
	events, err = calendar.GetEventsByDate(time.Date(2016, 9, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Errorf("Failed to get event: %s", err.Error())
	}
	if len(events) != 1 {
		t.Errorf("Expected 1 event on the start day, got %d", len(events))
	}

	// Test a random day between start and end date
	events, err = calendar.GetEventsByDate(time.Date(2016, 10, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Errorf("Failed to get event: %s", err.Error())
	}
	if len(events) != 1 {
		t.Errorf("Expected 1 event between start and end, got %d", len(events))
	}

	// Test a day after the end day
	events, err = calendar.GetEventsByDate(time.Date(2016, 11, 1, 0, 0, 0, 0, time.UTC))
	if err == nil {
		t.Errorf("Expected no event after the end day, got %d", len(events))
	}
}
