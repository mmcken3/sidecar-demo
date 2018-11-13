package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

// Config is a struct for the env config of the project.
type Config struct {
	Endpoint string `envconfig:"SIDECAR_ENDPOINT" default:"127.0.0.1"`
	Port     int    `envconfig:"SIDECAR_PORT" default:"8125"`

	Namespace   string `envconfig:"DD_NAMESPACE"`
	Environment string `envconfig:"ENVIRONMENT"`
}

func main() {
	var config Config

	// Get the projects environment variables.
	err := envconfig.Process("DEMO", &config)
	if err != nil {
		log.Println(errors.Wrap(err, "envconfig: process"))
		return
	}

	// Created a new buffered client to connect to the DogStatsD sidecar.
	ddClient, err := statsd.NewBuffered(fmt.Sprintf("%v:%v", config.Endpoint, config.Port), 1)
	if err != nil {
		log.Println(errors.Wrap(err, "initialize dogstatsd"))
		return
	}

	// Prefix every metric with the app name, some common tags.
	ddClient.Namespace = config.Namespace

	// Use the environment to populate some tags.
	switch config.Environment {
	case "prod":
		ddClient.Tags = append(ddClient.Tags, "account:mmcken3-demos", "environment:prod")
	case "test":
		ddClient.Tags = append(ddClient.Tags, "account:mmcken3-demos-dev", "environment:test")
	default:
		ddClient.Tags = append(ddClient.Tags, "account:"+config.Environment, "environment:"+config.Environment)
	}

	// Create a slice of strings that are also ints except for 1.
	values := []string{"1", "2", "3ag", "4", "5"}
	var convValues []int

	// Notify the sidecar of initial value count.
	ddClient.Gauge("value.count", float64(len(values)), nil, 1)

	// Range over these values and try to convert the strings to ints.
	for _, v := range values {
		conv, err := strconv.Atoi(v)

		// Log and send an error to the sidecar when the converstion fails.
		if err != nil {
			log.Println(errors.Wrapf(err, "error converting value: %v", v))
			ddClient.Incr("conversion.error", nil, 1)
			continue
		}

		convValues = append(convValues, conv)
	}

	// Send the successful converted value count a task complete notification to the sidecar.
	ddClient.Gauge("converted.count", float64(len(convValues)), nil, 1)

	// Example of how you can use tags for metrics that you want to track
	// across different functions in code
	somethingNormal(ddClient)
	somethingSpecial(ddClient)

	ddClient.Incr("task.complete", nil, 1)

	time.Sleep(10000 * time.Millisecond)
	log.Println("Done!")
	return
}

func somethingSpecial(ddClient *statsd.Client) {
	log.Println("Doing something special")
	ddClient.Incr("task.happened", []string{"type:somethingspecial"}, 1)
	return
}

func somethingNormal(ddClient *statsd.Client) {
	log.Println("Doing something normal")
	ddClient.Incr("task.happened", []string{"type:somethingNormal"}, 1)
}
