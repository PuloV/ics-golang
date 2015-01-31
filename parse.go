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

// parses the iCal formated string
func (p *Parser) parseICalContent(iCalContent string) {
	ical := NewCalendar()
	p.parsedCalendars = append(p.parsedCalendars, ical)

	// split the data into calendar info and events data
	_, calInfo := explodeICal(iCalContent)
	idCounter++
	ical.SetName(p.parseICalName(calInfo))
	ical.SetDesc(p.parseICalDesc(calInfo))
	ical.SetVersion(p.parseICalVersion(calInfo))
	ical.SetTimezone(p.parseICalTimezone(calInfo))
	fmt.Printf("%#v \n", ical)
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
	loc, err := time.LoadLocation(timezone)

	// if fails with the timezone => go UTC
	if err != nil {
		loc, err := time.LoadLocation("UTC")
	}
	return *loc
}
