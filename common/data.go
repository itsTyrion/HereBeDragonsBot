package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/disgoorg/snowflake/v2"
)

type BotState struct {
	LastNumber int          `json:"last_number"`
	ChannelID  snowflake.ID `json:"channel_id"`
	LastPerson snowflake.ID `json:"last_person"`
}

type BotConfig struct {
	Token               string       `json:"token"`
	MemberListChannelID snowflake.ID `json:"memberListChannelID"`
}

func InitData() {
	workDir := os.Getenv("WORK_DIR")
	if workDir == "" {
		workDir = "."
	}
	stateFile = filepath.Join(workDir, "state.json")
	configFile = filepath.Join(workDir, "config.json")
}

var stateFile string
var configFile string

func readFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		return nil, fmt.Errorf("read file %s: %w", path, err)
	}
	return data, nil
}

func LoadState() (*BotState, error) {
	data, err := readFile(stateFile)
	if err != nil {
		return nil, err
	}

	var state BotState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("decode state: %w", err)
	}

	return &state, nil
}

func SaveState(state BotState) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("encode state: %w", err)
	}

	return os.WriteFile(stateFile, data, 0o644)
}

func LoadConfig() BotConfig {
	data, err := readFile(configFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if err := SaveConfig(BotConfig{}); err != nil {
				slog.Error("failed to save default config", "error", err)
				os.Exit(1)
			}
			return BotConfig{}
		}
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	var config BotConfig
	if err := json.Unmarshal(data, &config); err != nil {
		slog.Error("failed to decode config", "error", err)
		os.Exit(1)
	}

	return config
}

func SaveConfig(config BotConfig) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}

	return os.WriteFile(configFile, data, 0o644)
}
