package ics_test

import (
	"fmt"
	"github.com/PuloV/ics-golang"
	"reflect"
	"testing"
)

func TestNewParser(t *testing.T) {
	parser := ics.New()
	rType := fmt.Sprintf("%v", reflect.TypeOf(parser))
	if rType != "*ics.Parser" {
		t.Errorf("Failed to create a Parser !")
	}
}

func TestNewParserChans(t *testing.T) {
	parser := ics.New()
	input := parser.GetInputChan()
	output := parser.GetOutputChan()

	rType := fmt.Sprintf("%v", reflect.TypeOf(input))

	if rType != "chan string" {
		t.Errorf("Failed to create a input chan ! Received : Type %s Value %s", rType, input)
	}

	rType = fmt.Sprintf("%v", reflect.TypeOf(output))
	if rType != "chan ics.Event" {
		t.Errorf("Failed to create a output chan! Received : Type %s Value %s", rType, output)
	}
}
