package ics

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

func init() {
	mutex = new(sync.Mutex)
	DeleteTempFiles = true
	FilePath = "tmp/"
	RepeatRuleApply = false
	MaxRepeats = 10
}

type Parser struct {
	inputChan       chan string
	outputChan      chan *Event
	bufferedChan    chan *Event
	errorsOccured   []error
	parsedCalendars []*Calendar
	parsedEvents    []*Event
	statusCalendars int
	wg              *sync.WaitGroup
}

// creates new parser
func New() *Parser {
	p := new(Parser)
	p.inputChan = make(chan string)
	p.outputChan = make(chan *Event)
	p.bufferedChan = make(chan *Event)
	p.errorsOccured = []error{}
	p.wg = new(sync.WaitGroup)
	p.parsedCalendars = []*Calendar{}
	p.parsedEvents = []*Event{}

	// buffers the events output chan
	go func() {
		for {
			if len(p.parsedEvents) > 0 {
				select {
				case p.outputChan <- p.parsedEvents[0]:
					p.parsedEvents = p.parsedEvents[1:]
				case event := <-p.bufferedChan:
					p.parsedEvents = append(p.parsedEvents, event)
				}
			} else {
				event := <-p.bufferedChan
				p.parsedEvents = append(p.parsedEvents, event)
			}
		}
	}()

	go func(input chan string) {
		// endless loop for getting the ics urls
		for {
			link := <-input

			// mark calendar in the wait group as not parsed
			p.wg.Add(1)

			// marks that we have statusCalendars +1 calendars to be parsed
			mutex.Lock()
			p.statusCalendars++
			mutex.Unlock()

			go func(link string) {
				// mark calendar in the wait group as  parsed
				defer p.wg.Done()

				iCalContent, err := p.getICal(link)
				if err != nil {
					p.errorsOccured = append(p.errorsOccured, err)

					mutex.Lock()
					// marks that we have parsed 1 calendar and we have statusCalendars -1 left to be parsed
					p.statusCalendars--
					mutex.Unlock()
					return
				}

				// parse the ICal calendar
				p.parseICalContent(iCalContent, link)

				mutex.Lock()
				// marks that we have parsed 1 calendar and we have statusCalendars -1 left to be parsed
				p.statusCalendars--
				mutex.Unlock()

			}(link)
		}
	}(p.inputChan)
	// p.wg.Wait()
	// return p.inputChan
	return p
}

// Load calender from content
func (p *Parser) Load(iCalContent string) {
	p.parseICalContent(iCalContent, "")
}

//  returns the chan for calendar urls
func (p *Parser) GetInputChan() chan string {
	return p.inputChan
}

// returns the chan where will be received events
func (p *Parser) GetOutputChan() chan *Event {
	return p.outputChan
}

// returns the chan where will be received events
func (p *Parser) GetCalendars() ([]*Calendar, error) {
	if !p.Done() {
		return nil, errors.New("Calendars not parsed")
	}
	return p.parsedCalendars, nil
}

// returns the array with the errors occurred while parsing the events
func (p *Parser) GetErrors() ([]error, error) {
	if !p.Done() {
		return nil, errors.New("Calendars not parsed")
	}
	return p.errorsOccured, nil
}

// is everything is parsed
func (p *Parser) Done() bool {
	return p.statusCalendars == 0
}

// wait until everything is parsed
func (p *Parser) Wait() {
	p.wg.Wait()
}

//  get the data from the calendar
func (p *Parser) getICal(url string) (string, error) {
	re, _ := regexp.Compile(`http(s){0,1}:\/\/`)

	var fileName string
	var errDownload error

	if re.FindString(url) != "" {
		// download the file and store it local
		fileName, errDownload = downloadFromUrl(url)

		if errDownload != nil {
			return "", errDownload
		}

	} else { //  use a file from local storage

		//  check if file exists
		if fileExists(url) {
			fileName = url
		} else {
			err := fmt.Sprintf("File %s does not exists", url)
			return "", errors.New(err)
		}
	}

	//  read the file with the ical data
	fileContent, errReadFile := ioutil.ReadFile(fileName)

	if errReadFile != nil {
		return "", errReadFile
	}

	if DeleteTempFiles && re.FindString(url) != "" {
		os.Remove(fileName)
	}

	return fmt.Sprintf("%s", fileContent), nil
}

// ======================== CALENDAR PARSING ===================

// parses the iCal formated string to a calendar object
func (p *Parser) parseICalContent(iCalContent, url string) {
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
	ical.SetUrl(url)

	// parse the events and add them to ical
	p.parseEvents(ical, eventsData)

}

// explodes the ICal content to array of events and calendar info
func explodeICal(iCalContent string) ([]string, string) {
	reEvents, _ := regexp.Compile(`(BEGIN:VEVENT(.*\n)*?END:VEVENT\r?\n)`)
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
	// create location instance
	loc, err := time.LoadLocation(timezone)

	// if fails with the timezone => go Local
	if err != nil {
		p.errorsOccured = append(p.errorsOccured, err)
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
		event.SetRRule(p.parseEventRRule(eventData))
		event.SetLocation(p.parseEventLocation(eventData))
		event.SetGeo(p.parseEventGeo(eventData))
		event.SetStart(start)
		event.SetEnd(end)
		event.SetWholeDayEvent(wholeDay)
		event.SetAttendees(p.parseEventAttendees(eventData))
		event.SetOrganizer(p.parseEventOrganizer(eventData))
		event.SetCalendar(cal)
		event.SetID(event.GenerateEventId())

		cal.SetEvent(*event)
		p.bufferedChan <- event

		if RepeatRuleApply && event.GetRRule() != "" {

			// until field
			reUntil, _ := regexp.Compile(`UNTIL=(\d)*T(\d)*Z(;){0,1}`)
			untilString := trimField(reUntil.FindString(event.GetRRule()), `(UNTIL=|;)`)
			//  set until date
			var until *time.Time
			if untilString == "" {
				until = nil
			} else {
				untilV, _ := time.Parse(IcsFormat, untilString)
				until = &untilV
			}

			// INTERVAL field
			reInterval, _ := regexp.Compile(`INTERVAL=(\d)*(;){0,1}`)
			intervalString := trimField(reInterval.FindString(event.GetRRule()), `(INTERVAL=|;)`)
			interval, _ := strconv.Atoi(intervalString)

			if interval == 0 {
				interval = 1
			}

			// count field
			reCount, _ := regexp.Compile(`COUNT=(\d)*(;){0,1}`)
			countString := trimField(reCount.FindString(event.GetRRule()), `(COUNT=|;)`)
			count, _ := strconv.Atoi(countString)
			if count == 0 {
				count = MaxRepeats
			}

			// freq field
			reFr, _ := regexp.Compile(`FREQ=[^;]*(;){0,1}`)
			freq := trimField(reFr.FindString(event.GetRRule()), `(FREQ=|;)`)

			// by month field
			reBM, _ := regexp.Compile(`BYMONTH=[^;]*(;){0,1}`)
			bymonth := trimField(reBM.FindString(event.GetRRule()), `(BYMONTH=|;)`)

			// by day field
			reBD, _ := regexp.Compile(`BYDAY=[^;]*(;){0,1}`)
			byday := trimField(reBD.FindString(event.GetRRule()), `(BYDAY=|;)`)

			// fmt.Printf("%#v \n", reBD.FindString(event.GetRRule()))
			// fmt.Println("untilString", reUntil.FindString(event.GetRRule()))

			//  set the freq modification of the dates
			var years, days, months int
			switch freq {
			case "DAILY":
				days = interval
				months = 0
				years = 0
				break
			case "WEEKLY":
				days = 7
				months = 0
				years = 0
				break
			case "MONTHLY":
				days = 0
				months = interval
				years = 0
				break
			case "YEARLY":
				days = 0
				months = 0
				years = interval
				break
			}

			// number of current repeats
			current := 0
			// the current date in the main loop
			freqDateStart := start
			freqDateEnd := end

			// loops by freq
			for {
				weekDaysStart := freqDateStart
				weekDaysEnd := freqDateEnd

				// check repeating by month
				if bymonth == "" || strings.Contains(bymonth, weekDaysStart.Format("1")) {

					if byday != "" {
						// loops the weekdays
						for i := 0; i < 7; i++ {
							day := parseDayNameToIcsName(weekDaysStart.Format("Mon"))
							if strings.Contains(byday, day) && weekDaysStart != start {
								current++
								count--
								newE := *event
								newE.SetStart(weekDaysStart)
								newE.SetEnd(weekDaysEnd)
								newE.SetID(newE.GenerateEventId())
								newE.SetSequence(current)
								if until == nil || (until != nil && until.Format(YmdHis) >= weekDaysStart.Format(YmdHis)) {
									cal.SetEvent(newE)
								}

							}
							weekDaysStart = weekDaysStart.AddDate(0, 0, 1)
							weekDaysEnd = weekDaysEnd.AddDate(0, 0, 1)
						}
					} else {
						//  we dont have loop by day so we put it on the same day
						if weekDaysStart != start {
							current++
							count--
							newE := *event
							newE.SetStart(weekDaysStart)
							newE.SetEnd(weekDaysEnd)
							newE.SetID(newE.GenerateEventId())
							newE.SetSequence(current)
							if until == nil || (until != nil && until.Format(YmdHis) >= weekDaysStart.Format(YmdHis)) {
								cal.SetEvent(newE)
							}

						}
					}

				}

				freqDateStart = freqDateStart.AddDate(years, months, days)
				freqDateEnd = freqDateEnd.AddDate(years, months, days)
				if current > MaxRepeats || count == 0 {
					break
				}

				if until != nil && until.Format(YmdHis) <= freqDateStart.Format(YmdHis) {
					break
				}
			}

		}
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
	re, _ := regexp.Compile(`DESCRIPTION:.*?\n(?:\s+.*?\n)*`)
	result := re.FindString(eventData)
	return trimField(strings.Replace(result, "\r\n ", "", -1), "DESCRIPTION:")
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

// parses the event start time
func (p *Parser) parseEventStart(eventData string) time.Time {
	reWholeDay, _ := regexp.Compile(`DTSTART;VALUE=DATE:.*?\n`)
	re, _ := regexp.Compile(`DTSTART(;TZID=.*?){0,1}:.*?\n`)
	resultWholeDay := reWholeDay.FindString(eventData)
	var t time.Time

	if resultWholeDay != "" {
		// whole day event
		modified := trimField(resultWholeDay, "DTSTART;VALUE=DATE:")
		t, _ = time.Parse(IcsFormatWholeDay, modified)
	} else {
		// event that has start hour and minute
		result := re.FindString(eventData)
		modified := trimField(result, "DTSTART(;TZID=.*?){0,1}:")

		if !strings.Contains(modified, "Z") {
			modified = fmt.Sprintf("%sZ", modified)
		}

		t, _ = time.Parse(IcsFormat, modified)
	}

	return t
}

// parses the event end time
func (p *Parser) parseEventEnd(eventData string) time.Time {
	reWholeDay, _ := regexp.Compile(`DTEND;VALUE=DATE:.*?\n`)
	re, _ := regexp.Compile(`DTEND(;TZID=.*?){0,1}:.*?\n`)
	resultWholeDay := reWholeDay.FindString(eventData)
	var t time.Time

	if resultWholeDay != "" {
		// whole day event
		modified := trimField(resultWholeDay, "DTEND;VALUE=DATE:")
		t, _ = time.Parse(IcsFormatWholeDay, modified)
	} else {
		// event that has end hour and minute
		result := re.FindString(eventData)
		modified := trimField(result, "DTEND(;TZID=.*?){0,1}:")

		if !strings.Contains(modified, "Z") {
			modified = fmt.Sprintf("%sZ", modified)
		}
		t, _ = time.Parse(IcsFormat, modified)
	}
	return t

}

// parses the event RRULE (the repeater)
func (p *Parser) parseEventRRule(eventData string) string {
	re, _ := regexp.Compile(`RRULE:.*?\n`)
	result := re.FindString(eventData)
	return trimField(result, "RRULE:")
}

// parses the event LOCATION
func (p *Parser) parseEventLocation(eventData string) string {
	re, _ := regexp.Compile(`LOCATION:.*?\n`)
	result := re.FindString(eventData)
	return trimField(result, "LOCATION:")
}

// parses the event GEO
func (p *Parser) parseEventGeo(eventData string) *Geo {
	re, _ := regexp.Compile(`GEO:.*?\n`)
	result := re.FindString(eventData)

	value := trimField(result, "GEO:")
	values := strings.Split(value, ";")
	if len(values) < 2 {
		return nil
	}

	return NewGeo(values[0], values[1])
}

// ======================== ATTENDEE PARSING ===================

// parses the event attendees
func (p *Parser) parseEventAttendees(eventData string) []*Attendee {
	attendeesObj := []*Attendee{}
	re, _ := regexp.Compile(`ATTENDEE(:|;)(.*?\r?\n)(\s.*?\r?\n)*`)
	attendees := re.FindAllString(eventData, len(eventData))

	for _, attendeeData := range attendees {
		if attendeeData == "" {
			continue
		}
		attendee := p.parseAttendee(strings.Replace(strings.Replace(attendeeData, "\r", "", 1), "\n ", "", 1))
		//  check for any fields set
		if attendee.GetEmail() != "" || attendee.GetName() != "" || attendee.GetRole() != "" || attendee.GetStatus() != "" || attendee.GetType() != "" {
			attendeesObj = append(attendeesObj, attendee)
		}
	}
	return attendeesObj
}

// parses the event organizer
func (p *Parser) parseEventOrganizer(eventData string) *Attendee {

	re, _ := regexp.Compile(`ORGANIZER(:|;)(.*?\r?\n)(\s.*?\r?\n)*`)
	organizerData := re.FindString(eventData)
	if organizerData == "" {
		return nil
	}
	organizerDataFormated := strings.Replace(strings.Replace(organizerData, "\r", "", 1), "\n ", "", 1)

	a := NewAttendee()
	a.SetEmail(p.parseAttendeeMail(organizerDataFormated))
	a.SetName(p.parseOrganizerName(organizerDataFormated))

	return a
}

//  parse attendee properties
func (p *Parser) parseAttendee(attendeeData string) *Attendee {

	a := NewAttendee()
	a.SetEmail(p.parseAttendeeMail(attendeeData))
	a.SetName(p.parseAttendeeName(attendeeData))
	a.SetRole(p.parseAttendeeRole(attendeeData))
	a.SetStatus(p.parseAttendeeStatus(attendeeData))
	a.SetType(p.parseAttendeeType(attendeeData))
	return a
}

// parses the attendee email
func (p *Parser) parseAttendeeMail(attendeeData string) string {
	re, _ := regexp.Compile(`mailto:.*?\n`)
	result := re.FindString(attendeeData)
	return trimField(result, "mailto:")
}

// parses the attendee status
func (p *Parser) parseAttendeeStatus(attendeeData string) string {
	re, _ := regexp.Compile(`PARTSTAT=.*?;`)
	result := re.FindString(attendeeData)
	if result == "" {
		return ""
	}
	return trimField(result, `(PARTSTAT=|;)`)
}

// parses the attendee role
func (p *Parser) parseAttendeeRole(attendeeData string) string {
	re, _ := regexp.Compile(`ROLE=.*?;`)
	result := re.FindString(attendeeData)

	if result == "" {
		return ""
	}
	return trimField(result, `(ROLE=|;)`)
}

// parses the attendee Name
func (p *Parser) parseAttendeeName(attendeeData string) string {
	re, _ := regexp.Compile(`CN=.*?;`)
	result := re.FindString(attendeeData)
	if result == "" {
		return ""
	}
	return trimField(result, `(CN=|;)`)
}

// parses the organizer Name
func (p *Parser) parseOrganizerName(orgData string) string {
	re, _ := regexp.Compile(`CN=.*?:`)
	result := re.FindString(orgData)
	if result == "" {
		return ""
	}
	return trimField(result, `(CN=|:)`)
}

// parses the attendee type
func (p *Parser) parseAttendeeType(attendeeData string) string {
	re, _ := regexp.Compile(`CUTYPE=.*?;`)
	result := re.FindString(attendeeData)
	if result == "" {
		return ""
	}
	return trimField(result, `(CUTYPE=|;)`)
}
