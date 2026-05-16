package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func Pretty(v any) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}

func Debug(format string, args ...any) {
	if DEBUG {
		fmt.Printf("[DEBUG] "+format+"\n", args...)
	}
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	cfg := &Config{}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// skip empty lines or comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), `"`) // remove quotes

		switch key {
		case "base_url":
			cfg.BaseURL = value
		case "api_key":
			cfg.APIKey = value
		case "model_id":
			cfg.ModelID = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return cfg, nil
}
