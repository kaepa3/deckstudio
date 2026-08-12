package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	Port int `yaml:"port"`
}

type ButtonConfig struct {
	ID    string   `yaml:"id" json:"id"`
	Label string   `yaml:"label" json:"label"`
	Icon  string   `yaml:"icon" json:"icon"`
	Color string   `yaml:"color" json:"color"`
	Type  string   `yaml:"type" json:"type"` // "key" or "shortcut"
	Keys  []string `yaml:"keys" json:"keys"`
}

type ProfileConfig struct {
	ID       string         `yaml:"id" json:"id"`
	Name     string         `yaml:"name" json:"name"`
	AppNames []string       `yaml:"app_names" json:"app_names"`
	Icon     string         `yaml:"icon" json:"icon"`
	Buttons  []ButtonConfig `yaml:"buttons" json:"buttons"`
}

type Config struct {
	Server        ServerConfig    `yaml:"server" json:"server"`
	GlobalButtons []ButtonConfig  `yaml:"global_buttons" json:"global_buttons"`
	Profiles      []ProfileConfig `yaml:"profiles" json:"profiles"`
	LoadedPath    string          `yaml:"-" json:"-"`
}

const ConfigFileName = ".deckstudio.yaml"

func LoadConfig() (*Config, error) {
	var candidates []string

	// 1. Current working directory
	candidates = append(candidates, ConfigFileName)

	// 2. Executable directory
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		candidates = append(candidates, filepath.Join(exeDir, ConfigFileName))
	}

	// 3. User Home Directory (C:\Users\<Username>\.deckstudio.yaml)
	if homeDir, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates, filepath.Join(homeDir, ConfigFileName))
		candidates = append(candidates, filepath.Join(homeDir, ".deckstudio", ConfigFileName))
	}

	var foundPath string
	var data []byte
	var err error

	for _, path := range candidates {
		if _, statErr := os.Stat(path); statErr == nil {
			data, err = os.ReadFile(path)
			if err == nil {
				foundPath = path
				break
			}
		}
	}

	if foundPath == "" {
		return nil, fmt.Errorf("%s not found in candidate locations: %v", ConfigFileName, candidates)
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", foundPath, err)
	}

	cfg.LoadedPath = foundPath
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}

	log.Printf("📄 Loaded configuration from: %s", foundPath)
	return &cfg, nil
}

func (c *Config) FindButtonByID(id string) *ButtonConfig {
	for _, btn := range c.GlobalButtons {
		if btn.ID == id {
			return &btn
		}
	}
	for _, p := range c.Profiles {
		for _, btn := range p.Buttons {
			if btn.ID == id {
				return &btn
			}
		}
	}
	return nil
}
