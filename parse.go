package ics

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	urlRegex    = regexp.MustCompile(`http(s){0,1}:\/\/`)
	eventsRegex = regexp.MustCompile(`(BEGIN:VEVENT(.*\n)*?END:VEVENT\r?\n)`)

	calNameRegex     = regexp.MustCompile(`X-WR-CALNAME:.*?\n`)
	calDescRegex     = regexp.MustCompile(`X-WR-CALDESC:.*?\n`)
	calVersionRegex  = regexp.MustCompile(`VERSION:.*?\n`)
	calTimezoneRegex = regexp.MustCompile(`X-WR-TIMEZONE:.*?\n`)

	eventSummaryRegex       = regexp.MustCompile(`SUMMARY:.*?\n`)
	eventStatusRegex        = regexp.MustCompile(`STATUS:.*?\n`)
	eventDescRegex          = regexp.MustCompile(`DESCRIPTION:.*?\n`)
	eventUIDRegex           = regexp.MustCompile(`UID:.*?\n`)
	eventClassRegex         = regexp.MustCompile(`CLASS:.*?\n`)
	eventSequenceRegex      = regexp.MustCompile(`SEQUENCE:.*?\n`)
	eventCreatedRegex       = regexp.MustCompile(`CREATED:.*?\n`)
	eventModifiedRegex      = regexp.MustCompile(`LAST-MODIFIED:.*?\n`)
	eventStartRegex         = regexp.MustCompile(`DTSTART(;TZID=.*?){0,1}:.*?\n`)
	eventStartWholeDayRegex = regexp.MustCompile(`DTSTART;VALUE=DATE:.*?\n`)
	eventEndRegex           = regexp.MustCompile(`DTEND(;TZID=.*?){0,1}:.*?\n`)
	eventEndWholeDayRegex   = regexp.MustCompile(`DTEND;VALUE=DATE:.*?\n`)
	eventRRuleRegex         = regexp.MustCompile(`RRULE:.*?\n`)
	eventLocationRegex      = regexp.MustCompile(`LOCATION:.*?\n`)

	attendeesRegex = regexp.MustCompile(`ATTENDEE(:|;)(.*?\r?\n)(\s.*?\r?\n)*`)
	organizerRegex = regexp.MustCompile(`ORGANIZER(:|;)(.*?\r?\n)(\s.*?\r?\n)*`)

	attendeeEmailRegex  = regexp.MustCompile(`mailto:.*?\n`)
	attendeeStatusRegex = regexp.MustCompile(`PARTSTAT=.*?;`)
	attendeeRoleRegex   = regexp.MustCompile(`ROLE=.*?;`)
	attendeeNameRegex   = regexp.MustCompile(`CN=.*?;`)
	organizerNameRegex  = regexp.MustCompile(`CN=.*?:`)
	attendeeTypeRegex   = regexp.MustCompile(`CUTYPE=.*?;`)

	untilRegex    = regexp.MustCompile(`UNTIL=(\d)*T(\d)*Z(;){0,1}`)
	intervalRegex = regexp.MustCompile(`INTERVAL=(\d)*(;){0,1}`)
	countRegex    = regexp.MustCompile(`COUNT=(\d)*(;){0,1}`)
	freqRegex     = regexp.MustCompile(`FREQ=.*?;`)
	byMonthRegex  = regexp.MustCompile(`BYMONTH=.*?;`)
	byDayRegex    = regexp.MustCompile(`BYDAY=.*?(;|){0,1}\z`)
)

func init() {
	FilePath = "tmp/"
	RepeatRuleApply = false
	MaxRepeats = 10
}

func ParseCalendar(url string) (Calendar, error) {
	content, err := getICal(url)
	if err != nil {
		return Calendar{}, err
	}

	return parseICalContent(content, url), nil
}

func getICal(url string) (fileName string, err error) {
	var isRemote = urlRegex.FindString(url) != ""

	if isRemote {
		fileName, err = downloadFromUrl(url)
		if err != nil {
			return
		}
	} else {
		if fileExists(url) {
			fileName = url
		} else {
			err = errors.New(fmt.Sprintf("File %s does not exists", url))
			return
		}
	}

	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		return
	}

	if isRemote {
		os.Remove(fileName)
	}

	return string(content), nil
}

func parseICalContent(content, url string) Calendar {
	cal := NewCalendar()
	eventsData, info := explodeICal(content)
	cal.Name = parseICalName(info)
	cal.Description = parseICalDesc(info)
	cal.Version = parseICalVersion(info)
	cal.Timezone = parseICalTimezone(info)
	cal.URL = url
	parseEvents(&cal, eventsData)
	return cal
}

func explodeICal(content string) ([]string, string) {
	events := eventsRegex.FindAllString(content, -1)
	info := eventsRegex.ReplaceAllString(content, "")
	return events, info
}

func parseICalName(content string) string {
	return trimField(calNameRegex.FindString(content), "X-WR-CALNAME:")
}

func parseICalDesc(content string) string {
	return trimField(calDescRegex.FindString(content), "X-WR-CALDESC:")
}

func parseICalVersion(content string) float64 {
	version, _ := strconv.ParseFloat(trimField(calVersionRegex.FindString(content), "VERSION:"), 64)
	return version
}

func parseICalTimezone(content string) *time.Location {
	timezone := trimField(calTimezoneRegex.FindString(content), "X-WR-TIMEZONE:")
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Local
	}

	return loc
}

func parseEvents(cal *Calendar, eventsData []string) {
	for _, eventData := range eventsData {
		event := NewEvent()

		start := parseEventStart(eventData)
		end := parseEventEnd(eventData)
		wholeDay := start.Hour() == 0 && end.Hour() == 0 && start.Minute() == 0 && end.Minute() == 0 && start.Second() == 0 && end.Second() == 0

		event.Status = parseEventStatus(eventData)
		event.Summary = parseEventSummary(eventData)
		event.Description = parseEventDescription(eventData)
		event.ImportedID = parseEventId(eventData)
		event.Class = parseEventClass(eventData)
		event.Sequence = parseEventSequence(eventData)
		event.Created = parseEventCreated(eventData)
		event.Modified = parseEventModified(eventData)
		event.RRule = parseEventRRule(eventData)
		event.Location = parseEventLocation(eventData)
		event.Start = start
		event.End = end
		event.WholeDayEvent = wholeDay
		event.Attendees = parseEventAttendees(eventData)
		event.Organizer = parseEventOrganizer(eventData)
		event.ID = event.GenerateEventId()
		duration := end.Sub(start)
		cal.Events = append(cal.Events, *event)

		if RepeatRuleApply && event.RRule != "" {
			until := parseUntil(event.RRule)
			interval := parseInterval(event.RRule)
			count := parseCount(event.RRule)
			freq := trimField(freqRegex.FindString(event.RRule), `(FREQ=|;)`)
			byMonth := trimField(byMonthRegex.FindString(event.RRule), `(BYMONTH=|;)`)
			byDay := trimField(byDayRegex.FindString(event.RRule), `(BYDAY=|;)`)

			var years, days, months int

			switch freq {
			case "DAILY":
				days = interval
			case "WEEKLY":
				days = 7
			case "MONTHLY":
				months = interval
			case "YEARLY":
				years = interval
			}

			current := 0
			freqDate := start

			for {
				weekDays := freqDate
				commitEvent := func() {
					current++
					count--
					newEvent := event.Clone()
					newEvent.Start = weekDays
					newEvent.End = weekDays.Add(duration)
					newEvent.ID = newEvent.GenerateEventId()
					newEvent.Sequence = current
					if until.IsZero() || (!until.IsZero() && until.Format(YmdHis) >= weekDays.Format(YmdHis)) {
						cal.Events = append(cal.Events, *newEvent)
					}
				}

				if byMonth == "" || strings.Contains(byMonth, weekDays.Format("1")) {
					if byDay != "" {
						for i := 0; i < 7; i++ {
							day := parseDayNameToIcsName(weekDays.Format("Mon"))
							if strings.Contains(byDay, day) && weekDays != start {
								commitEvent()
							}
							weekDays = weekDays.AddDate(0, 0, 1)
						}
					} else {
						if weekDays != start {
							commitEvent()
						}
					}
				}

				freqDate = freqDate.AddDate(years, months, days)
				if current > MaxRepeats || count == 0 {
					break
				}

				if !until.IsZero() && until.Format(YmdHis) <= freqDate.Format(YmdHis) {
					break
				}
			}
		}
	}
}

func parseEventSummary(eventData string) string {
	return trimField(eventSummaryRegex.FindString(eventData), "SUMMARY:")
}

func parseEventStatus(eventData string) string {
	return trimField(eventStatusRegex.FindString(eventData), "STATUS:")
}

func parseEventDescription(eventData string) string {
	return trimField(eventDescRegex.FindString(eventData), "DESCRIPTION:")
}

func parseEventId(eventData string) string {
	return trimField(eventUIDRegex.FindString(eventData), "UID:")
}

func parseEventClass(eventData string) string {
	return trimField(eventClassRegex.FindString(eventData), "CLASS:")
}

func parseEventSequence(eventData string) int {
	seq, _ := strconv.Atoi(trimField(eventSequenceRegex.FindString(eventData), "SEQUENCE:"))
	return seq
}

func parseEventCreated(eventData string) time.Time {
	created := trimField(eventCreatedRegex.FindString(eventData), "CREATED:")
	t, _ := time.Parse(IcsFormat, created)
	return t
}

func parseEventModified(eventData string) time.Time {
	date := trimField(eventModifiedRegex.FindString(eventData), "LAST-MODIFIED:")
	t, _ := time.Parse(IcsFormat, date)
	return t
}

func parseEventStart(eventData string) time.Time {
	var (
		t  time.Time
		tz string
	)

	resultWholeDay := eventStartWholeDayRegex.FindString(eventData)
	if resultWholeDay != "" {
		tz = trimField(resultWholeDay, "DTSTART;VALUE=DATE:")
		t, _ = time.Parse(IcsFormatWholeDay, tz)
	} else {
		result := eventStartRegex.FindString(eventData)
		tz = trimField(result, "DTSTART(;TZID=.*?){0,1}:")

		if !strings.Contains(tz, "Z") {
			tz = fmt.Sprintf("%sZ", tz)
		}

		t, _ = time.Parse(IcsFormat, tz)
	}

	return t
}

func parseEventEnd(eventData string) time.Time {
	var (
		t  time.Time
		tz string
	)

	resultWholeDay := eventEndWholeDayRegex.FindString(eventData)
	if resultWholeDay != "" {
		tz = trimField(resultWholeDay, "DTEND;VALUE=DATE:")
		t, _ = time.Parse(IcsFormatWholeDay, tz)
	} else {
		result := eventEndRegex.FindString(eventData)
		tz = trimField(result, "DTEND(;TZID=.*?){0,1}:")

		if !strings.Contains(tz, "Z") {
			tz = fmt.Sprintf("%sZ", tz)
		}
		t, _ = time.Parse(IcsFormat, tz)
	}

	return t
}

func parseEventRRule(eventData string) string {
	return trimField(eventRRuleRegex.FindString(eventData), "RRULE:")
}

func parseEventLocation(eventData string) string {
	return trimField(eventLocationRegex.FindString(eventData), "LOCATION:")
}

func parseEventAttendees(eventData string) []Attendee {
	attendeesList := []Attendee{}
	attendees := attendeesRegex.FindAllString(eventData, -1)

	for _, a := range attendees {
		if a == "" {
			continue
		}
		attendee := parseAttendee(strings.Replace(strings.Replace(a, "\r", "", 1), "\n ", "", 1))
		if attendee.Email != "" || attendee.Name != "" {
			attendeesList = append(attendeesList, attendee)
		}
	}

	return attendeesList
}

func parseEventOrganizer(eventData string) Attendee {
	organizer := organizerRegex.FindString(eventData)
	if organizer == "" {
		return Attendee{}
	}

	organizer = strings.Replace(strings.Replace(organizer, "\r", "", 1), "\n ", "", 1)
	return Attendee{
		Email: parseAttendeeMail(organizer),
		Name:  parseOrganizerName(organizer),
	}
}

func parseAttendee(data string) Attendee {
	return Attendee{
		Email:  parseAttendeeMail(data),
		Name:   parseAttendeeName(data),
		Role:   parseAttendeeRole(data),
		Status: parseAttendeeStatus(data),
		Type:   parseAttendeeType(data),
	}
}

func parseAttendeeMail(attendeeData string) string {
	return trimField(attendeeEmailRegex.FindString(attendeeData), "mailto:")
}

func parseAttendeeStatus(attendeeData string) string {
	return trimField(attendeeStatusRegex.FindString(attendeeData), `(PARTSTAT=|;)`)
}

func parseAttendeeRole(attendeeData string) string {
	return trimField(attendeeRoleRegex.FindString(attendeeData), `(ROLE=|;)`)
}

func parseAttendeeName(attendeeData string) string {
	return trimField(attendeeNameRegex.FindString(attendeeData), `(CN=|;)`)
}

func parseOrganizerName(orgData string) string {
	return trimField(organizerNameRegex.FindString(orgData), `(CN=|:)`)
}

func parseAttendeeType(attendeeData string) string {
	return trimField(attendeeTypeRegex.FindString(attendeeData), `(CUTYPE=|;)`)
}

func parseUntil(rrule string) time.Time {
	until := trimField(untilRegex.FindString(rrule), `(UNTIL=|;)`)
	var t time.Time
	if until == "" {
	} else {
		t, _ = time.Parse(IcsFormat, until)
	}
	return t
}

func parseInterval(rrule string) int {
	interval := trimField(intervalRegex.FindString(rrule), `(INTERVAL=|;)`)
	i, _ := strconv.Atoi(interval)
	if i == 0 {
		i = 1
	}

	return i
}

func parseCount(rrule string) int {
	c := trimField(countRegex.FindString(rrule), `(COUNT=|;)`)
	count, _ := strconv.Atoi(c)
	if count == 0 {
		count = MaxRepeats
	}

	return count
}
