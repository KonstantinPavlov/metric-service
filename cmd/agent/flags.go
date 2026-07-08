package main

import (
	"flag"
	"os"
	"strconv"
)

var flagServerAddr string
var flagReportInterval int
var flagPollInterval int

func parseFlags() error {
	flag.StringVar(&flagServerAddr, "a", "localhost:8080", "address and port metric server")
	flag.IntVar(&flagReportInterval, "r", 10, "report interval in seconds")
	flag.IntVar(&flagPollInterval, "p", 2, "poll interval in seconds")
	flag.Parse()

	address := os.Getenv("ADDRESS")
	if address != "" {
		flagServerAddr = address
	}
	reportInterval, err := parseIntEnvVal("REPORT_INTERVAL")
	if err != nil {
		return err
	}
	if reportInterval != nil {
		flagReportInterval = *reportInterval
	}
	pollInterval, err := parseIntEnvVal("POLL_INTERVAL")
	if err != nil {
		return err
	}
	if pollInterval != nil {
		flagPollInterval = *pollInterval
	}

	return nil
}

func parseIntEnvVal(env string) (*int, error) {
	valueStr := os.Getenv(env)
	if valueStr != "" {
		intVal, error := strconv.Atoi(valueStr)
		if error != nil {
			return nil, error
		}
		return &intVal, nil
	}
	return nil, nil
}
