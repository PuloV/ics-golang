# ics-golang
Golang ics parser, taken from github.com/PuloV/ics-golang
Changes made to improve parsing plus fixes of race conditions

## Installation
`go get github.com/slockij/ics-golang`

## How to use it
* Import the package
* Create a new parser
* Pass as many ics urls as you want asynchronously (don't rely that the calendars will be parsed in this order)
* Waitfor parsing to complete

```sh
import (
	"github.com/slockij/ics-golang"
)

func main() {
    parser := ics.New()
    parser.LoadAsyncFromUrl("http://www.google.com/calendar/ical/bg.bulgarian%23holiday%40group.v.calendar.google.com/public/basic.ics")
    parser.LoadAsyncFromUrl("http://www.google.com/calendar/ical/en.bulgarian%23holiday%40group.v.calendar.google.com/public/basic.ics")
    parser.LoadAsyncFromUrl("http://www.google.com/calendar/ical/de.bulgarian%23holiday%40group.v.calendar.google.com/public/basic.ics")
	// wait to for parsing to be done
	parser.Wait()
}
```

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

