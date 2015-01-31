package ics

import (
	"fmt"
	// "io/ioutil"
	"strings"
	// "errors"
	// "fmt"
	"io"
	"net/http"
	"os"
	// "regexp"
	"sync"
	"time"
)

var Wg *sync.WaitGroup
var o sync.Once
var mutex *sync.Mutex
var idCounter int

//  unixtimestamp
const uts = "1136239445"

// Waiting to get the job done
func Wait() {
	Wg.Wait()
}

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
	return strings.TrimRight(strings.Trim(field, cutset), "\r\n")
}
