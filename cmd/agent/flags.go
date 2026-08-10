package main

import (
	
	"flag"
	"github.com/KonstantinPavlov/metric-service/internal/config"
	"os"
)

var flagServerAddr string
var flagReportInterval int
var flagPollInterval int
var flagCryptoKey string

func parseFlags() error {
	flag.StringVar(&flagServerAddr, "a", "localhost:8080", "address and port metric server")
	flag.IntVar(&flagReportInterval, "r", 10, "report interval in seconds")
	flag.IntVar(&flagPollInterval, "p", 2, "poll interval in seconds")
	flag.StringVar(&flagCryptoKey, "k", "", "SHA-256 key as 64-character hex string")
	flag.Parse()

	address := os.Getenv("ADDRESS")
	if address != "" {
		flagServerAddr = address
	}
	reportInterval, err := config.ParseIntEnvVal("REPORT_INTERVAL")
	if err != nil {
		return err
	}
	if reportInterval != nil {
		flagReportInterval = *reportInterval
	}
	pollInterval, err := config.ParseIntEnvVal("POLL_INTERVAL")
	if err != nil {
		return err
	}
	if pollInterval != nil {
		flagPollInterval = *pollInterval
	}
	
	key := os.Getenv("KEY")
	if key != "" {		
		flagCryptoKey = key
	}

	return nil
}
