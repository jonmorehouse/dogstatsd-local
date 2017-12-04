package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type dogstatsdJsonMetric struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Path      string `json:"path"`

	Value      float64  `json:"value"`
	Extras     []string `json:"extras"`
	SampleRate float64  `json:"sample_rate"`
	Tags       []string `json:"tags"`
}

func newJsonDogstatsdMsgHandler(extraTags []string) msgHandler {
	return func(msg []byte) error {
		dMsg, err := parseDogstatsdMsg(msg)
		if err != nil {
			log.Println(err.Error())
		}

		if dMsg.Type() != metricMsgType {
			log.Println("Unable to serialize non metric messages to JSON yet")
			return nil
		}

		metric, ok := dMsg.(dogstatsdMetric)
		if !ok {
			log.Fatalf("Programming error: invalid Type() = type matching")
		}

		jsonMsg := dogstatsdJsonMetric{
			Namespace:  metric.namespace,
			Name:       metric.name,
			Path:       fmt.Sprintf("%s.%s", metric.namespace, metric.name),
			Value:      metric.floatValue,
			Extras:     metric.extras,
			SampleRate: metric.sampleRate,
			Tags:       metric.tags,
		}

		enc := json.NewEncoder(os.Stdout)
		if err := enc.Encode(&jsonMsg); err != nil {
			log.Println("JSON serialize error:", err.Error())
			return nil
		}

		return nil
	}
}

func newHumanDogstatsdMsgHandler(extraTags []string) msgHandler {
	return func(msg []byte) error {
		dMsg, err := parseDogstatsdMsg(msg)
		if err != nil {
			log.Println(err.Error())
			return nil
		}

		metric, ok := dMsg.(dogstatsdMetric)
		if dMsg.Type() != metricMsgType || !ok {
			return nil
		}

		tmpl := "metric:%s|%s.%s|%.2f"
		str := fmt.Sprintf(tmpl, metric.metricType.String(), metric.namespace, metric.name, metric.floatValue)

		if metric.metricType == timerMetricType {
			str += "ms"
		}

		// iterate through tags
		for _, tag := range append(extraTags, metric.tags...) {
			str += " " + tag
		}

		fmt.Fprintf(os.Stdout, str)
		fmt.Fprintf(os.Stdout, "\n")
		return nil
	}
}

func newRawDogstatsdMsgHandler() msgHandler {
	return func(msg []byte) error {
		fmt.Fprintf(os.Stdout, string(msg))
		fmt.Fprintf(os.Stdout, "\n")
		return nil
	}
}
