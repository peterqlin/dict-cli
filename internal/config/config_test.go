package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	// Test defaults using a non-existent temp file path (no config file on disk,
	// no env vars) so only Viper defaults apply.
	os.Unsetenv("MWEB_API_KEY_DICT")
	os.Unsetenv("MWEB_API_KEY_THESAURUS")
	os.Unsetenv("MWEB_OUTPUT_FORMAT")
	os.Unsetenv("MWEB_MAX_DEFINITIONS")

	cfg, err := loadFromFile("/tmp/mweb_nonexistent_test_config.yaml")
	if err != nil {
		t.Fatalf("loadFromFile() error: %v", err)
	}
	if cfg.OutputFormat != "plain" {
		t.Errorf("default OutputFormat = %q, want \"plain\"", cfg.OutputFormat)
	}
	if cfg.MaxDefinitions != 5 {
		t.Errorf("default MaxDefinitions = %d, want 5", cfg.MaxDefinitions)
	}
	// API keys should be empty when neither file nor env vars are present.
	if cfg.APIKeyDict != "" {
		t.Errorf("expected empty APIKeyDict, got %q", cfg.APIKeyDict)
	}
}

func TestLoad_EnvVarOverrides(t *testing.T) {
	os.Setenv("MWEB_API_KEY_DICT", "env-dict-key")
	os.Setenv("MWEB_API_KEY_THESAURUS", "env-thes-key")
	os.Setenv("MWEB_OUTPUT_FORMAT", "json")
	os.Setenv("MWEB_MAX_DEFINITIONS", "10")
	defer func() {
		os.Unsetenv("MWEB_API_KEY_DICT")
		os.Unsetenv("MWEB_API_KEY_THESAURUS")
		os.Unsetenv("MWEB_OUTPUT_FORMAT")
		os.Unsetenv("MWEB_MAX_DEFINITIONS")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.APIKeyDict != "env-dict-key" {
		t.Errorf("APIKeyDict = %q, want \"env-dict-key\"", cfg.APIKeyDict)
	}
	if cfg.APIKeyThesaurus != "env-thes-key" {
		t.Errorf("APIKeyThesaurus = %q, want \"env-thes-key\"", cfg.APIKeyThesaurus)
	}
	if cfg.OutputFormat != "json" {
		t.Errorf("OutputFormat = %q, want \"json\"", cfg.OutputFormat)
	}
	if cfg.MaxDefinitions != 10 {
		t.Errorf("MaxDefinitions = %d, want 10", cfg.MaxDefinitions)
	}
}

func TestLoad_ConfigFile(t *testing.T) {
	// Write a temp config file and point configFilePath at it via a custom load.
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "config.yaml")
	content := `api_key_dict: file-dict-key
api_key_thesaurus: file-thes-key
output_format: json
max_definitions: 3
`
	if err := os.WriteFile(cfgFile, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	// Use loadFromFile to test config-file parsing directly.
	cfg, err := loadFromFile(cfgFile)
	if err != nil {
		t.Fatalf("loadFromFile() error: %v", err)
	}
	if cfg.APIKeyDict != "file-dict-key" {
		t.Errorf("APIKeyDict = %q, want \"file-dict-key\"", cfg.APIKeyDict)
	}
	if cfg.APIKeyThesaurus != "file-thes-key" {
		t.Errorf("APIKeyThesaurus = %q, want \"file-thes-key\"", cfg.APIKeyThesaurus)
	}
	if cfg.OutputFormat != "json" {
		t.Errorf("OutputFormat = %q, want \"json\"", cfg.OutputFormat)
	}
	if cfg.MaxDefinitions != 3 {
		t.Errorf("MaxDefinitions = %d, want 3", cfg.MaxDefinitions)
	}
}

func TestLoad_MissingConfigFile_NotError(t *testing.T) {
	os.Unsetenv("MWEB_API_KEY_DICT")
	// A missing config file should not be an error.
	cfg, err := Load()
	if err != nil {
		t.Errorf("missing config file should not be an error, got: %v", err)
	}
	if cfg == nil {
		t.Error("expected non-nil config even with missing file")
	}
}
