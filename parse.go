package ics

import (
	"fmt"
	"io/ioutil"
	"regexp"
	// "strings"
	// "errors"
	// "fmt"
	// "io"
	// "net/http"
	// "os"
	"sync"
	// "time"
)

type Parser struct {
	inputChan chan string
	errChan   chan error
}

// creates new parser and chan for calendar
func New() chan string {
	p := new(Parser)
	p.inputChan = make(chan string)
	p.errChan = make(chan error)

	//  init only once the WaitGroup
	o.Do(func() {
		Wg = new(sync.WaitGroup)
		// fmt.Println(Wg)
	})

	go func(input chan string) {
		// fmt.Println("Goroute")
		for {
			link := <-input
			Wg.Add(1)
			go func(link string) {
				iCalContent, err := p.getICal(link)
				if err != nil {
					p.errChan <- err
					return
				}
				p.parseICalContent(iCalContent)
			}(link)
		}
	}(p.inputChan)
	// p.wg.Wait()
	return p.inputChan
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
	ical.SetName(p.parseICalName(iCalContent))
	ical.SetDesc(p.parseICalDesc(iCalContent))
	fmt.Println(ical.name)
	fmt.Println(ical.description)
	Wg.Done()
}

// parses the iCal Name
func (p *Parser) parseICalName(iCalContent string) string {
	re, _ := regexp.Compile(`X-WR-CALNAME:.*?\n`)
	result := re.Find(stringToByte(iCalContent))
	return trimField(fmt.Sprintf("%s", result), "X-WR-CALNAME:")
}

// parses the iCal description
func (p *Parser) parseICalDesc(iCalContent string) string {
	re, _ := regexp.Compile(`X-WR-CALDESC:.*?\n`)
	result := re.Find(stringToByte(iCalContent))
	return trimField(fmt.Sprintf("%s", result), "X-WR-CALDESC:")
}
