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
var flagCryptoKeyPath string
var flagConfigPath string

type ServerConfig struct {
	RunAddr              string `json:"address"`
	Restore              *bool  `json:"restore"`
	StoreIntervalSeconds int    `json:"store_interval"`
	StorePath            string `json:"store_file"`
	DBDSN                string `json:"database_dsn"`
	CryptoKeyPath        string `json:"crypto_key"`
}

func parseFlags() error {
	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&flagStoreIntervalSeconds, "i", 300, "store interval on disk - if 0 - sync store on disk")
	flag.StringVar(&flagStorePath, "f", "", "path to storage file")
	flag.BoolVar(&flagRestore, "r", false, "restore data from storage on startup")
	flag.StringVar(&flagCryptoKey, "k", "", "SHA-256 key as 64-character hex string")
	flag.StringVar(&flagDBDSN, "d", "", "DSN for PostgresSQL")
	flag.StringVar(&flagAuditFile, "audit-file", "", "Path to audit file")
	flag.StringVar(&flagAuditURL, "audit-url", "", "Url for audit endpoint")
	flag.StringVar(&flagCryptoKeyPath, "crypto-key", "", "Private key path")
	flag.StringVar(&flagConfigPath, "config", "", "path to configuration file")
	flag.Parse()

	configPath := flagConfigPath
	envConfigPath := os.Getenv("CONFIG")

	if envConfigPath != "" {
		configPath = envConfigPath
	}

	if configPath != "" {
		fileConf, err := config.LoadConfigFromFile[ServerConfig](configPath)
		if err != nil {
			return err
		}

		actualFlags := make(map[string]bool)
		flag.Visit(func(f *flag.Flag) {
			actualFlags[f.Name] = true
		})

		if !actualFlags["a"] && fileConf.RunAddr != "" {
			flagRunAddr = fileConf.RunAddr
		}
		if !actualFlags["i"] && fileConf.StoreIntervalSeconds != 0 {
			flagStoreIntervalSeconds = fileConf.StoreIntervalSeconds
		}
		if !actualFlags["f"] && fileConf.StorePath != "" {
			flagStorePath = fileConf.StorePath
		}
		if !actualFlags["r"] && fileConf.Restore != nil {
			flagRestore = *fileConf.Restore
		}
		if !actualFlags["d"] && fileConf.DBDSN != "" {
			flagDBDSN = fileConf.DBDSN
		}
		if !actualFlags["crypto-key"] && fileConf.CryptoKeyPath != "" {
			flagCryptoKeyPath = fileConf.CryptoKeyPath
		}
	}

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

	cryptoKeyPath := os.Getenv("CRYPTO_KEY")
	if cryptoKeyPath != "" {
		flagCryptoKeyPath = cryptoKeyPath
	}

	return nil
}
