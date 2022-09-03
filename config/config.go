package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type NexpConfig struct {
	Token  string
	Images ImageConfig
}

type ImageConfig struct {
	SavePath          string
	IgnoreImages      bool
	OverwriteExisting bool
}

func LoadNexpConfig() (*NexpConfig, error) {
	dir, err := ResolveConfigDirectory()
	if err != nil {
		return nil, fmt.Errorf("failed resolving home directory, "+
			"error: %s\n", err)
	}

	c, err := os.ReadFile(dir)
	if err != nil {
		return nil, fmt.Errorf("failed loading configuraiton file, "+
			"error: %s\n", err)
	}
	config := NexpConfig{}
	err = yaml.Unmarshal(c, &config)
	if err != nil {
		return nil, fmt.Errorf("failed loading configuraiton file, "+
			"error: %s\n", err)
	}

	return &config, nil
}

func SaveNexpConfig(c NexpConfig) error {
	dir, err := ResolveConfigDirectory()
	if err != nil {
		return fmt.Errorf("failed resolving home directory, "+
			"error: %s\n", err)
	}

	yConf, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("Failed marshalling config into bytes "+
			"error: %s\n", err)
	}
	err = os.WriteFile(dir, yConf, 0666)
	if err != nil {
		return fmt.Errorf("Failed to write config file "+
			"error: %s\n", err)
	}

	return nil
}

func ResolveConfigDirectory() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "nexp.yaml"), nil
}
