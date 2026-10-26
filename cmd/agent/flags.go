package main

import (
	"flag"
	"os"

	"github.com/KonstantinPavlov/metric-service/internal/config"
)

var flagServerAddr string
var flagReportInterval int
var flagPollInterval int
var flagCryptoKey string
var flagRateLimit int
var flagPublicCryptoKeyPath string
var flagConfigPath string

type AgentConfig struct {
	ServerAddr      string `json:"address"`
	ReportInterval  int    `json:"report_interval"`
	PollInterval    int    `json:"poll_interval"`
	PublicCryptoKey string `json:"crypto_key"`
}

func parseFlags() error {
	flag.StringVar(&flagServerAddr, "a", "localhost:8080", "address and port metric server")
	flag.IntVar(&flagReportInterval, "r", 10, "report interval in seconds")
	flag.IntVar(&flagPollInterval, "p", 2, "poll interval in seconds")
	flag.StringVar(&flagCryptoKey, "k", "", "SHA-256 key as 64-character hex string")
	flag.IntVar(&flagRateLimit, "l", 1, "requests to server rate limit")
	flag.StringVar(&flagPublicCryptoKeyPath, "crypto-key", "", "Public key path")
	flag.StringVar(&flagConfigPath, "config", "", "path to configuration file")
	flag.Parse()

	configPath := flagConfigPath
	envConfigPath := os.Getenv("CONFIG")
	if envConfigPath != "" {
		configPath = envConfigPath
	}

	if configPath != "" {
		fileConf, err := config.LoadConfigFromFile[AgentConfig](configPath)
		if err != nil {
			return err
		}
		actualFlags := make(map[string]bool)
		flag.Visit(func(f *flag.Flag) {
			actualFlags[f.Name] = true
		})

		if !actualFlags["a"] && fileConf.ServerAddr != "" {
			flagServerAddr = fileConf.ServerAddr
		}
		if !actualFlags["r"] && fileConf.ReportInterval != 0 {
			flagReportInterval = fileConf.ReportInterval
		}
		if !actualFlags["p"] && fileConf.PollInterval != 0 {
			flagPollInterval = fileConf.PollInterval
		}
		if !actualFlags["k"] && fileConf.PublicCryptoKey != "" {
			flagCryptoKey = fileConf.PublicCryptoKey
		}

		if !actualFlags["crypto-key"] && fileConf.PublicCryptoKey != "" {
			flagPublicCryptoKeyPath = fileConf.PublicCryptoKey
		}
	}

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

	rateLimit, err := config.ParseIntEnvVal("RATE_LIMIT")
	if err != nil {
		return err
	}
	if rateLimit != nil {
		flagRateLimit = *rateLimit
	}

	cryptoKeyPath := os.Getenv("CRYPTO_KEY")
	if cryptoKeyPath != "" {
		flagPublicCryptoKeyPath = cryptoKeyPath
	}

	return nil
}
