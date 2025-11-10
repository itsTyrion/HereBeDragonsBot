package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/disgoorg/snowflake/v2"
)

type botState struct {
	LastNumber int    `json:"last_number"`
	ChannelID  string `json:"channel_id"`
	LastPerson string `json:"last_person"`
}

type botConfig struct {
	Token string `json:"token"`
}

func initData() {
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

func loadState() error {
	data, err := readFile(stateFile)
	if err != nil {
		return err
	}

	var state botState
	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("decode state: %w", err)
	}

	lastNumber = state.LastNumber

	if id, err := snowflake.Parse(state.ChannelID); err != nil {
		return fmt.Errorf("parse channel id: %w", err)
	} else {
		channelID = id
	}

	if id, err := snowflake.Parse(state.LastPerson); err != nil {
		return fmt.Errorf("parse lastPerson id: %w", err)
	} else {
		lastPerson = id
	}

	return nil
}

func saveState() error {
	state := botState{
		LastNumber: lastNumber,
		ChannelID:  channelID.String(),
		LastPerson: lastPerson.String(),
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("encode state: %w", err)
	}

	return os.WriteFile(stateFile, data, 0o644)
}

func loadConfig() botConfig {
	data, err := readFile(configFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			saveConfig(botConfig{})
			return botConfig{}
		}
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	var config botConfig
	if err := json.Unmarshal(data, &config); err != nil {
		slog.Error("failed to decode config", "error", err)
		os.Exit(1)
	}

	return config
}

func saveConfig(config botConfig) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}

	return os.WriteFile(configFile, data, 0o644)
}
