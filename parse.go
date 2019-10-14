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

	duration "github.com/channelmeter/iso8601duration"
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
	MainLoop:
		for {
			if len(p.parsedEvents) > 0 {
				select {
				case p.outputChan <- p.parsedEvents[0]:
					p.parsedEvents = p.parsedEvents[1:]
				case event, more := <-p.bufferedChan:
					if !more {
						close(p.outputChan)
						break MainLoop
					}
					p.parsedEvents = append(p.parsedEvents, event)
				}
			} else {
				event, more := <-p.bufferedChan
				if !more {
					close(p.outputChan)
					break MainLoop
				}
				p.parsedEvents = append(p.parsedEvents, event)
			}
		}
	}()

	go func(input chan string) {
		// endless loop for getting the ics urls
		for {
			link, more := <-input

			if !more {
				break
			}

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
				calendar, events, err := ParseICalContent(iCalContent, link)

				p.parsedCalendars = append(p.parsedCalendars, calendar)
				if err != nil {
					p.errorsOccured = append(p.errorsOccured, err)
				}

				for _, v := range events {
					p.bufferedChan <- v
				}

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
	cal, events, err := ParseICalContent(iCalContent, "")
	p.parsedCalendars = append(p.parsedCalendars, cal)
	for _, v := range events {
		p.bufferedChan <- v
	}
	if err != nil {
		p.errorsOccured = append(p.errorsOccured, err)
	}
}

//  returns the chan for calendar urls
func (p *Parser) GetInputChan() chan string {
	return p.inputChan
}

// returns the chan where will be received events
func (p *Parser) GetOutputChan() chan *Event {
	return p.outputChan
}

// returns Calendars that were parsed
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

// wait until everything is parsed and close channels / close routines
func (p *Parser) WaitAndClose() {
	p.Wait()
	close(p.inputChan)
	close(p.bufferedChan)
}

//  get the data from the calendar
func (p *Parser) getICal(url string) (string, error) {
	re := regexp.MustCompile(`http(s){0,1}:\/\/`)

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
func ParseICalContent(iCalContent, url string) (*Calendar, []*Event, error) {
	ical := NewCalendar()

	// split the data into calendar info and events data
	eventsData, calInfo := explodeICal(iCalContent)
	idCounter++

	// fill the calendar fields
	ical.SetName(parseICalName(calInfo))
	ical.SetDesc(parseICalDesc(calInfo))
	ical.SetVersion(parseICalVersion(calInfo))
	tz, err := parseICalTimezone(calInfo)
	ical.SetTimezone(tz)
	ical.SetUrl(url)

	// parse the events and add them to ical
	events := parseEvents(ical, eventsData)

	return ical, events, err
}

var reEvents = regexp.MustCompile(`(BEGIN:VEVENT(.*\n)*?END:VEVENT\r?\n)`)

// explodes the ICal content to array of events and calendar info
func explodeICal(iCalContent string) ([]string, string) {
	allEvents := reEvents.FindAllString(iCalContent, len(iCalContent))
	calInfo := reEvents.ReplaceAllString(iCalContent, "")
	return allEvents, calInfo
}

var reICalName = regexp.MustCompile(`X-WR-CALNAME:.*?\n`)

// parses the iCal Name
func parseICalName(iCalContent string) string {
	result := reICalName.FindString(iCalContent)
	return trimField(result, "X-WR-CALNAME:")
}

var reICalDesc = regexp.MustCompile(`X-WR-CALDESC:.*?\n`)

// parses the iCal description
func parseICalDesc(iCalContent string) string {
	result := reICalDesc.FindString(iCalContent)
	return trimField(result, "X-WR-CALDESC:")
}

var reICalVersion = regexp.MustCompile(`VERSION:.*?\n`)

// parses the iCal version
func parseICalVersion(iCalContent string) float64 {
	result := reICalVersion.FindString(iCalContent)
	// parse the version result to float
	ver, _ := strconv.ParseFloat(trimField(result, "VERSION:"), 64)
	return ver
}

var reICalTimezone = regexp.MustCompile(`X-WR-TIMEZONE:.*?\n`)

// parses the iCal timezone
func parseICalTimezone(iCalContent string) (time.Location, error) {
	result := reICalTimezone.FindString(iCalContent)

	// parse the timezone result to time.Location
	timezone := trimField(result, "X-WR-TIMEZONE:")
	// create location instance
	loc, err := time.LoadLocation(timezone)

	// if fails with the timezone => go Local
	if err != nil {
		loc, _ = time.LoadLocation("UTC")
	}
	return *loc, err
}

// ======================== EVENTS PARSING ===================

var (
	reUntil    = regexp.MustCompile(`UNTIL=(\d)*T(\d)*Z(;){0,1}`)
	csUntil    = `(UNTIL=|;)`
	reInterval = regexp.MustCompile(`INTERVAL=(\d)*(;){0,1}`)
	csInterval = `(INTERVAL=|;)`
	reCount    = regexp.MustCompile(`COUNT=(\d)*(;){0,1}`)
	csCount    = `(COUNT=|;)`
	reFr       = regexp.MustCompile(`FREQ=[^;]*(;){0,1}`)
	csFr       = `(FREQ=|;)`
	reBM       = regexp.MustCompile(`BYMONTH=[^;]*(;){0,1}`)
	csBM       = `(BYMONTH=|;)`
	reBD       = regexp.MustCompile(`BYDAY=[^;]*(;){0,1}`)
	csBD       = `(BYDAY=|;)`
)

// parses the iCal events Data
func parseEvents(cal *Calendar, eventsData []string) []*Event {
	events := make([]*Event, len(eventsData))
	for i, eventData := range eventsData {
		event := NewEvent()

		start, startTZID := parseEventStart(eventData)
		end, endTZID := parseEventEnd(eventData)
		eventDuration := parseEventDuration(eventData)

		if end.Before(start) {
			end = start.Add(eventDuration)
		}
		// whole day event when both times are 00:00:00
		wholeDay := start.Hour() == 0 && end.Hour() == 0 && start.Minute() == 0 && end.Minute() == 0 && start.Second() == 0 && end.Second() == 0

		event.SetStartTZID(startTZID)
		event.SetEndTZID(endTZID)
		event.SetStatus(parseEventStatus(eventData))
		event.SetSummary(parseEventSummary(eventData))
		event.SetDescription(parseEventDescription(eventData))
		event.SetImportedID(parseEventId(eventData))
		event.SetClass(parseEventClass(eventData))
		event.SetSequence(parseEventSequence(eventData))
		event.SetCreated(parseEventCreated(eventData))
		event.SetLastModified(parseEventModified(eventData))
		event.SetRRule(parseEventRRule(eventData))
		event.SetLocation(parseEventLocation(eventData))
		event.SetGeo(parseEventGeo(eventData))
		event.SetStart(start)
		event.SetEnd(end)
		event.SetWholeDayEvent(wholeDay)
		event.SetAttendees(parseEventAttendees(eventData))
		event.SetOrganizer(parseEventOrganizer(eventData))
		event.SetCalendar(cal)
		event.SetID(event.GenerateEventId())

		cal.addEvent(*event)
		events[i] = event

		if RepeatRuleApply && event.GetRRule() != "" {

			// until field
			untilString := trimField(reUntil.FindString(event.GetRRule()), csUntil)
			//  set until date
			var until *time.Time
			if untilString == "" {
				until = nil
			} else {
				untilV, _ := time.Parse(IcsFormat, untilString)
				until = &untilV
			}

			// INTERVAL field
			intervalString := trimField(reInterval.FindString(event.GetRRule()), csInterval)
			interval, _ := strconv.Atoi(intervalString)

			if interval == 0 {
				interval = 1
			}

			// count field
			countString := trimField(reCount.FindString(event.GetRRule()), csCount)
			count, _ := strconv.Atoi(countString)
			if count == 0 {
				count = MaxRepeats
			}

			// freq field
			freq := trimField(reFr.FindString(event.GetRRule()), csFr)

			// by month field
			bymonth := trimField(reBM.FindString(event.GetRRule()), csBM)

			// by day field
			byday := trimField(reBD.FindString(event.GetRRule()), csBD)

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

	return events
}

var reEventSummary = regexp.MustCompile(`SUMMARY(?:;LANGUAGE=[a-zA-Z\-]+)?.*?\n`)

// parses the event summary
func parseEventSummary(eventData string) string {
	result := reEventSummary.FindString(eventData)
	return trimField(result, `SUMMARY(?:;LANGUAGE=[a-zA-Z\-]+)?:`)
}

var reEventStatus = regexp.MustCompile(`STATUS:.*?\n`)

// parses the event status
func parseEventStatus(eventData string) string {
	result := reEventStatus.FindString(eventData)
	return trimField(result, "STATUS:")
}

var reEventDescription = regexp.MustCompile(`DESCRIPTION:.*?\n(?:\s+.*?\n)*`)

// parses the event description
func parseEventDescription(eventData string) string {
	result := reEventDescription.FindString(eventData)
	return trimField(strings.Replace(result, "\r\n ", "", -1), "DESCRIPTION:")
}

var reEventId = regexp.MustCompile(`UID:.*?\n`)

// parses the event id provided form google
func parseEventId(eventData string) string {
	result := reEventId.FindString(eventData)
	return trimField(result, "UID:")
}

var reEeventClass = regexp.MustCompile(`CLASS:.*?\n`)

// parses the event class
func parseEventClass(eventData string) string {
	result := reEeventClass.FindString(eventData)
	return trimField(result, "CLASS:")
}

var reEventSequence = regexp.MustCompile(`SEQUENCE:.*?\n`)

// parses the event sequence
func parseEventSequence(eventData string) int {
	result := reEventSequence.FindString(eventData)
	sq, _ := strconv.Atoi(trimField(result, "SEQUENCE:"))
	return sq
}

var reEventCreated = regexp.MustCompile(`CREATED:.*?\n`)

// parses the event created time
func parseEventCreated(eventData string) time.Time {
	result := reEventCreated.FindString(eventData)
	created := trimField(result, "CREATED:")
	t, _ := time.Parse(IcsFormat, created)
	return t
}

var reEventModified = regexp.MustCompile(`LAST-MODIFIED:.*?\n`)

// parses the event modified time
func parseEventModified(eventData string) time.Time {
	result := reEventModified.FindString(eventData)
	modified := trimField(result, "LAST-MODIFIED:")
	t, _ := time.Parse(IcsFormat, modified)
	return t
}

// parses the event start time
func parseTimeField(fieldName string, eventData string) (time.Time, string) {
	reWholeDay := regexp.MustCompile(fmt.Sprintf(`%s;VALUE=DATE:.*?\n`, fieldName))
	resultWholeDay := reWholeDay.FindString(eventData)
	var t time.Time
	var tzID string

	if resultWholeDay != "" {
		// whole day event
		modified := trimField(resultWholeDay, fmt.Sprintf("%s;VALUE=DATE:", fieldName))
		t, _ = time.Parse(IcsFormatWholeDay, modified)
	} else {
		// event that has start hour and minute
		re := regexp.MustCompile(fmt.Sprintf(`%s(;TZID=(.*?))?(;VALUE=DATE-TIME)?:(.*?)\n`, fieldName))
		result := re.FindStringSubmatch(eventData)
		if result == nil || len(result) < 4 {
			return t, tzID
		}
		tzID = result[2]
		dt := result[4]
		if !strings.Contains(dt, "Z") {
			dt = fmt.Sprintf("%sZ", dt)
		}
		t, _ = time.Parse(IcsFormat, dt)
	}

	return t, tzID
}

// parses the event start time
func parseEventStart(eventData string) (time.Time, string) {
	return parseTimeField("DTSTART", eventData)
}

// parses the event end time
func parseEventEnd(eventData string) (time.Time, string) {
	return parseTimeField("DTEND", eventData)
}

var reEventDuration = regexp.MustCompile(`DURATION:.*?\n`)

func parseEventDuration(eventData string) time.Duration {
	result := reEventDuration.FindString(eventData)
	trimmed := trimField(result, "DURATION:")
	parsedDuration, err := duration.FromString(trimmed)
	var output time.Duration

	if err == nil {
		output = parsedDuration.ToDuration()
	}

	return output
}

var reEventRRule = regexp.MustCompile(`RRULE:.*?\n`)

// parses the event RRULE (the repeater)
func parseEventRRule(eventData string) string {
	result := reEventRRule.FindString(eventData)
	return trimField(result, "RRULE:")
}

var reEventLocation = regexp.MustCompile(`LOCATION:.*?\n`)

// parses the event LOCATION
func parseEventLocation(eventData string) string {
	result := reEventLocation.FindString(eventData)
	return trimField(result, "LOCATION:")
}

var reEventGeo = regexp.MustCompile(`GEO:.*?\n`)

// parses the event GEO
func parseEventGeo(eventData string) *Geo {
	result := reEventGeo.FindString(eventData)

	value := trimField(result, "GEO:")
	values := strings.Split(value, ";")
	if len(values) < 2 {
		return nil
	}

	return NewGeo(values[0], values[1])
}

// ======================== ATTENDEE PARSING ===================

var reEventAttendees = regexp.MustCompile(`ATTENDEE(:|;)(.*?\r?\n)(\s.*?\r?\n)*`)

// parses the event attendees
func parseEventAttendees(eventData string) []*Attendee {
	attendeesObj := []*Attendee{}
	attendees := reEventAttendees.FindAllString(eventData, len(eventData))

	for _, attendeeData := range attendees {
		if attendeeData == "" {
			continue
		}
		attendee := parseAttendee(strings.Replace(strings.Replace(attendeeData, "\r", "", 1), "\n ", "", 1))
		//  check for any fields set
		if attendee.GetEmail() != "" || attendee.GetName() != "" || attendee.GetRole() != "" || attendee.GetStatus() != "" || attendee.GetType() != "" {
			attendeesObj = append(attendeesObj, attendee)
		}
	}
	return attendeesObj
}

var reEventOrganizer = regexp.MustCompile(`ORGANIZER(:|;)(.*?\r?\n)(\s.*?\r?\n)*`)

// parses the event organizer
func parseEventOrganizer(eventData string) *Attendee {
	organizerData := reEventOrganizer.FindString(eventData)
	if organizerData == "" {
		return nil
	}
	organizerDataFormated := strings.Replace(strings.Replace(organizerData, "\r", "", 1), "\n ", "", 1)

	a := NewAttendee()
	a.SetEmail(parseAttendeeMail(organizerDataFormated))
	a.SetName(parseOrganizerName(organizerDataFormated))

	return a
}

//  parse attendee properties
func parseAttendee(attendeeData string) *Attendee {
	a := NewAttendee()
	a.SetEmail(parseAttendeeMail(attendeeData))
	a.SetName(parseAttendeeName(attendeeData))
	a.SetRole(parseAttendeeRole(attendeeData))
	a.SetStatus(parseAttendeeStatus(attendeeData))
	a.SetType(parseAttendeeType(attendeeData))
	return a
}

var reAttendeeMail = regexp.MustCompile(`mailto:.*?\n`)

// parses the attendee email
func parseAttendeeMail(attendeeData string) string {
	result := reAttendeeMail.FindString(attendeeData)
	return trimField(result, "mailto:")
}

var reAttendeeStatus = regexp.MustCompile(`PARTSTAT=.*?;`)

// parses the attendee status
func parseAttendeeStatus(attendeeData string) string {
	result := reAttendeeStatus.FindString(attendeeData)
	if result == "" {
		return ""
	}
	return trimField(result, `(PARTSTAT=|;)`)
}

var reAttendeeRole = regexp.MustCompile(`ROLE=.*?;`)

// parses the attendee role
func parseAttendeeRole(attendeeData string) string {
	result := reAttendeeRole.FindString(attendeeData)

	if result == "" {
		return ""
	}
	return trimField(result, `(ROLE=|;)`)
}

var reAttendeeName = regexp.MustCompile(`CN=.*?;`)

// parses the attendee Name
func parseAttendeeName(attendeeData string) string {
	result := reAttendeeName.FindString(attendeeData)
	if result == "" {
		return ""
	}
	return trimField(result, `(CN=|;)`)
}

var reOrganizerName = regexp.MustCompile(`CN=.*?:`)

// parses the organizer Name
func parseOrganizerName(orgData string) string {
	result := reOrganizerName.FindString(orgData)
	if result == "" {
		return ""
	}
	return trimField(result, `(CN=|:)`)
}

var reAttendeeType = regexp.MustCompile(`CUTYPE=.*?;`)

// parses the attendee type
func parseAttendeeType(attendeeData string) string {
	result := reAttendeeType.FindString(attendeeData)
	if result == "" {
		return ""
	}
	return trimField(result, `(CUTYPE=|;)`)
}
