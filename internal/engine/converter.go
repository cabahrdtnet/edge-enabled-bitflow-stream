package engine

import (
	"fmt"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"strings"
	"time"
)

// layout
const (
	tLayout = "2006-01-02T15:04:05.000000000"
	layout  = "2006-01-02 15:04:05.000000000"
)

// current time for timestamps
var (
	now = time.Now().UTC().Format(layout)
)

// header for stream based on readings
func Header(readings []models.Reading) string {
	header := "time,tags"
	for _, r := range readings {
		header += "," + r.Name
	}
	return header
}

// convert EdgeX event bitflow csv sample
func Etos(e models.Event) (string, error) {
	if cmp.Equal(e, models.Event{}, cmpopts.IgnoreUnexported(models.Event{})) {
		return "", fmt.Errorf("event is empty")
	}

	nanos := e.Origin * 1000 * 1000
	origin := time.Unix(0, nanos).UTC().Format(tLayout)
	sample := now
	tags := ",origin=" + origin + " " + "device=" + e.Device
	sample += tags

	metrics := ""
	for _, r := range e.Readings {
		metrics += "," + r.Value
	}
	sample += metrics

	return sample, nil
}

// convert bitflow csv sample to EdgeX event
func Stoe(deviceName string, sample string, header string) (models.Event, error) {
	if sample == "" || deviceName == "" || header == "" {
		return models.Event{}, fmt.Errorf(
			"sample, deviceName or header are empty (sample:%s, deviceName:%s, header:%s)",
			sample, deviceName, header)
	}

	headerEntries := strings.Split(header, ",")
	entries := strings.Split(sample, ",")

	metrics := entries[2:]
	metricNames := headerEntries[2:]

	eventTime, err := time.Parse(layout, now)
	if err != nil {
		return models.Event{}, fmt.Errorf("parsing error in Stoe: %s", err.Error())
	}

	readings := []models.Reading{}
	for index := range metricNames {
		reading := models.Reading{
			Name:   metricNames[index],
			Value:  metrics[index],
			Origin: eventTime.UnixNano() / 1000 / 1000,
		}
		readings = append(readings, reading)
	}

	return models.Event{
		Device:   deviceName,
		Origin:   eventTime.UnixNano() / 1000 / 1000,
		Readings: readings,
	}, nil
}
