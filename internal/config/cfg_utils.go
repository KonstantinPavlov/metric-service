package config

import (
	"encoding/json"
	"os"
	"strconv"
)

func ParseIntEnvVal(env string) (*int, error) {
	valueStr := os.Getenv(env)
	if valueStr != "" {
		valueInt, err := strconv.Atoi(valueStr)
		if err != nil {
			return nil, err
		}
		return &valueInt, nil
	}
	return nil, nil
}

func LoadConfigFromFile[T any](path string) (*T, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cfg T
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
