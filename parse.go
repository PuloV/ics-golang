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
	"sync"
	// "time"
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
	// _, calInfo := explodeICal("(BEGIN:VEVENT\ngot e\nEND:VEVENT)")
	_, calInfo := explodeICal(iCalContent)
	idCounter++
	ical.SetName(p.parseICalName(calInfo))
	ical.SetDesc(p.parseICalDesc(calInfo))
	fmt.Println(ical.name)
	fmt.Println(ical.description)
	Wg.Done()
}

// parses the iCal Name
func (p *Parser) parseICalName(iCalContent string) string {
	re, _ := regexp.Compile(`X-WR-CALNAME:.*?\n`)
	result := re.FindString(iCalContent)
	return trimField(fmt.Sprintf("%s", result), "X-WR-CALNAME:")
}

// parses the iCal description
func (p *Parser) parseICalDesc(iCalContent string) string {
	re, _ := regexp.Compile(`X-WR-CALDESC:.*?\n`)
	result := re.FindString(iCalContent)
	return trimField(fmt.Sprintf("%s", result), "X-WR-CALDESC:")
}

// explodes the ICal content to array of events and calendar info
func explodeICal(iCalContent string) ([]string, string) {
	// fmt.Println(iCalContent)
	reEvents, _ := regexp.Compile(`(BEGIN:VEVENT(.*\n)*?END:VEVENT\r\n)`)
	allEvents := reEvents.FindAllString(iCalContent, len(iCalContent))
	calInfo := reEvents.ReplaceAllString(iCalContent, "")

	return allEvents, calInfo
}
