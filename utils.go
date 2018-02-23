package ics

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	// temporary path for downloading calendars
	filePath = "tmp/"

	//  unixtimestamp
	uts = "1136239445"

	//ics date time format
	icsFormat = "20060102T150405Z"

	// ics date format
	icsFormatWholeDay = "20060102"

	// Y-m-d H:i:S time format
	ymdHis = "2006-01-02 15:04:05"
)

// downloads the calendar before parsing it
func downloadFromUrl(url string) (string, error) {
	// split the url to get the name of the file (like basic.ics)
	tokens := strings.Split(url, "/")

	// create the name of the file
	fileName := fmt.Sprintf("%s%s_%s", filePath, time.Now().Format(uts), tokens[len(tokens)-1])

	// creates the path
	os.MkdirAll(filePath, 0777)

	// creates the file in the path folder
	output, err := os.Create(fileName)

	if err != nil {

		return "", err
	}
	// close the file
	defer output.Close()

	// get the URL
	response, err := http.Get(url)

	if err != nil {

		return "", err
	}
	// close the response body
	defer response.Body.Close()

	if response.StatusCode != 200 {

		return "", fmt.Errorf("url returned status code: %v", response.StatusCode)
	}
	// copy the response from the url to the temp local file
	_, err = io.Copy(output, response.Body)

	if err != nil {

		return "", err
	}

	//return the file that contains the info
	return fileName, nil
}

//  checks if file exists
func fileExists(fileName string) bool {
	_, err := os.Stat(fileName)
	return err == nil
}
