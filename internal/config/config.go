package config

import (
	"fmt"
	"os"

	"github.com/stretchr/testify/assert/yaml"
)


type Config struct {
	Database DatabaseConfig `yaml:"database"`
	Server ServerConfig `yaml:"server"`
	Storage StorageConfig `yaml:"storage"`
}


type DatabaseConfig struct {
    Host     string `yaml:"host"`
    Port     int    `yaml:"port"`
    User     string `yaml:"user"`
    Password string `yaml:"password"`
    Name     string `yaml:"name"`
}

type ServerConfig struct {
    Port int `yaml:"port"`
}

type StorageConfig struct {
    UploadPath string `yaml:"upload_path"`
}

func Load(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("reading config: %w", err)
    }

	var cfg Config
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("parsing config: %w", err)
    }

    // Allow env var overrides for Docker
    if host := os.Getenv("DB_HOST"); host != "" {
        cfg.Database.Host = host
    }

    return &cfg, nil
}

func (c *DatabaseConfig) DSN() string {
    return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
        c.User, c.Password, c.Host, c.Port, c.Name)
}