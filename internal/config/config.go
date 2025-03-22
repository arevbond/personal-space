package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server Server `yaml:"server"`
}

type Server struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

func New(configPath string) (Config, error) {
	if _, err := os.Stat(configPath); err != nil {
		return Config{}, fmt.Errorf("file '%s' : %w", configPath, err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, fmt.Errorf("can't read file %s: %w", configPath, err)
	}

	var config Config

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return Config{}, fmt.Errorf("can't unmarshall data to config: %w", err)
	}

	return config, nil
}
