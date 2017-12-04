package main

import (
	"bytes"
	"errors"
	"strconv"
	"strings"
	"time"
)

type dogstatsdMsgType int

func (d dogstatsdMsgType) String() string {
	switch d {
	case metricMsgType:
		return "metric"
	case eventMsgType:
		return "event"
	case serviceCheckMsgType:
		return "service_check"
	}

	return "unknown"
}

const (
	metricMsgType dogstatsdMsgType = iota
	serviceCheckMsgType
	eventMsgType
)

type dogstatsdMsg interface {
	Type() dogstatsdMsgType
	Data() []byte
}

func parseDogstatsdMetricMsg(buf []byte) (dogstatsdMsg, error) {
	metric := dogstatsdMetric{
		ts:         time.Now(),
		tags:       make([]string, 0),
		sampleRate: 1.0,
	}

	// sample message: metric.name:value|type|@sample_rate|#tag1:value,tag2
	pieces := strings.Split(string(buf), "|")
	if len(pieces) < 2 {
		return nil, errors.New("INVALID_MSG_MISSING_NAME_VALUE_OR_TYPE")
	}

	addrAndValue := strings.Split(pieces[0], ":")
	if len(addrAndValue) < 2 {
		return nil, errors.New("INVALID_MSG_MISSING_NAME_AND_VALUE")
	}

	namespaceAndName := strings.SplitN(addrAndValue[0], ".", 2)
	if len(namespaceAndName) > 1 {
		metric.namespace = namespaceAndName[0]
		metric.name = namespaceAndName[1]
	} else {
		metric.name = namespaceAndName[0]
	}

	metric.rawValue = addrAndValue[1]

	switch pieces[1] {
	case "c":
		metric.metricType = counterMetricType
	case "g":
		metric.metricType = gaugeMetricType
	case "s":
		metric.metricType = setMetricType
	case "ms":
		metric.metricType = timerMetricType
	case "h":
		metric.metricType = histogramMetricType
	default:
		return nil, errors.New("INVALID_MSG_INVALID_TYPE")
	}

	// all values are stored as a float
	floatValue, err := strconv.ParseFloat(metric.rawValue, 64)
	if err != nil {
		return nil, errors.New("INVALID_MSG_INVALID_VALUE")
	}
	metric.floatValue = floatValue

	if metric.metricType == timerMetricType {
		metric.durationValue = time.Duration(metric.floatValue) / time.Millisecond
	}

	// parse out sample rate, tags and any extras
	for _, piece := range pieces[2:] {
		if strings.HasPrefix(piece, "@") {
			sampleRate, err := strconv.ParseFloat(piece[1:], 64)
			if err != nil {
				return nil, errors.New("E_INVALID_SAMPLE_RATE")
			}
			metric.sampleRate = sampleRate
			continue
		}

		if strings.HasPrefix(piece, "#") {
			tags := strings.Split(piece[1:], ",")
			metric.tags = append(metric.tags, tags...)
			continue
		}

		metric.extras = append(metric.extras, piece)
	}

	return metric, nil
}

type dogstatsdMetricType int

func (d dogstatsdMetricType) String() string {
	switch d {
	case gaugeMetricType:
		return "gauge"
	case counterMetricType:
		return "counter"
	case setMetricType:
		return "set"
	case timerMetricType:
		return "timer"
	case histogramMetricType:
		return "histogram"
	}

	return "unknown"
}

const (
	gaugeMetricType dogstatsdMetricType = iota
	counterMetricType
	setMetricType
	timerMetricType
	histogramMetricType
)

type dogstatsdMetric struct {
	data []byte
	ts   time.Time

	namespace string
	name      string

	metricType    dogstatsdMetricType
	rawValue      string
	floatValue    float64
	durationValue time.Duration

	extras     []string
	tags       []string
	sampleRate float64
}

func (d dogstatsdMetric) Data() []byte {
	return d.data
}

func (d dogstatsdMetric) Type() dogstatsdMsgType {
	return metricMsgType
}

func parseDogstatsdServiceCheckMsg(buf []byte) (dogstatsdMsg, error) {
	return nil, errors.New("dogstatsd service check messages not supported ...")
}

// service check
type dogstatsdServiceCheck struct {
	data []byte
}

func (dogstatsdServiceCheck) Type() dogstatsdMsgType {
	return serviceCheckMsgType
}

func (d dogstatsdServiceCheck) Data() []byte {
	return d.data
}

func parseDogstatsdEventMsg(buf []byte) (dogstatsdMsg, error) {
	return nil, errors.New("dogstatsd event messages not supported ...")
}

type dogstatsdEvent struct {
	data []byte
}

// parse a dogstatsdMsg, returning the correct message back
func parseDogstatsdMsg(buf []byte) (dogstatsdMsg, error) {
	if bytes.HasPrefix(buf, []byte("_e{")) {
		return parseDogstatsdEventMsg(buf)
	}

	if bytes.HasPrefix(buf, []byte("_sc{")) {
		return parseDogstatsdServiceCheckMsg(buf)
	}

	return parseDogstatsdMetricMsg(buf)
}
