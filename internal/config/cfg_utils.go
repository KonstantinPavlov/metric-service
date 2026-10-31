package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"time"
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

func LoadAndMerge[T any](fs *flag.FlagSet, configPath string, cfgStruct *T) error {
	// 1. Flag parsing
	if !fs.Parsed() {
		if fs == flag.CommandLine {
			flag.Parse()
		} else {
			if err := fs.Parse(os.Args[1:]); err != nil {
				return err
			}
		}
	}

	// 2. Path to config-file
	if envPath := os.Getenv("CONFIG"); envPath != "" {
		configPath = envPath
	}

	// no config file - skipping
	if configPath == "" {
		return nil
	}

	// 3. Reading json to struct
	fileConf, err := LoadConfigFromFile[T](configPath)
	if err != nil {
		return fmt.Errorf("failed to load config file: %w", err)
	}

	// 4. actula flags
	actualFlags := make(map[string]bool)
	fs.Visit(func(f *flag.Flag) {
		actualFlags[f.Name] = true
	})

	// 5. Trasfering values from fileConf to cfgStruct
	valTarget := reflect.ValueOf(cfgStruct).Elem()
	valSource := reflect.ValueOf(fileConf).Elem() 

	for i := 0; i < valTarget.NumField(); i++ {
		field := valTarget.Type().Field(i)
		flagName := field.Tag.Get("flag")

		if flagName == "" {
			continue 
		}

		// no flag- taking from file 
		if !actualFlags[flagName] {
			targetField := valTarget.Field(i)
			sourceField := valSource.Field(i)

			if !sourceField.IsZero() {
				if targetField.Kind() == reflect.Ptr && sourceField.Kind() != reflect.Ptr {					
					ptr := reflect.New(sourceField.Type())
					ptr.Elem().Set(sourceField)
					targetField.Set(ptr)
				} else if targetField.Kind() != reflect.Ptr && sourceField.Kind() == reflect.Ptr {
					if !sourceField.IsNil() {
						targetField.Set(sourceField.Elem())
					}
				} else {					
					targetField.Set(sourceField)
				}
			}
		}
	}

	return nil
}

func EnvOrStr(envName string, currentVal string) string {
	if val := os.Getenv(envName); val != "" {
		return val
	}
	return currentVal
}


type DurationSeconds int

func (d *DurationSeconds) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {		
		var i int
		if err := json.Unmarshal(b, &i); err != nil {
			return err
		}
		*d = DurationSeconds(i)
		return nil
	}
	dur, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = DurationSeconds(dur.Seconds())
	return nil
}