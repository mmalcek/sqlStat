package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

type tConfig struct {
	Server         string `yaml:"server"`
	Instance       string `yaml:"instance"`
	Database       string `yaml:"database"`
	Port           int    `yaml:"port"`
	User           string `yaml:"user"`
	Password       string `yaml:"password"`
	OutputPath     string `yaml:"outputPath"`
	OutputDateMark bool   `yaml:"outputDateMark"`
}

func getConfig() error {
	file, err := os.ReadFile("config.yaml")
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(file, &config); err != nil {
		return err
	}
	return nil
}
