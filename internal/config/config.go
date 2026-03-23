package config

import (
	"os"

	"gopkg.in/yaml.v3"

	"scenario2test/internal/generator/e2e"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
	E2E    e2e.Config   `yaml:"e2e"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

func Default() Config {
	return Config{
		Server: ServerConfig{Port: 8080},
		E2E: e2e.Config{
			Provider: "AUTOTEST",
			Mode:     e2e.ModeMock,
			Timeout:  60,
		},
	}
}

func Load(path string) (Config, error) {
	cfg := Default()
	if path == "" {
		return cfg, nil
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return Config{}, err
	}

	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return Config{}, err
	}

	cfg.applyDefaults()
	return cfg, nil
}

func (c *Config) applyDefaults() {
	defaults := Default()
	if c.Server.Port == 0 {
		c.Server.Port = defaults.Server.Port
	}
	if c.E2E.Provider == "" {
		c.E2E.Provider = defaults.E2E.Provider
	}
	if c.E2E.Mode == "" {
		c.E2E.Mode = defaults.E2E.Mode
	}
	if c.E2E.Timeout == 0 {
		c.E2E.Timeout = defaults.E2E.Timeout
	}
}
