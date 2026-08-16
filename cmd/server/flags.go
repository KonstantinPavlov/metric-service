package main

import (
	"flag"
	"github.com/KonstantinPavlov/metric-service/internal/config"
	"os"
	"strconv"
)

var flagRunAddr string
var flagStoreIntervalSeconds int
var flagStorePath string
var flagRestore bool
var flagDbDSN string
var flagCryptoKey string

func parseFlags() error {
	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&flagStoreIntervalSeconds, "i", 300, "store interval on disk - if 0 - sync store on disk")
	flag.StringVar(&flagStorePath, "f", "", "path to storage file")
	flag.BoolVar(&flagRestore, "r", false, "restore data from storage on startup")
	flag.StringVar(&flagDbDSN, "d", "", "DSN for PostgresSQL")
	flag.StringVar(&flagCryptoKey, "k", "", "SHA-256 key as 64-character hex string")
	flag.Parse()

	address := os.Getenv("ADDRESS")
	if address != "" {
		flagRunAddr = address
	}

	storeInterval, err := config.ParseIntEnvVal("STORE_INTERVAL")
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

	dsn := os.Getenv("DATABASE_DSN")
	if dsn != "" {
		flagDbDSN = dsn
	}

	key := os.Getenv("KEY")
	if key != "" {
		flagCryptoKey = key
	}

	return nil
}
