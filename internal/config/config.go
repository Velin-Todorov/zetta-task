package config

import (
	"fmt"
	"os"

	"github.com/stretchr/testify/assert/yaml"
)


// Config holds the config for DB, server and storage
type Config struct {
	Database DatabaseConfig `yaml:"database"`
	Server ServerConfig `yaml:"server"`
	Storage StorageConfig `yaml:"storage"`
}

// Database config is the config for the DB
type DatabaseConfig struct {
    Host     string `yaml:"host"`
    Port     int    `yaml:"port"`
    User     string `yaml:"user"`
    Password string `yaml:"password"`
    Name     string `yaml:"name"`
}

// ServerConfig is the config for the server
type ServerConfig struct {
    Port int `yaml:"port"`
}

// StorageConfig is the config for the storage
type StorageConfig struct {
    UploadPath string `yaml:"upload_path"`
}

// Load loads the data from yaml config and creates the config
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

// Create a connection string for MySQL
func (c *DatabaseConfig) DSN() string {
    return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
        c.User, c.Password, c.Host, c.Port, c.Name)
}