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

var (
	repeatRuleApply = false
	maxRepeats      = 10

	re = struct {
		url, version, events, calName, calDesc, timezone, until, interval, count, freq, byMonth, byDay,
		summary, status, desc, uid, class, sequence, created, lastModified,
		tzid, dtstart, dtend,
		dstartDay, dstart, dendDay, dend,
		rrule, location, geo, attendee, organizer, mail, partStat, role, attName, orgName, cuType *regexp.Regexp
	}{}
)

type Parser struct {
	errorsOccured   []error
	parsedCalendars []*Calendar
	parsedEvents    []*Event
	statusCalendars int
	wg              sync.WaitGroup
	mutex           sync.Mutex
}

func init() {
	re.url = regexp.MustCompile(`http(s){0,1}:\/\/`)
	re.version = regexp.MustCompile(`VERSION:(.*?)\r?\n`)
	re.events = regexp.MustCompile(`(BEGIN:VEVENT(.*\n)*?END:VEVENT\r?\n)`)
	re.calName = regexp.MustCompile(`X-WR-CALNAME:(.*?)\r?\n`)
	re.calDesc = regexp.MustCompile(`X-WR-CALDESC:(.*?)\r?\n`)
	re.timezone = regexp.MustCompile(`X-WR-TIMEZONE:(.*?)\r?\n`)
	re.until = regexp.MustCompile(`UNTIL=(\d)*T(\d)*Z(;){0,1}`)
	re.interval = regexp.MustCompile(`INTERVAL=(\d)*(;){0,1}`)
	re.count = regexp.MustCompile(`COUNT=((\d)*)(;){0,1}`)
	re.freq = regexp.MustCompile(`FREQ=([^;]*)(;){0,1}`)
	re.byMonth = regexp.MustCompile(`BYMONTH=([^;]*)(;){0,1}`)
	re.byDay = regexp.MustCompile(`BYDAY=[^;]*(;){0,1}`)
	re.summary = regexp.MustCompile(`SUMMARY:(.*?)\r?\n`)
	re.status = regexp.MustCompile(`STATUS:(.*?)\r?\n`)
	re.desc = regexp.MustCompile(`DESCRIPTION:(.*?)\n(?:\s+.*?\n)*`)
	re.uid = regexp.MustCompile(`UID:(.*?)\r?\n`)
	re.class = regexp.MustCompile(`CLASS:(.*?)\r?\n`)
	re.sequence = regexp.MustCompile(`SEQUENCE:(.*?)\r?\n`)
	re.created = regexp.MustCompile(`CREATED:(.*?)\r?\n`)
	re.lastModified = regexp.MustCompile(`LAST-MODIFIED:(.*?)\r?\n`)
	re.tzid = regexp.MustCompile(`TZID=(.*)`)
	re.dtstart = regexp.MustCompile(`DTSTART;{0,1}(.*?)\r?\n`)
	re.dtend = regexp.MustCompile(`DTEND;{0,1}(.*?)\r?\n`)
	re.rrule = regexp.MustCompile(`RRULE:(.*?)\r?\n`)
	re.location = regexp.MustCompile(`LOCATION:(.*?)\r?\n`)
	re.geo = regexp.MustCompile(`GEO:(.*?)\r?\n`)
	re.attendee = regexp.MustCompile(`ATTENDEE(:|;)(.*?\r?\n)(\s.*?\r?\n)*`)
	re.organizer = regexp.MustCompile(`ORGANIZER(:|;)(.*?\r?\n)(\s.*?\r?\n)*`)
	re.mail = regexp.MustCompile(`mailto:(.*?)\r?\n`)
	re.partStat = regexp.MustCompile(`PARTSTAT=(.*?);`)
	re.role = regexp.MustCompile(`ROLE=(.*?);`)
	re.attName = regexp.MustCompile(`CN=(.*?);`)
	re.orgName = regexp.MustCompile(`CN=(.*?):`)
	re.cuType = regexp.MustCompile(`CUTYPE=(.*?);`)
}

// creates new parser
func New() *Parser {
	p := new(Parser)
	p.errorsOccured = []error{}
	p.parsedCalendars = []*Calendar{}
	p.parsedEvents = []*Event{}

	return p
}

func (p *Parser) LoadAsyncFromUrl(link string) {
	// mark calendar in the wait group as not parsed
	p.wg.Add(1)
	go func() {
		// mark calendar in the wait group as  parsed
		defer p.wg.Done()
		// marks that we have statusCalendars +1 calendars to be parsed
		p.atomicStatusCalendars(1)
		defer p.atomicStatusCalendars(-1)

		iCalContent, err := p.getICal(link)
		if err != nil {
			p.atomicAddError(err)
			return
		}

		// parse the ICal calendar
		p.parseICalContent(iCalContent, link)
	}()
}

// Load calender from content
func (p *Parser) Load(iCalContent string) {
	p.parseICalContent(iCalContent, "")
}

// returns the chan where will be received events
func (p *Parser) GetCalendars() ([]*Calendar, error) {
	if !p.Done() {
		return nil, errors.New("Calendars not parsed")
	}
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.parsedCalendars, nil
}

// returns the array with the errors occurred while parsing the events
func (p *Parser) GetErrors() ([]error, error) {
	if !p.Done() {
		return nil, errors.New("Calendars not parsed")
	}
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.errorsOccured, nil
}

// is everything is parsed
func (p *Parser) Done() bool {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.statusCalendars == 0
}

// wait until everything is parsed
func (p *Parser) Wait() {
	p.wg.Wait()
}

//  get the data from the calendar
func (p *Parser) getICal(url string) (string, error) {
	var fileName string
	var errDownload error

	urlFound := re.url.MatchString(url)
	if urlFound {
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

	if urlFound {
		os.Remove(fileName)
	}

	return fmt.Sprintf("%s", fileContent), nil
}

// ======================== CALENDAR PARSING ===================

// parses the iCal formated string to a calendar object
func (p *Parser) parseICalContent(iCalContent, url string) {
	ical := NewCalendar()

	// split the data into calendar info and events data
	eventsData, calInfo := explodeICal(iCalContent)

	// fill the calendar fields
	ical.SetName(p.parseICalName(calInfo))
	ical.SetDesc(p.parseICalDesc(calInfo))
	ical.SetVersion(p.parseICalVersion(calInfo))
	ical.SetTimezone(p.parseICalTimezone(calInfo))
	ical.SetUrl(url)

	// parse the events and add them to ical
	p.parseEvents(ical, eventsData)

	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.parsedCalendars = append(p.parsedCalendars, ical)
}

// explodes the ICal content to array of events and calendar info
func explodeICal(iCalContent string) ([]string, string) {
	allEvents := re.events.FindAllString(iCalContent, -1)
	calInfo := re.events.ReplaceAllString(iCalContent, "")
	return allEvents, calInfo
}

// parses the iCal Name
func (p *Parser) parseICalName(iCalContent string) string {
	return p.extractData(re.calName, iCalContent)
}

// parses the iCal description
func (p *Parser) parseICalDesc(iCalContent string) string {
	return p.extractData(re.calDesc, iCalContent)
}

// parses the iCal version
func (p *Parser) parseICalVersion(iCalContent string) float64 {
	// parse the version result to float
	ver, _ := strconv.ParseFloat(p.extractData(re.version, iCalContent), 64)
	return ver
}

// parses the iCal timezone
func (p *Parser) parseICalTimezone(iCalContent string) *time.Location {
	// parse the timezone result to time.Location
	timezone := p.extractData(re.timezone, iCalContent)
	// create location instance
	loc, err := time.LoadLocation(timezone)

	// if fails with the timezone => go with UTC
	if err != nil {
		p.errorsOccured = append(p.errorsOccured, err)
		loc, _ = time.LoadLocation("UTC")
	}
	return loc
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

		if repeatRuleApply && event.GetRRule() != "" {

			// until field
			untilString := p.extractData(re.until, event.GetRRule())
			//  set until date
			var until *time.Time
			if untilString == "" {
				until = nil
			} else {
				untilV, _ := time.Parse(icsFormat, untilString)
				until = &untilV
			}

			// INTERVAL field
			intervalString := p.extractData(re.interval, event.GetRRule())
			interval, _ := strconv.Atoi(intervalString)

			if interval == 0 {
				interval = 1
			}

			// count field
			countString := p.extractData(re.count, event.GetRRule())
			count, _ := strconv.Atoi(countString)
			if count == 0 {
				count = maxRepeats
			}

			// freq field
			freq := p.extractData(re.freq, event.GetRRule())

			// by month field
			bymonth := p.extractData(re.byMonth, event.GetRRule())

			// by day field
			byday := p.extractData(re.byDay, event.GetRRule())

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
								if until == nil || (until != nil && until.Format(ymdHis) >= weekDaysStart.Format(ymdHis)) {
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
							if until == nil || (until != nil && until.Format(ymdHis) >= weekDaysStart.Format(ymdHis)) {
								cal.SetEvent(newE)
							}

						}
					}

				}

				freqDateStart = freqDateStart.AddDate(years, months, days)
				freqDateEnd = freqDateEnd.AddDate(years, months, days)
				if current > maxRepeats || count == 0 {
					break
				}

				if until != nil && until.Format(ymdHis) <= freqDateStart.Format(ymdHis) {
					break
				}
			}

		}
	}
}

// parses the event summary
func (p *Parser) parseEventSummary(eventData string) string {
	return p.extractData(re.summary, eventData)
}

// parses the event status
func (p *Parser) parseEventStatus(eventData string) string {
	return p.extractData(re.status, eventData)
}

// parses the event description
func (p *Parser) parseEventDescription(eventData string) string {
	return strings.Replace(p.extractData(re.desc, eventData), "\r\n ", "", -1)
}

// parses the event id provided form google
func (p *Parser) parseEventId(eventData string) string {
	return p.extractData(re.uid, eventData)
}

// parses the event class
func (p *Parser) parseEventClass(eventData string) string {
	return p.extractData(re.class, eventData)
}

// parses the event sequence
func (p *Parser) parseEventSequence(eventData string) int {
	sq, _ := strconv.Atoi(p.extractData(re.sequence, eventData))
	return sq
}

// parses the event created time
func (p *Parser) parseEventCreated(eventData string) time.Time {
	created := p.extractData(re.created, eventData)
	t, _ := time.Parse(icsFormat, created)
	return t
}

// parses the event modified time
func (p *Parser) parseEventModified(eventData string) time.Time {
	modified := p.extractData(re.lastModified, eventData)
	t, _ := time.Parse(icsFormat, modified)
	return t
}

// parses multiple versions of event time (DTSTART/DTEND)
// possible options for VALUE:
// not present = implicit datetime
// DATE-TIME = explicit datetime
// DATE = explicit date only
func (p *Parser) parseEventTime(timeStr string) time.Time {
	// split time parameters and value
	tmSlice := strings.Split(timeStr, ":")
	tmParams := strings.Split(tmSlice[0], ";")
	var t time.Time

	format := icsFormat
	var loc *time.Location
	for _, param := range tmParams {
		if param == `VALUE=DATE` {
			format = icsFormatWholeDay
		}
		if result := re.tzid.FindStringSubmatch(param); len(result) > 0 {
			// found timezone
			var err error
			loc, err = time.LoadLocation(result[1])
			if err != nil {
				loc = nil
			}
		}
	}
	if format == icsFormat {
		if !strings.Contains(tmSlice[1], "Z") {
			tmSlice[1] = fmt.Sprintf("%sZ", tmSlice[1])
		}
	}
	if loc != nil {
		t, _ = time.ParseInLocation(format, tmSlice[1], loc)
	} else {
		t, _ = time.Parse(format, tmSlice[1])
	}
	return t
}

// parses the event start time
func (p *Parser) parseEventStart(eventData string) time.Time {
	var t time.Time
	if info := p.extractData(re.dtstart, eventData); info != "" {
		t = p.parseEventTime(info)
	}
	return t
}

// parses the event end time
func (p *Parser) parseEventEnd(eventData string) time.Time {
	var t time.Time
	if info := p.extractData(re.dtend, eventData); info != "" {
		t = p.parseEventTime(info)
	}
	return t
}

// parses the event RRULE (the repeater)
func (p *Parser) parseEventRRule(eventData string) string {
	return p.extractData(re.rrule, eventData)
}

// parses the event LOCATION
func (p *Parser) parseEventLocation(eventData string) string {
	return p.extractData(re.location, eventData)
}

// parses the event GEO
func (p *Parser) parseEventGeo(eventData string) *Geo {
	geo := p.extractData(re.geo, eventData)
	values := strings.Split(geo, ";")
	if len(values) < 2 {
		return nil
	}

	return NewGeo(values[0], values[1])
}

// ======================== ATTENDEE PARSING ===================

// parses the event attendees
func (p *Parser) parseEventAttendees(eventData string) []*Attendee {
	attendeesObj := []*Attendee{}
	attendees := re.attendee.FindAllString(eventData, len(eventData))

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
	organizerData := re.organizer.FindString(eventData)
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
	return p.extractData(re.mail, attendeeData)
}

// parses the attendee status
func (p *Parser) parseAttendeeStatus(attendeeData string) string {
	return p.extractData(re.partStat, attendeeData)
}

// parses the attendee role
func (p *Parser) parseAttendeeRole(attendeeData string) string {
	return p.extractData(re.role, attendeeData)
}

// parses the attendee Name
func (p *Parser) parseAttendeeName(attendeeData string) string {
	return p.extractData(re.attName, attendeeData)
}

// parses the organizer Name
func (p *Parser) parseOrganizerName(orgData string) string {
	return p.extractData(re.orgName, orgData)
}

// parses the attendee type
func (p *Parser) parseAttendeeType(attendeeData string) string {
	return p.extractData(re.cuType, attendeeData)
}

func parseDayNameToIcsName(day string) string {
	var dow string
	switch day {
	case "Mon":
		dow = "MO"
		break
	case "Tue":
		dow = "TU"
		break
	case "Wed":
		dow = "WE"
		break
	case "Thu":
		dow = "TH"
		break
	case "Fri":
		dow = "FR"
		break
	case "Sat":
		dow = "ST"
		break
	case "Sun":
		dow = "SU"
		break
	default:
		// fmt.Println("DEFAULT :", start.Format("Mon"))
		dow = ""
		break
	}
	return dow
}

func (p *Parser) extractData(r *regexp.Regexp, str string) string {
	data := r.FindStringSubmatch(str)
	if len(data) > 1 {
		return data[1]
	}
	return ""
}

func (p *Parser) atomicStatusCalendars(value int) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.statusCalendars += value
}

func (p *Parser) atomicAddError(err error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.errorsOccured = append(p.errorsOccured, err)
}
