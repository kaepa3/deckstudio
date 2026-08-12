package main

import (
	"os"

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
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}

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
