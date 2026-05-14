package config

import (
	"fmt"
	"os"
	"regexp"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Database DatabaseConfig `yaml:"database"`
	JWT      JWTConfig      `yaml:"jwt"`
	Server   ServerConfig   `yaml:"server"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	Schema   string `yaml:"schema"`
}

type JWTConfig struct {
	Secret     string `yaml:"secret"`
	ExpiryMins int    `yaml:"expiry_mins"`
}

type ServerConfig struct {
	Port string `yaml:"port"`
	Env  string `yaml:"env"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	// Expand environment variables in YAML content
	expanded := expandEnvVars(string(data))

	var cfg Config
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	// Override with environment variables (higher priority)
	cfg.applyEnvOverrides()

	return &cfg, nil
}

// expandEnvVars expands ${VAR:-default} syntax in YAML content
func expandEnvVars(content string) string {
	re := regexp.MustCompile(`\$\{([^}:]+)(?::-([^}]*))?\}`)
	return re.ReplaceAllStringFunc(content, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) < 2 {
			return match
		}
		varName := parts[1]
		defaultVal := ""
		if len(parts) >= 3 {
			defaultVal = parts[2]
		}
		if val := os.Getenv(varName); val != "" {
			return val
		}
		return defaultVal
	})
}

func (c *Config) applyEnvOverrides() {
	if host := os.Getenv("DB_HOST"); host != "" {
		c.Database.Host = host
	}
	if port := os.Getenv("DB_PORT"); port != "" {
		c.Database.Port = port
	}
	if user := os.Getenv("DB_USER"); user != "" {
		c.Database.User = user
	}
	if password := os.Getenv("DB_PASSWORD"); password != "" {
		c.Database.Password = password
	}
	if database := os.Getenv("DB_DATABASE"); database != "" {
		c.Database.Database = database
	}
	if schema := os.Getenv("DB_SCHEMA"); schema != "" {
		c.Database.Schema = schema
	}
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		c.JWT.Secret = secret
	}
	if expiry := os.Getenv("JWT_EXPIRY_MINS"); expiry != "" {
		if exp, err := strconv.Atoi(expiry); err == nil {
			c.JWT.ExpiryMins = exp
		}
	}
	if port := os.Getenv("SERVER_PORT"); port != "" {
		c.Server.Port = port
	}
	if env := os.Getenv("SERVER_ENV"); env != "" {
		c.Server.Env = env
	}
}

func (c *DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?search_path=%s",
		c.User, c.Password, c.Host, c.Port, c.Database, c.Schema,
	)
}
