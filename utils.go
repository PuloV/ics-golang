package ics

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

// Describes the file path to the folder with the temp ics files
var FilePath string

// if RepeatRuleApply is true , the rrule will create new objects for the repeated events
var RepeatRuleApply bool

// max of the rrule repeat for single event
var MaxRepeats int

//  unixtimestamp
const uts = "1136239445"

//ics date time format
const IcsFormat = "20060102T150405Z"

// Y-m-d H:i:S time format
const YmdHis = "2006-01-02 15:04:05"

// ics date format ( describes a whole day)
const IcsFormatWholeDay = "20060102"

// downloads the calendar before parsing it
func downloadFromUrl(url string) (fileName string, err error) {
	tokens := strings.Split(url, "/")
	fileName = fmt.Sprintf("%s%s_%s", FilePath, time.Now().Format(uts), tokens[len(tokens)-1])
	os.MkdirAll(FilePath, 0777)

	output, err := os.Create(fileName)
	if err != nil {
		return
	}
	defer output.Close()

	response, err := http.Get(url)
	if err != nil {
		return
	}
	defer response.Body.Close()

	_, err = io.Copy(output, response.Body)
	if err != nil {
		return
	}

	return fileName, nil
}

func trimField(field, cutset string) string {
	re, _ := regexp.Compile(cutset)
	cutsetRem := re.ReplaceAllString(field, "")
	return strings.TrimRight(cutsetRem, "\r\n")
}

func fileExists(fileName string) bool {
	_, err := os.Stat(fileName)
	return err == nil
}

func parseDayNameToIcsName(day string) string {
	var dow string

	switch day {
	case "Mon":
		dow = "MO"
	case "Tue":
		dow = "TU"
	case "Wed":
		dow = "WE"
	case "Thu":
		dow = "TH"
	case "Fri":
		dow = "FR"
	case "Sat":
		dow = "ST"
	case "Sun":
		dow = "SU"
	default:
		dow = ""
	}

	return dow
}
