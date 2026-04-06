package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const DefaultBaseURL = "https://api.monologue.to"

type StoredConfig struct {
	BaseURL string `json:"base_url,omitempty"`
	Token   string `json:"token,omitempty"`
}

type Config struct {
	BaseURL string
	Token   string
	Path    string
}

func Load(baseURLFlag string, tokenFlag string) (Config, error) {
	path, err := Path()
	if err != nil {
		return Config{}, err
	}

	stored, err := read(path)
	if err != nil {
		return Config{}, err
	}

	return Config{
		BaseURL: firstNonEmpty(
			strings.TrimSpace(baseURLFlag),
			strings.TrimSpace(os.Getenv("MONOLOGUE_API_BASE_URL")),
			stored.BaseURL,
			DefaultBaseURL,
		),
		Token: firstNonEmpty(
			strings.TrimSpace(tokenFlag),
			strings.TrimSpace(os.Getenv("MONOLOGUE_API_TOKEN")),
			stored.Token,
		),
		Path: path,
	}, nil
}

func Save(stored StoredConfig) (string, error) {
	path, err := Path()
	if err != nil {
		return "", err
	}

	stored.BaseURL = firstNonEmpty(strings.TrimSpace(stored.BaseURL), DefaultBaseURL)
	stored.Token = strings.TrimSpace(stored.Token)
	if stored.Token == "" {
		return "", errors.New("token is required")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return "", err
	}

	payload, err := json.MarshalIndent(stored, "", "  ")
	if err != nil {
		return "", err
	}
	payload = append(payload, '\n')

	tempPath := path + ".tmp"
	if err := os.WriteFile(tempPath, payload, 0o600); err != nil {
		return "", err
	}
	if err := os.Rename(tempPath, path); err != nil {
		return "", err
	}

	return path, nil
}

func Path() (string, error) {
	if override := strings.TrimSpace(os.Getenv("MONOLOGUE_CONFIG_DIR")); override != "" {
		return filepath.Join(override, "config.json"), nil
	}

	if xdgConfigHome := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME")); xdgConfigHome != "" {
		return filepath.Join(xdgConfigHome, "monologue", "config.json"), nil
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "monologue", "config.json"), nil
}

func read(path string) (StoredConfig, error) {
	payload, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return StoredConfig{}, nil
	}
	if err != nil {
		return StoredConfig{}, err
	}

	var stored StoredConfig
	if err := json.Unmarshal(payload, &stored); err != nil {
		return StoredConfig{}, err
	}

	stored.BaseURL = strings.TrimSpace(stored.BaseURL)
	stored.Token = strings.TrimSpace(stored.Token)

	return stored, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}

	return ""
}
