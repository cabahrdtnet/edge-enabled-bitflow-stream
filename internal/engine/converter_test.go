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
	header := Header(readings)

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
	header := Header(readings)

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
	sample, err := Etos(event)

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
	sample, err := Etos(event)

	// assert
	expected := `2019-09-30 12:15:52.091497000,origin=2016-08-21T19:06:26.919000000 device=countcamera1,1,0`
	if sample != expected || err != nil {
		t.Errorf("Wrong bitflow csv sample!\n-Result was: %s\n-Expected was: %s", sample, expected)
	}

	// clean up
	now = time.Now().UTC().Format("2006-01-02 15:04:05.000000000")
}

// Etor
func TestStoe_EmptyInput_EmptyOutput(t *testing.T) {
	// arrange
	header := "time,tags"
	str := ``

	// act
	event, err := Stoe("", str, header)

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
	resultEvent, err := Stoe("countcamera1", str, header)

	// assert
	expected := event
	if ! cmp.Equal(resultEvent, expected, cmpopts.IgnoreUnexported(models.Event{}),
		cmpopts.IgnoreUnexported(models.Reading{})) || err != nil {
		t.Errorf("Wrong event !\n-Result was: %s\n-Expected was: %s", resultEvent, expected)
	}

	// clean up
	now = time.Now().UTC().Format("2006-01-02 15:04:05.000000000")
}