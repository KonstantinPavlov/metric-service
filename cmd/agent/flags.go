package main

import (
	"flag"
	"github.com/KonstantinPavlov/metric-service/internal/config"
)

type AgentConfig struct {
	ServerAddr      string                 `json:"address" flag:"a"`
	ReportInterval  config.DurationSeconds `json:"report_interval" flag:"r"`
	PollInterval    config.DurationSeconds `json:"poll_interval" flag:"p"`
	CryptoKey       string                 `json:"-" flag:"k"`
	RateLimit       int                    `json:"-" flag:"l"`
	PublicCryptoKey string                 `json:"crypto_key" flag:"crypto-key"`
	ConfigPath      string                 `json:"-" flag:"config"`
}

var cfg AgentConfig

func parseFlags() error {
	fs := flag.CommandLine
	fs.StringVar(&cfg.ServerAddr, "a", "localhost:8080", "address and port metric server")
	fs.IntVar((*int)(&cfg.ReportInterval), "r", 10, "report interval in seconds")
	fs.IntVar((*int)(&cfg.PollInterval), "p", 2, "poll interval in seconds")
	fs.StringVar(&cfg.CryptoKey, "k", "", "SHA-256 key as 64-character hex string")
	fs.IntVar(&cfg.RateLimit, "l", 1, "requests to server rate limit")
	fs.StringVar(&cfg.PublicCryptoKey, "crypto-key", "", "Public key path")
	fs.StringVar(&cfg.ConfigPath, "config", "", "path to configuration file")

	if err := config.LoadAndMerge(fs, cfg.ConfigPath, &cfg); err != nil {
		return err
	}

	cfg.ServerAddr = config.EnvOrStr("ADDRESS", cfg.ServerAddr)
	cfg.CryptoKey = config.EnvOrStr("KEY", cfg.CryptoKey)
	cfg.PublicCryptoKey = config.EnvOrStr("CRYPTO_KEY", cfg.PublicCryptoKey)

	if reportInterval, err := config.ParseIntEnvVal("REPORT_INTERVAL"); err == nil && reportInterval != nil {
		cfg.ReportInterval = config.DurationSeconds(*reportInterval)
	} else if err != nil {
		return err
	}

	if pollInterval, err := config.ParseIntEnvVal("POLL_INTERVAL"); err == nil && pollInterval != nil {
		cfg.PollInterval = config.DurationSeconds(*pollInterval)
	} else if err != nil {
		return err
	}

	if rateLimit, err := config.ParseIntEnvVal("RATE_LIMIT"); err == nil && rateLimit != nil {
		cfg.RateLimit = *rateLimit
	} else if err != nil {
		return err
	}

	return nil
}
