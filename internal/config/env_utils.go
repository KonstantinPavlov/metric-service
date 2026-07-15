package config

import (
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
