package ics

import (
	"fmt"
	// "io/ioutil"
	"strings"
	// "errors"
	"io"
	"net/http"
	"os"
	"regexp"
	"sync"
	"time"
)

var o sync.Once
var mutex *sync.Mutex
var idCounter int
var DeleteTempFiles bool

//  unixtimestamp
const uts = "1136239445"

//ics date time format
const IcsFormat = "20060102T150405Z"

// Y-m-d H:i:S time format
const YmdHis = "2006-01-02 15:04:05"

// ics date format ( describes a whole day)
const IcsFormatWholeDay = "20060102"

// downloads the calendar before parsing it
func downloadFromUrl(url string) (string, error) {
	tokens := strings.Split(url, "/")
	fileName := fmt.Sprintf("%s_%s", time.Now().Format(uts), tokens[len(tokens)-1])
	// fmt.Println("Downloading", url, "to", fileName)

	// TODO: check file existence first with io.IsExist
	output, err := os.Create(fileName)
	if err != nil {
		// fmt.Println("Error while creating", fileName, "-", err)
		return "", err
	}
	defer output.Close()

	response, err := http.Get(url)
	if err != nil {
		// fmt.Println("Error while downloading", url, "-", err)
		return "", err
	}
	defer response.Body.Close()
	fmt.Println(response.Body)
	_, err = io.Copy(output, response.Body)
	if err != nil {
		// fmt.Println("Error while downloading", url, "-", err)
		return "", err
	}

	// fmt.Println(n, "bytes downloaded.")
	return fileName, nil
}

func stringToByte(str string) []byte {
	return []byte(str)
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
