package ics

import (
	"fmt"
	"io/ioutil"
	"regexp"
	// "strings"
	"errors"
	// "fmt"
	// "io"
	// "net/http"
	// "os"
	"strconv"
	"sync"
	"time"
)

func init() {
	Wg = new(sync.WaitGroup)
	mutex = new(sync.Mutex)
	idCounter = 0
}

type Parser struct {
	inputChan       chan string
	outputChan      chan Event
	errChan         chan error
	parsedCalendars []*Calendar
	statusCalendars int
}

// creates new parser
func New() *Parser {
	p := new(Parser)
	p.inputChan = make(chan string)
	p.outputChan = make(chan Event)
	p.errChan = make(chan error)
	p.parsedCalendars = []*Calendar{}

	go func(input chan string) {
		// fmt.Println("Goroute")
		for {
			link := <-input
			Wg.Add(1)
			p.statusCalendars++
			go func(link string) {
				iCalContent, err := p.getICal(link)
				if err != nil {
					p.errChan <- err
					return
				}
				p.parseICalContent(iCalContent)
				mutex.Lock()
				p.statusCalendars--
				mutex.Unlock()
			}(link)
		}
	}(p.inputChan)
	// p.wg.Wait()
	// return p.inputChan
	return p
}

//  returns the chan for calendar urls
func (p *Parser) GetInputChan() chan string {
	return p.inputChan
}

// returns the chan where will be received events
func (p *Parser) GetOutputChan() chan Event {
	return p.outputChan
}

// returns the chan where will be received events
func (p *Parser) GetCalendars() ([]*Calendar, error) {
	if !p.Done() {
		return nil, errors.New("Calendars not parsed")
	}
	return p.parsedCalendars, nil
}

// is everything is parsed
func (p *Parser) Done() bool {
	return p.statusCalendars == 0
}

//  get the data from the calendar
func (p *Parser) getICal(url string) (string, error) {
	fileName, errDownload := downloadFromUrl(url)

	if errDownload != nil {
		return "", errDownload
	}

	fileContent, errReadFile := ioutil.ReadFile(fileName)
	if errReadFile != nil {
		return "", errReadFile
	}
	return fmt.Sprintf("%s", fileContent), nil
}

// ======================== CALENDAR PARSING ===================

// parses the iCal formated string
func (p *Parser) parseICalContent(iCalContent string) {
	ical := NewCalendar()
	p.parsedCalendars = append(p.parsedCalendars, ical)

	// split the data into calendar info and events data
	eventsData, calInfo := explodeICal(iCalContent)
	idCounter++

	// fill the calendar fields
	ical.SetName(p.parseICalName(calInfo))
	ical.SetDesc(p.parseICalDesc(calInfo))
	ical.SetVersion(p.parseICalVersion(calInfo))
	ical.SetTimezone(p.parseICalTimezone(calInfo))

	// parse the events and add them to ical
	p.parseEvents(ical, eventsData)
	// fmt.Printf("%#v \n", ical)
	Wg.Done()
}

// explodes the ICal content to array of events and calendar info
func explodeICal(iCalContent string) ([]string, string) {
	reEvents, _ := regexp.Compile(`(BEGIN:VEVENT(.*\n)*?END:VEVENT\r\n)`)
	allEvents := reEvents.FindAllString(iCalContent, len(iCalContent))
	calInfo := reEvents.ReplaceAllString(iCalContent, "")
	return allEvents, calInfo
}

// parses the iCal Name
func (p *Parser) parseICalName(iCalContent string) string {
	re, _ := regexp.Compile(`X-WR-CALNAME:.*?\n`)
	result := re.FindString(iCalContent)
	return trimField(result, "X-WR-CALNAME:")
}

// parses the iCal description
func (p *Parser) parseICalDesc(iCalContent string) string {
	re, _ := regexp.Compile(`X-WR-CALDESC:.*?\n`)
	result := re.FindString(iCalContent)
	return trimField(result, "X-WR-CALDESC:")
}

// parses the iCal version
func (p *Parser) parseICalVersion(iCalContent string) float64 {
	re, _ := regexp.Compile(`VERSION:.*?\n`)
	result := re.FindString(iCalContent)
	// parse the version result to float
	ver, _ := strconv.ParseFloat(trimField(result, "VERSION:"), 64)
	return ver
}

// parses the iCal timezone
func (p *Parser) parseICalTimezone(iCalContent string) time.Location {
	re, _ := regexp.Compile(`X-WR-TIMEZONE:.*?\n`)
	result := re.FindString(iCalContent)

	// parse the timezone result to time.Location
	timezone := trimField(result, "X-WR-TIMEZONE:")
	fmt.Println(result)
	loc, err := time.LoadLocation(timezone)

	// if fails with the timezone => go Local
	if err != nil {
		fmt.Println(err)
		loc, _ = time.LoadLocation("UTC")
	}
	return *loc
}

// ======================== EVENTS PARSING ===================

// parses the iCal events Data
func (p *Parser) parseEvents(cal *Calendar, eventsData []string) {
	for _, eventData := range eventsData {
		event := NewEvent()

		start := p.parseEventStart(eventData)
		end := p.parseEventEnd(eventData)

		// whole day event when both times are 00:00:00
		wholeDay := start.Hour() == 0 && end.Hour() == 0 && start.Minute() == 0 && end.Minute() == 0 && start.Second() == 0 && end.Second() == 0

		event.SetStatus(p.parseEventStatus(eventData))
		event.SetSummary(p.parseEventSummary(eventData))
		event.SetDescription(p.parseEventDescription(eventData))
		event.SetImportedID(p.parseEventId(eventData))
		event.SetClass(p.parseEventClass(eventData))
		event.SetSequence(p.parseEventSequence(eventData))
		event.SetCreated(p.parseEventCreated(eventData))
		event.SetLastModified(p.parseEventModified(eventData))
		event.SetStart(start)
		event.SetEnd(end)
		event.SetWholeDayEvent(wholeDay)
		if !event.IsWholeDay() {
			fmt.Printf("%#v \n", event)
		}
		// break
	}
}

// parses the event summary
func (p *Parser) parseEventSummary(eventData string) string {
	re, _ := regexp.Compile(`SUMMARY:.*?\n`)
	result := re.FindString(eventData)
	return trimField(result, "SUMMARY:")
}

// parses the event status
func (p *Parser) parseEventStatus(eventData string) string {
	re, _ := regexp.Compile(`STATUS:.*?\n`)
	result := re.FindString(eventData)
	return trimField(result, "STATUS:")
}

// parses the event description
func (p *Parser) parseEventDescription(eventData string) string {
	re, _ := regexp.Compile(`DESCRIPTION:.*?\n`)
	result := re.FindString(eventData)
	return trimField(result, "DESCRIPTION:")
}

// parses the event id provided form google
func (p *Parser) parseEventId(eventData string) string {
	re, _ := regexp.Compile(`UID:.*?\n`)
	result := re.FindString(eventData)
	return trimField(result, "UID:")
}

// parses the event class
func (p *Parser) parseEventClass(eventData string) string {
	re, _ := regexp.Compile(`CLASS:.*?\n`)
	result := re.FindString(eventData)
	return trimField(result, "CLASS:")
}

// parses the event sequence
func (p *Parser) parseEventSequence(eventData string) int {
	re, _ := regexp.Compile(`SEQUENCE:.*?\n`)
	result := re.FindString(eventData)
	sq, _ := strconv.Atoi(trimField(result, "SEQUENCE:"))
	return sq
}

// parses the event created time
func (p *Parser) parseEventCreated(eventData string) time.Time {
	re, _ := regexp.Compile(`CREATED:.*?\n`)
	result := re.FindString(eventData)
	created := trimField(result, "CREATED:")
	t, _ := time.Parse(IcsFormat, created)
	return t
}

// parses the event modified time
func (p *Parser) parseEventModified(eventData string) time.Time {
	re, _ := regexp.Compile(`LAST-MODIFIED:.*?\n`)
	result := re.FindString(eventData)
	modified := trimField(result, "LAST-MODIFIED:")
	t, _ := time.Parse(IcsFormat, modified)
	return t
}

// parses the event modified time
func (p *Parser) parseEventStart(eventData string) time.Time {
	reWholeDay, _ := regexp.Compile(`DTSTART;VALUE=DATE:.*?\n`)
	re, _ := regexp.Compile(`DTSTART:.*?\n`)
	resultWholeDay := reWholeDay.FindString(eventData)
	var t time.Time

	if resultWholeDay != "" {
		// whole day event
		modified := trimField(resultWholeDay, "DTSTART;VALUE=DATE:")
		t, _ = time.Parse(IcsFormatWholeDay, modified)
	} else {
		// event that has start hour and minute
		result := re.FindString(eventData)
		modified := trimField(result, "DTSTART:")
		t, _ = time.Parse(IcsFormat, modified)
	}

	return t
}

// parses the event modified time
func (p *Parser) parseEventEnd(eventData string) time.Time {
	reWholeDay, _ := regexp.Compile(`DTEND;VALUE=DATE:.*?\n`)
	re, _ := regexp.Compile(`DTEND:.*?\n`)
	resultWholeDay := reWholeDay.FindString(eventData)
	var t time.Time

	if resultWholeDay != "" {
		// whole day event
		modified := trimField(resultWholeDay, "DTEND;VALUE=DATE:")
		t, _ = time.Parse(IcsFormatWholeDay, modified)
	} else {
		// event that has end hour and minute
		result := re.FindString(eventData)
		modified := trimField(result, "DTEND:")
		t, _ = time.Parse(IcsFormat, modified)
	}
	return t
}
