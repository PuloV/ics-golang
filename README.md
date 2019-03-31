# ics-golang
Golang ics parser

## Installation
`go get github.com/PuloV/ics-golang`

## How to use it
* Import the package:
```sh
import (
	"github.com/PuloV/ics-golang"
)
```
* Create a new parser:
```sh
    parser := ics.New()
```
* Pass as many ics urls as you want to the input chan :
```sh
    parserChan := parser.GetInputChan()
    parserChan <- "http://www.google.com/calendar/ical/bg.bulgarian%23holiday%40group.v.calendar.google.com/public/basic.ics"
    parserChan <- "http://www.google.com/calendar/ical/en.bulgarian%23holiday%40group.v.calendar.google.com/public/basic.ics"
    parserChan <- "http://www.google.com/calendar/ical/de.bulgarian%23holiday%40group.v.calendar.google.com/public/basic.ics"
```
###### * don't rely that the calendars will be parsed in this order
* Wait for the result of the parsing :
```sh

    outputChan := parser.GetOutputChan()
    //  print events
	go func() {
		for event := range outputChan {
			fmt.Println(event.GetImportedID())
		}
	}()

	// wait to kill the main goroute
	parser.Wait()
```
###### * the data form the calendars may be mixed

## Different usage
You can see diferent usage in the [ics-golang-examples](https://github.com/PuloV/ics-golang-examples) or in the test files `<filename>_test.go`

## LICENCE
The MIT License (MIT)

Copyright (c) 2014 Йордан Пулов

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

