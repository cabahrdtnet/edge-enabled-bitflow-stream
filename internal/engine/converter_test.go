package engine

import (
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"testing"
	"time"
)

var (
	humancount = models.Reading{
		Name:        "humancount",
		Value:       "1",
		Origin:      1471806386919,
	}

	caninecount = models.Reading{
		Name:        "caninecount",
		Value:       "0",
		Origin:      1471806386919,
	}

	event = models.Event{
		Device:   "countcamera1",
		Created:  0,
		Modified: 0,
		Origin:   1471806386919,
		Readings: []models.Reading{
			humancount,
			caninecount},
	}
)

func TestHeader_EmptyReadings_InitialHeader(t *testing.T) {
	// arrange
	readings := []models.Reading{}

	// act
	header := header(readings)

	// assert
	expected := "time,tags"
	if header != expected {
		t.Errorf("Wrong bitflow header!\n-Result was: %s\n-Expected was: %s", header, expected)
	}
}

func TestHeader_AverageReadings_CorrectHeader(t *testing.T) {
	// arrange
	readings := event.Readings

	// act
	header := header(readings)

	// assert
	expected := "time,tags,humancount,caninecount"
	if header != expected {
		t.Errorf("Wrong bitflow header!\n-Result was: %s\n-Expected was: %s", header, expected)
	}
}

func TestEtos_EmptyInput_EmptyOutput(t *testing.T) {
	// arrange
	event := models.Event{}

	// act
	sample, err := etos(event)

	// assert
	expected := ``
	if sample != expected || err == nil {
		t.Errorf("Wrong bitflow csv sample!\n-Result was: %s\n-Expected was: %s", sample, expected)
	}
}

func TestEtos_AverageInput_SuccessfulConversion(t *testing.T) {
	// arrange
	now = time.Unix(0, 1569845752091497000).UTC().Format("2006-01-02 15:04:05.000000000")

	// act
	sample, err := etos(event)

	// assert
	expected := `2019-09-30 12:15:52.091497000,origin=2016-08-21T19:06:26.919000000 device=countcamera1,1,0`
	if sample != expected || err != nil {
		t.Errorf("Wrong bitflow csv sample!\n-Result was: %s\n-Expected was: %s", sample, expected)
	}

	// clean up
	now = time.Now().UTC().Format("2006-01-02 15:04:05.000000000")
}

func TestStoe_EmptyInput_EmptyOutput(t *testing.T) {
	// arrange
	header := "time,tags"
	str := ``

	// act
	event, err := stoe("", str, header)

	// assert
	expected := models.Event{}

	if ! cmp.Equal(event, expected, cmpopts.IgnoreUnexported(models.Event{})) || err == nil {
		t.Errorf("Wrong event !\n-Result was: %s\n-Expected was: %s", event, expected)
	}
}

func TestStoe_AverageInput_SuccessfulConversion(t *testing.T) {
	// arrange
	header := "time,tags,humancount,caninecount"
	str := `2019-09-30 12:15:52.091497000,origin=2016-08-21T19:06:26.919000000 device="countcamera1",1,0`
	now = time.Unix(0, 1471806386919000000).UTC().Format("2006-01-02 15:04:05.000000000")
	// act
	resultEvent, err := stoe("countcamera1", str, header)

	// assert
	expected := event
	if ! cmp.Equal(resultEvent, expected, cmpopts.IgnoreUnexported(models.Event{}),
		cmpopts.IgnoreUnexported(models.Reading{})) || err != nil {
		t.Errorf("Wrong event !\n-Result was: %s\n-Expected was: %s", resultEvent, expected)
	}

	// clean up
	now = time.Now().UTC().Format("2006-01-02 15:04:05.000000000")
}

func TestMapIO_EmptyScript_ShouldReturnEmptyStringAndError(t *testing.T) {
	// arrange
	str := ""

	// act
	result, err := mapIO(str)

	// assert
	expected := ""
	if result != expected || err == nil {
		t.Errorf("Can't handle empty script!\n-Result was: %s\n-Expected was: %s", result, expected)
	}
}

func TestMapIO_ErrorInValuesToMap_ShouldReturnFaultyScriptAndError(t *testing.T) {
	// arrange
	str := "iput -> avg() -> outpit"

	// act
	result, err := mapIO(str)

	// assert
	expected := "iput -> avg() -> outpit"
	if result != expected || err == nil {
		t.Errorf("Can't handle faulty script!\n-Result was: %s\n-Expected was: %s", result, expected)
	}
}

func TestMapIO_MinimalScript_ShouldSucceed(t *testing.T) {
	// arrange
	str := "input -> output"

	// act
	result, err := mapIO(str)

	// assert
	expected := "std://- -> std+csv://-"
	if result != expected || err != nil {
		t.Errorf("Can't handle minimal script!\n-Result was: %s\n-Expected was: %s", result, expected)
	}
}

func TestMapIO_MinimalScriptWithWhiteSpaces_ShouldSucceed(t *testing.T) {
	// arrange
	str := `	
			    input -> output	
	              
			`

	// act
	result, err := mapIO(str)

	// assert
	expected := "std://- -> std+csv://-"
	if result != expected || err != nil {
		t.Errorf("Can't handle whitespaced script!\n-Result was: %s\n-Expected was: %s", result, expected)
	}
}

func TestMapIO_AverageScript_ShouldSucceed(t *testing.T) {
	// arrange
	str := `input -> avg() -> do(expr='set("input1",output * 10, set("field2", now(), set_tag("input", "output")))) -> output`

	// act
	result, err := mapIO(str)

	// assert
	expected := `std://- -> avg() -> do(expr='set("input1",output * 10, set("field2", now(), set_tag("input", "output")))) -> std+csv://-`
	if result != expected || err != nil {
		t.Errorf("Can't handle average script!\n-Result was: %s\n-Expected was: %s", result, expected)
	}
}

// int -> I; float -> F; bool -> B; string -> S
func TestTypeOf_EmptyString_ReturnS(t *testing.T) {
	// arrange
	str := ""

	// act
	result := typeOf(str)

	// assert
	expected := "S"
	if result != expected {
		t.Errorf("Couldn't parse supplied string!\n-Result was: %s\n-Expected was: %s", result, expected)
	}
}

func TestTypeOf_Integer_ReturnI(t *testing.T) {
	// arrange
	str := "1"

	// act
	result := typeOf(str)

	// assert
	expected := "I"
	if result != expected {
		t.Errorf("Couldn't parse supplied string!\n-Result was: %s\n-Expected was: %s", result, expected)
	}
}

func TestTypeOf_Float_ReturnF(t *testing.T) {
	// arrange
	str := "3.14"

	// act
	result := typeOf(str)

	// assert
	expected := "F"
	if result != expected {
		t.Errorf("Couldn't parse supplied string!\n-Result was: %s\n-Expected was: %s", result, expected)
	}
}

func TestTypeOf_Bool_ReturnB(t *testing.T) {
	// arrange
	str := "true"

	// act
	result := typeOf(str)

	// assert
	expected := "B"
	if result != expected {
		t.Errorf("Couldn't parse supplied string!\n-Result was: %s\n-Expected was: %s", result, expected)
	}
}

func TestTypeOf_StringOrOtherObject_ReturnS(t *testing.T) {
	// arrange
	str := "such a short string"

	// act
	result := typeOf(str)

	// assert
	expected := "S"
	if result != expected {
		t.Errorf("Couldn't parse supplied string!\n-Result was: %s\n-Expected was: %s", result, expected)
	}
}