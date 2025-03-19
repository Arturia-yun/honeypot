package config

import (
    "gopkg.in/yaml.v2"
    "honeypot/server/pkg/vars"
    "os"
)

func LoadConfig(configPath string) error {
    data, err := os.ReadFile(configPath)
    if err != nil {
        return err
    }

    return yaml.Unmarshal(data, &vars.GlobalConfig)
}