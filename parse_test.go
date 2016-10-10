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
		t.Errorf("Failed to create a input chan ! Received : Type %s Value %s", rType, input)
	}

	rType = fmt.Sprintf("%v", reflect.TypeOf(output))
	if rType != "chan *ics.Event" {
		t.Errorf("Failed to create a output chan! Received : Type %s Value %s", rType, output)
	}
}

func TestParsing0Calendars(t *testing.T) {
	parser := ics.New()
	parser.Wait()

	parseErrors, err := parser.GetErrors()

	if err != nil {
		t.Errorf("Failed to wait the parse of the calendars ( %s ) \n", err)
	}
	for i, pErr := range parseErrors {
		t.Errorf("Parsing Error №%d : %s  \n", i, pErr)
	}
}

func TestParsing1Calendars(t *testing.T) {
	parser := ics.New()
	input := parser.GetInputChan()
	input <- "testCalendars/2eventsCal.ics"
	parser.Wait()

	parseErrors, err := parser.GetErrors()

	if err != nil {
		t.Errorf("Failed to wait the parse of the calendars ( %s ) \n", err)
	}
	for i, pErr := range parseErrors {
		t.Errorf("Parsing Error №%d : %s  \n", i, pErr)
	}

	calendars, errCal := parser.GetCalendars()

	if errCal != nil {
		t.Errorf("Failed to get calendars ( %s ) \n", errCal)
	}

	if len(calendars) != 1 {
		t.Errorf("Expected 1 calendar , found %d calendars \n", len(calendars))
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
		t.Errorf("Failed to wait the parse of the calendars ( %s ) \n", err)
	}
	for i, pErr := range parseErrors {
		t.Errorf("Parsing Error №%d : %s  \n", i, pErr)
	}

	calendars, errCal := parser.GetCalendars()

	if errCal != nil {
		t.Errorf("Failed to get calendars ( %s ) \n", errCal)
	}

	if len(calendars) != 2 {
		t.Errorf("Expected 1 calendar , found %d calendars \n", len(calendars))
	}

}

func TestParsingNotExistingCalendar(t *testing.T) {
	parser := ics.New()
	input := parser.GetInputChan()
	input <- "testCalendars/notFound.ics"
	parser.Wait()

	parseErrors, err := parser.GetErrors()

	if err != nil {
		t.Errorf("Failed to wait the parse of the calendars ( %s ) \n", err)
	}
	if len(parseErrors) != 1 {
		t.Errorf("Expected 1 error , found %d in :\n  %#v  \n", len(parseErrors), parseErrors)
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
		t.Errorf("Failed to wait the parse of the calendars ( %s ) \n", err)
	}
	if len(parseErrors) != 1 {
		t.Errorf("Expected 1 error , found %d in :\n  %#v  \n", len(parseErrors), parseErrors)
	}

	calendars, errCal := parser.GetCalendars()

	if errCal != nil {
		t.Errorf("Failed to get calendars ( %s ) \n", errCal)
	}

	if len(calendars) != 1 {
		t.Errorf("Expected 1 calendar , found %d calendars \n", len(calendars))
	}

}
func TestParsingWrongCalendarUrls(t *testing.T) {
	parser := ics.New()
	input := parser.GetInputChan()
	input <- "http://localhost/goTestFails"
	parser.Wait()

	parseErrors, err := parser.GetErrors()

	if err != nil {
		t.Errorf("Failed to wait the parse of the calendars ( %s ) \n", err)
	}
	if len(parseErrors) != 1 {
		t.Errorf("Expected 1 error , found %d in :\n  %#v  \n", len(parseErrors), parseErrors)
	}

	calendars, errCal := parser.GetCalendars()

	if errCal != nil {
		t.Errorf("Failed to get calendars ( %s ) \n", errCal)
	}

	if len(calendars) != 0 {
		t.Errorf("Expected 0 calendar , found %d calendars \n", len(calendars))
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
		t.Errorf("Failed to create %s  \n", ics.FilePath)
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
		t.Errorf("Failed to wait the parse of the calendars ( %s ) \n", err)
	}
	if len(parseErrors) != 0 {
		t.Errorf("Expected 0 error , found %d in :\n  %#v  \n", len(parseErrors), parseErrors)
	}

	calendars, errCal := parser.GetCalendars()

	if errCal != nil {
		t.Errorf("Failed to get calendars ( %s ) \n", errCal)
	}

	if len(calendars) != 1 {
		t.Errorf("Expected 1 calendar , found %d calendars \n", len(calendars))
		return
	}

	calendar := calendars[0]

	if calendar.GetName() != "2 Events Cal" {
		t.Errorf("Expected name '%s' calendar , got '%s' calendars \n", "2 Events Cal", calendar.GetName())
	}

	if calendar.GetDesc() != "The cal has 2 events(1st with attendees and second without)" {
		t.Errorf("Expected description '%s' calendar , got '%s' calendars \n", "The cal has 2 events(1st with attendees and second without)", calendar.GetDesc())
	}

	if calendar.GetVersion() != 2.0 {
		t.Errorf("Expected version %s calendar , got %s calendars \n", 2.0, calendar.GetVersion())
	}

	events := calendar.GetEvents()
	if len(events) != 2 {
		t.Errorf("Expected  %s events in calendar , got %s events \n", 2, len(events))
	}

	eventsByDates := calendar.GetEventsByDates()
	if len(eventsByDates) != 2 {
		t.Errorf("Expected  %s events in calendar , got %s events \n", 2, len(eventsByDates))
	}

	geometryExamIcsFormat, errICS := time.Parse(ics.IcsFormat, "20140616T060000Z")
	if err != nil {
		t.Errorf("(ics time format) Unexpected error %s \n", errICS)
	}

	geometryExamYmdHis, errYMD := time.Parse(ics.YmdHis, "2014-06-16 06:00:00")
	if err != nil {
		t.Errorf("(YmdHis time format) Unexpected error %s \n", errYMD)
	}
	eventsByDate, err := calendar.GetEventsByDate(geometryExamIcsFormat)
	if err != nil {
		t.Errorf("(ics time format) Unexpected error %s \n", err)
	}
	if len(eventsByDate) != 1 {
		t.Errorf("(ics time format) Expected  %s events in calendar for the date 2014-06-16 , got %s events \n", 1, len(eventsByDate))
	}

	eventsByDate, err = calendar.GetEventsByDate(geometryExamYmdHis)
	if err != nil {
		t.Errorf("(YmdHis time format) Unexpected error %s \n", err)
	}
	if len(eventsByDate) != 1 {
		t.Errorf("(YmdHis time format) Expected  %s events in calendar for the date 2014-06-16 , got %s events \n", 1, len(eventsByDate))
	}

}

func TestCalendarEvents(t *testing.T) {
	parser := ics.New()
	input := parser.GetInputChan()
	input <- "testCalendars/2eventsCal.ics"
	parser.Wait()

	parseErrors, err := parser.GetErrors()

	if err != nil {
		t.Errorf("Failed to wait the parse of the calendars ( %s ) \n", err)
	}
	if len(parseErrors) != 0 {
		t.Errorf("Expected 0 error , found %d in :\n  %#v  \n", len(parseErrors), parseErrors)
	}

	calendars, errCal := parser.GetCalendars()

	if errCal != nil {
		t.Errorf("Failed to get calendars ( %s ) \n", errCal)
	}

	if len(calendars) != 1 {
		t.Errorf("Expected 1 calendar , found %d calendars \n", len(calendars))
		return
	}

	calendar := calendars[0]
	event, err := calendar.GetEventByImportedID("btb9tnpcnd4ng9rn31rdo0irn8@google.com")
	if err != nil {
		t.Errorf("Failed to get event by id with error %s \n", err)
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
		t.Errorf("Expected start %s , found %s  \n", start, event.GetStart())
	}

	if event.GetEnd() != end {
		t.Errorf("Expected end %s , found %s  \n", end, event.GetEnd())
	}

	if event.GetCreated() != created {
		t.Errorf("Expected created %s , found %s  \n", created, event.GetCreated())
	}

	if event.GetLastModified() != modified {
		t.Errorf("Expected modified %s , found %s  \n", modified, event.GetLastModified())
	}

	if event.GetLocation() != location {
		t.Errorf("Expected location %s , found %s  \n", location, event.GetLocation())
	}

	if event.GetDescription() != desc {
		t.Errorf("Expected description %s , found %s  \n", desc, event.GetDescription())
	}

	if event.GetSequence() != seq {
		t.Errorf("Expected sequence %s , found %s  \n", seq, event.GetSequence())
	}

	if event.GetStatus() != status {
		t.Errorf("Expected status %s , found %s  \n", status, event.GetStatus())
	}

	if event.GetSummary() != summary {
		t.Errorf("Expected status %s , found %s  \n", summary, event.GetSummary())
	}

	if event.GetRRule() != rrule {
		t.Errorf("Expected rrule %s , found %s  \n", rrule, event.GetRRule())
	}

	if len(event.GetAttendees()) != attendeesCount {
		t.Errorf("Expected attendeesCount %s , found %s  \n", attendeesCount, len(event.GetAttendees()))
	}

	eventOrg := event.GetOrganizer()
	if *eventOrg != *org {
		t.Errorf("Expected organizer %s , found %s  \n", org, event.GetOrganizer())
	}

	// SECOND EVENT WITHOUT ATTENDEES AND ORGANIZER

	eventNoAttendees, errNoAttendees := calendar.GetEventByImportedID("mhhesb7si5968njvthgbiub7nk@google.com")
	attendeesCount = 0
	org = new(ics.Attendee)

	if errNoAttendees != nil {
		t.Errorf("Failed to get event by id with error %s \n", errNoAttendees)
	}

	if len(eventNoAttendees.GetAttendees()) != attendeesCount {
		t.Errorf("Expected attendeesCount %s , found %s  \n", attendeesCount, len(event.GetAttendees()))
	}

	if eventNoAttendees.GetOrganizer() != nil {
		t.Errorf("Expected organizer %s , found %s  \n", org, eventNoAttendees.GetOrganizer())
	}
}

func TestCalendarEventAttendees(t *testing.T) {
	parser := ics.New()
	input := parser.GetInputChan()
	input <- "testCalendars/2eventsCal.ics"
	parser.Wait()

	parseErrors, err := parser.GetErrors()

	if err != nil {
		t.Errorf("Failed to wait the parse of the calendars ( %s ) \n", err)
	}
	if len(parseErrors) != 0 {
		t.Errorf("Expected 0 error , found %d in :\n  %#v  \n", len(parseErrors), parseErrors)
	}

	calendars, errCal := parser.GetCalendars()

	if errCal != nil {
		t.Errorf("Failed to get calendars ( %s ) \n", errCal)
	}

	if len(calendars) != 1 {
		t.Errorf("Expected 1 calendar , found %d calendars \n", len(calendars))
		return
	}

	calendar := calendars[0]
	event, err := calendar.GetEventByImportedID("btb9tnpcnd4ng9rn31rdo0irn8@google.com")
	if err != nil {
		t.Errorf("Failed to get event by id with error %s \n", err)
	}
	attendees := event.GetAttendees()
	attendeesCount := 3

	if len(attendees) != attendeesCount {
		t.Errorf("Expected attendeesCount %s , found %s  \n", attendeesCount, len(attendees))
		return
	}

	john := attendees[0]
	sue := attendees[1]
	travis := attendees[2]

	// check name
	if john.GetName() != "John Smith" {
		t.Errorf("Expected attendee name %s , found %s  \n", "John Smith", john.GetName())

	}
	if sue.GetName() != "Sue Zimmermann" {
		t.Errorf("Expected attendee name %s , found %s  \n", "Sue Zimmermann", sue.GetName())

	}
	if travis.GetName() != "Travis M. Vollmer" {
		t.Errorf("Expected attendee name %s , found %s  \n", "Travis M. Vollmer", travis.GetName())

	}

	// check email
	if john.GetEmail() != "j.smith@gmail.com" {
		t.Errorf("Expected attendee email %s , found %s  \n", "j.smith@gmail.com", john.GetEmail())

	}
	if sue.GetEmail() != "SueMZimmermann@dayrep.com" {
		t.Errorf("Expected attendee email %s , found %s  \n", "SueMZimmermann@dayrep.com", sue.GetEmail())

	}
	if travis.GetEmail() != "travis@dayrep.com" {
		t.Errorf("Expected attendee email %s , found %s  \n", "travis@dayrep.com", travis.GetEmail())

	}

	// check status
	if john.GetStatus() != "ACCEPTED" {
		t.Errorf("Expected attendee status %s , found %s  \n", "ACCEPTED", john.GetStatus())

	}
	if sue.GetStatus() != "NEEDS-ACTION" {
		t.Errorf("Expected attendee status %s , found %s  \n", "NEEDS-ACTION", sue.GetStatus())

	}
	if travis.GetStatus() != "NEEDS-ACTION" {
		t.Errorf("Expected attendee status %s , found %s  \n", "NEEDS-ACTION", travis.GetStatus())

	}

	// check role
	if john.GetRole() != "REQ-PARTICIPANT" {
		t.Errorf("Expected attendee status %s , found %s  \n", "REQ-PARTICIPANT", john.GetRole())

	}
	if sue.GetRole() != "REQ-PARTICIPANT" {
		t.Errorf("Expected attendee status %s , found %s  \n", "REQ-PARTICIPANT", sue.GetRole())

	}
	if travis.GetRole() != "REQ-PARTICIPANT" {
		t.Errorf("Expected attendee status %s , found %s  \n", "REQ-PARTICIPANT", travis.GetRole())

	}

}
