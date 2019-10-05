package engine

import (
	"fmt"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"regexp"
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
func header(readings []models.Reading) string {
	header := "time,tags"
	for _, r := range readings {
		header += "," + r.Name
	}
	return header
}

// value descriptors for EdgeX
func registerValueDescriptors(header string) {
	// time,tags,humancount,humancount_avg,caninecount,caninecount_avg
	//
	// TODO
	// set transmission QoS 2 or 3
	// receive first event
	// stop message receiving mqtt client (potentially incoming events should be queued now;
	// QoS 1 would lead to missing messages, which is potentially an option as well, as one or two or 10
	// missing events should not make a difference)
	// process sample in bitflow
	// read output header from bitflow
	// read first processed event from bitflow
	// stop StdoutReader
	//
	// from output header and first event: derive value descriptors via created Readings
	// marshal created VD and request DS (by publishing message over ReverseCommand topic)
	// to send them to metadata
	// await response from DS over ReverseCommandResponse topic
	// if positive:
	// send signal to StdoutReader for contine its execution
}

// convert EdgeX event bitflow csv sample
func etos(e models.Event) (string, error) {
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

func vtoh() {
	// Event: read VD from Event return header as string
	// example:
	//var (
	//	humancount = models.Reading{
	//		Name:        "humancount",
	//		Value:       "1",
	//		Origin:      1471806386919,
	//	}
	//
	//	caninecount = models.Reading{
	//		Name:        "caninecount",
	//		Value:       "0",
	//		Origin:      1471806386919,
	//	}
	//
	//	event = models.Event{
	//		Device:   "countcamera1",
	//		Created:  0,
	//		Modified: 0,
	//		Origin:   1471806386919,
	//		Readings: []models.Reading{
	//			humancount,
	//			caninecount},
	//	}
	//) -> time,tags,humancount,caninecount
	// Event -> Name of Readings -> (time tags) + Slice of Metrics ->
}