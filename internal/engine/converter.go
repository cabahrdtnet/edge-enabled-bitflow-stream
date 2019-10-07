package engine

import (
	"fmt"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// layout
const (
	tLayout = "2006-01-02T15:04:05.000000000"
	layout  = "2006-01-02 15:04:05.000000000"
)

// header for stream based on readings
func header(readings []models.Reading) string {
	header := "time,tags"
	for _, r := range readings {
		header += "," + r.Name
	}
	return header
}

// convert EdgeX event bitflow csv sample
func etos(e models.Event) (string, error) {
	if cmp.Equal(e, models.Event{}, cmpopts.IgnoreUnexported(models.Event{})) {
		return "", fmt.Errorf("event is empty")
	}

	nanos := e.Origin * 1000 * 1000
	origin := time.Unix(0, nanos).UTC().Format(tLayout)
	sample := time.Now().UTC().Format(layout)

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
func stoe(deviceName string, sample string, header string) (models.Event, error) {
	if sample == "" || deviceName == "" || header == "" {
		return models.Event{}, fmt.Errorf(
			"sample, deviceName or header are empty (sample:%s, deviceName:%s, header:%s)",
			sample, deviceName, header)
	}

	headerEntries := strings.Split(header, ",")
	entries := strings.Split(sample, ",")

	metrics := entries[2:]
	metricNames := headerEntries[2:]

	now := time.Now().UTC().Format(layout)
	eventTime, err := time.Parse(layout, now)
	if err != nil {
		return models.Event{}, fmt.Errorf("parsing error in stoe: %s", err.Error())
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

// replace input and output word in Configuration string by std://-
func mapIO(script string) (string, error) {
	if script == "" {
		return script, fmt.Errorf("script is empty")
	}
	const stdin = "std://-"
	const stdout = "std+csv://-"
	scriptCopy := script

	unchanged := scriptCopy
	r, _ := regexp.Compile(`\s+`)
	scriptCopy = r.ReplaceAllString(scriptCopy, " ")
	scriptCopy = strings.TrimSpace(script)

	unchanged = scriptCopy
	r, _ = regexp.Compile(`^input`)
	scriptCopy = r.ReplaceAllString(scriptCopy, stdin)
	if scriptCopy == unchanged {
		return script, fmt.Errorf("leading input designator not found")
	}

	unchanged = scriptCopy
	r, _ = regexp.Compile(`output$`)
	scriptCopy = r.ReplaceAllString(scriptCopy, stdout)
	if scriptCopy == unchanged {
		return script, fmt.Errorf("trailing output designator not found")
	}

	return scriptCopy, nil
}

func typeOf(value string) string {
	if value == "" {
		return "S"
	}

	_, err := strconv.ParseInt(value, 10, 64)
	if err == nil {
		return "I"
	}

	_, err = strconv.ParseFloat(value, 64)
	if err == nil {
		return "F"
	}

	_, err = strconv.ParseBool(value)
	if err == nil {
		return "B"
	}

	return "S"
}