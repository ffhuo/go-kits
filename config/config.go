package config

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// Config represents a configuration loader
type Config struct {
	v *viper.Viper
}

// New creates a new Config instance
func New() *Config {
	return &Config{
		v: viper.New(),
	}
}

// AddPath adds a path to search for config files
func (c *Config) AddPath(path string) *Config {
	c.v.AddConfigPath(path)
	return c
}

// SetEnvPrefix sets the prefix for environment variables
func (c *Config) SetEnvPrefix(prefix string) *Config {
	if prefix != "" {
		c.v.SetEnvPrefix(prefix)
	}
	return c
}

// AutomaticEnv enables automatic environment variable binding
func (c *Config) AutomaticEnv() *Config {
	c.v.AutomaticEnv()
	c.v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	return c
}

// LoadFile loads configuration from the specified file
func (c *Config) LoadFile(filename string) error {
	ext := filepath.Ext(filename)
	if ext == "" {
		return fmt.Errorf("file must have an extension")
	}

	c.v.SetConfigFile(filename)
	c.v.SetConfigType(strings.TrimPrefix(ext, "."))

	if err := c.v.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	return nil
}

// MergeFile merges configuration from the specified file
func (c *Config) MergeFile(filename string) error {
	ext := filepath.Ext(filename)
	if ext == "" {
		return fmt.Errorf("file must have an extension")
	}

	c.v.SetConfigFile(filename)
	c.v.SetConfigType(strings.TrimPrefix(ext, "."))

	if err := c.v.MergeInConfig(); err != nil {
		return fmt.Errorf("failed to merge config file: %w", err)
	}

	return nil
}

// Get retrieves a value from config with type inference
func (c *Config) Get(key string) interface{} {
	return c.v.Get(key)
}

// GetString retrieves a string value from config
func (c *Config) GetString(key string) string {
	return c.v.GetString(key)
}

// GetInt retrieves an integer value from config
func (c *Config) GetInt(key string) int {
	return c.v.GetInt(key)
}

// GetBool retrieves a boolean value from config
func (c *Config) GetBool(key string) bool {
	return c.v.GetBool(key)
}

// GetFloat64 retrieves a float64 value from config
func (c *Config) GetFloat64(key string) float64 {
	return c.v.GetFloat64(key)
}

// GetStringSlice retrieves a string slice from config
func (c *Config) GetStringSlice(key string) []string {
	return c.v.GetStringSlice(key)
}

// Unmarshal unmarshals the config into a struct
func (c *Config) Unmarshal(key string, val interface{}) error {
	return c.v.UnmarshalKey(key, val, func(dc *mapstructure.DecoderConfig) {
		dc.TagName = "config"
	})
}

// UnmarshalAll unmarshals the entire config into a struct
func (c *Config) UnmarshalAll(val interface{}) error {
	return c.v.Unmarshal(val, func(dc *mapstructure.DecoderConfig) {
		dc.TagName = "config"
	})
}

// Watch watches for config changes and calls the callback function
func (c *Config) Watch(callback func(fsnotify.Event)) {
	c.v.OnConfigChange(callback)
	c.v.WatchConfig()
}
