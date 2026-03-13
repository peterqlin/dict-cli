package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all mweb configuration values.
type Config struct {
	APIKeyDict      string `mapstructure:"api_key_dict"`
	APIKeyThesaurus string `mapstructure:"api_key_thesaurus"`
	OutputFormat    string `mapstructure:"output_format"`
	MaxDefinitions  int    `mapstructure:"max_definitions"`
}

// Load reads ~/.config/mweb/config.yaml and applies env var overrides.
// Missing config file is not an error — API keys may come from env vars alone.
func Load() (*Config, error) {
	cfgPath, err := configFilePath()
	if err != nil {
		return nil, err
	}
	return loadFromFile(cfgPath)
}

// loadFromFile reads config from the given YAML file path and applies env var
// overrides. A missing file is silently ignored.
func loadFromFile(cfgPath string) (*Config, error) {
	v := viper.New()

	// Defaults
	v.SetDefault("output_format", "plain")
	v.SetDefault("max_definitions", 5)

	v.SetConfigFile(cfgPath)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		// Accept "file not found" silently; anything else is a real error.
		if _, statErr := os.Stat(cfgPath); statErr == nil {
			return nil, fmt.Errorf("reading config file %s: %w", cfgPath, err)
		}
	}

	// Env var overrides: MWEB_API_KEY_DICT, MWEB_API_KEY_THESAURUS, etc.
	v.SetEnvPrefix("MWEB")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	return &cfg, nil
}

// configFilePath returns the path to ~/.config/mweb/config.yaml.
func configFilePath() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("finding user config directory: %w", err)
	}
	return filepath.Join(base, "mweb", "config.yaml"), nil
}
