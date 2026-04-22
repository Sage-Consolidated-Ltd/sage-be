package aws

import (
	"encoding/json"
	"errors"
)

type Config struct {
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
	Region    string `json:"region"`
}

func ParseConfig(raw []byte) (Config, error) {
	var cfg Config
	err := json.Unmarshal(raw, &cfg)
	return cfg, err
}

func Validate(cfg Config) error {
	if cfg.AccessKey == "" {
		return errors.New("missing access_key")
	}
	if cfg.SecretKey == "" {
		return errors.New("missing secret_key")
	}
	if cfg.Region == "" {
		return errors.New("missing region")
	}
	return nil
}

