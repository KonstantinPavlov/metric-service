package main

import (
	"flag"
	"os"
	"strconv"

	"github.com/KonstantinPavlov/metric-service/internal/config"
)

var flagRunAddr string
var flagStoreIntervalSeconds int
var flagStorePath string
var flagRestore bool
var flagDBDSN string
var flagCryptoKey string
var flagAuditFile string
var flagAuditURL string

func parseFlags() error {
	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&flagStoreIntervalSeconds, "i", 300, "store interval on disk - if 0 - sync store on disk")
	flag.StringVar(&flagStorePath, "f", "", "path to storage file")
	flag.BoolVar(&flagRestore, "r", false, "restore data from storage on startup")
	flag.StringVar(&flagDBDSN, "d", "", "DSN for PostgresSQL")
	flag.StringVar(&flagAuditFile, "audit-file", "", "Path to audit file")
	flag.StringVar(&flagAuditURL, "audit-url", "", "Url for audit endpoint")
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
		flagDBDSN = dsn
	}

	key := os.Getenv("KEY")
	if key != "" {
		flagCryptoKey = key
	}

	auditFile := os.Getenv("AUDIT_FILE")
	if auditFile != "" {
		flagAuditFile = auditFile
	}

	auditURL := os.Getenv("AUDIT_URL")
	if auditURL != "" {
		flagAuditURL = auditURL
	}

	return nil
}
