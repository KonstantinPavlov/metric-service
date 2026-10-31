package main

import (
	"flag"
	"os"
	"strconv"

	"github.com/KonstantinPavlov/metric-service/internal/config"
)

type ServerConfig struct {
	RunAddr              string `json:"address" flag:"a"`
	StoreIntervalSeconds int    `json:"store_interval" flag:"i"`
	StorePath            string `json:"store_file" flag:"f"`
	Restore              bool   `json:"restore" flag:"r"`
	DBDSN                string `json:"database_dsn" flag:"d"`
	CryptoKey            string `json:"-" flag:"k"`
	AuditFile            string `json:"-" flag:"audit-file"`
	AuditURL             string `json:"-" flag:"audit-url"`
	CryptoKeyPath        string `json:"crypto_key" flag:"crypto-key"`
	ConfigPath           string `json:"-" flag:"config"`
}

var cfg ServerConfig

func parseFlags() error {
	fs := flag.CommandLine
	
	fs.StringVar(&cfg.RunAddr, "a", "localhost:8080", "address and port to run server")
	fs.IntVar(&cfg.StoreIntervalSeconds, "i", 300, "store interval on disk")
	fs.StringVar(&cfg.StorePath, "f", "", "path to storage file")
	fs.BoolVar(&cfg.Restore, "r", false, "restore data from storage on startup")
	fs.StringVar(&cfg.CryptoKey, "k", "", "SHA-256 key as 64-character hex string")
	fs.StringVar(&cfg.DBDSN, "d", "", "DSN for PostgresSQL")
	fs.StringVar(&cfg.AuditFile, "audit-file", "", "Path to audit file")
	fs.StringVar(&cfg.AuditURL, "audit-url", "", "Url for audit endpoint")
	fs.StringVar(&cfg.CryptoKeyPath, "crypto-key", "", "Private key path")
	fs.StringVar(&cfg.ConfigPath, "config", "", "path to configuration file")

	if err := config.LoadAndMerge(fs, cfg.ConfigPath, &cfg); err != nil {
		return err
	}

	cfg.RunAddr = config.EnvOrStr("ADDRESS", cfg.RunAddr)
	cfg.StorePath = config.EnvOrStr("FILE_STORAGE_PATH", cfg.StorePath)
	cfg.DBDSN = config.EnvOrStr("DATABASE_DSN", cfg.DBDSN)
	cfg.CryptoKey = config.EnvOrStr("KEY", cfg.CryptoKey)
	cfg.AuditFile = config.EnvOrStr("AUDIT_FILE", cfg.AuditFile)
	cfg.AuditURL = config.EnvOrStr("AUDIT_URL", cfg.AuditURL)
	cfg.CryptoKeyPath = config.EnvOrStr("CRYPTO_KEY", cfg.CryptoKeyPath)

	if storeInterval, err := config.ParseIntEnvVal("STORE_INTERVAL"); err == nil && storeInterval != nil {
		cfg.StoreIntervalSeconds = *storeInterval
	} else if err != nil {
		return err
	}

	if restoreStr := os.Getenv("RESTORE"); restoreStr != "" {
		restore, err := strconv.ParseBool(restoreStr)
		if err != nil {
			return err
		}
		cfg.Restore = restore
	}

	return nil
}
