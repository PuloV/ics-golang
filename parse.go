package icsgolang

import "fmt"

type Parser struct {
	input_chan chan string
}

func New() chan string {
	p := new(Parser)
	// p.input_chan := make(chan string)
	go func(input chan string) {
		for link := range input {
			fmt.Println(link)
		}
	}(p.input_chan)
	return p.input_chan
}
