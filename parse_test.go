package ics_test

import (
	"fmt"
	"github.com/PuloV/ics-golang"
	"reflect"
	"testing"
)

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
