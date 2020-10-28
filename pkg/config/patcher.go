package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type PatcherConfig struct {
	Patchers []FieldPatcher `json:"patchers"`
}

type FieldPatcher struct {
	Source  HTTPActionSpec `json:"source"`
	Setters []FieldSetter  `json:"setters"`
}

type FieldSetter struct {
	Parser string         `json:"parser"`
	Type   string         `json:"type"`
	Target HTTPActionSpec `json:"target"`
}

type HTTPActionSpec struct {
	Action      string            `json:"action"`
	URL         string            `json:"url"`
	ValidateTLS bool              `json:"validateTLS,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
}

func ParsePatcherConfig(filePath string) (*PatcherConfig, error) {
	jsonFile, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("can not open file: %v", err)
	}
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, fmt.Errorf("can not read file: %v", err)
	}

	var cfg PatcherConfig
	if err := json.Unmarshal(byteValue, &cfg); err != nil {
		return nil, fmt.Errorf("invalid json file: %v", err)
	}
	return &cfg, nil
}
