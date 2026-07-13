package main

import (
	"flag"
	"os"
	"strconv"
)

var flagRunAddr string
var flagStoreIntervalSeconds int
var flagStorePath string
var flagRestore bool

func parseFlags() error {
	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&flagStoreIntervalSeconds, "i", 300, "store interval on disk - if 0 - sync store on disk")
	flag.StringVar(&flagStorePath, "f", "./storage", "path to storage file")
	flag.BoolVar(&flagRestore, "r", false, "restore data from storage on startup")
	flag.Parse()

	address := os.Getenv("ADDRESS")
	if address != "" {
		flagRunAddr = address
	}

	storeInterval, err := parseIntEnvVal("STORE_INTERVAL")
	if err != nil {
		return err
	}
	if storeInterval != nil {
		flagStoreIntervalSeconds = *storeInterval
	}
	path := os.Getenv("FILE_STORAGE_PATH")
	if path != "" {
		flagStorePath = path
	}

	restoreStr := os.Getenv("RESTORE")
	if restoreStr != "" {
		restore, err := strconv.ParseBool(restoreStr)
		if err != nil {
			return err
		}
		flagRestore = restore
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
