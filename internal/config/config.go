package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Config holds all application configuration
type Config struct {
	// Server settings
	Server ServerConfig

	// Directory paths
	Paths PathConfig

	// Database settings
	Database DatabaseConfig

	// API settings
	API APIConfig

	// Logging settings
	Logging LoggingConfig

	// Feature flags
	Features FeatureConfig
}

type ServerConfig struct {
	Port string
	Host string
}

type PathConfig struct {
	VideoDir        string
	ThumbnailDir    string
	PerformerDir    string
	OldLogsDir      string
	LogFile         string
}

type DatabaseConfig struct {
	Path         string
	MaxOpenConns int
	MaxIdleConns int
}

type APIConfig struct {
	AdultDataLinkKey string
}

type LoggingConfig struct {
	Level  string
	Format string
}

type FeatureConfig struct {
	EnableAutoScan      bool
	EnableMetadataFetch bool
}

var globalConfig *Config

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Host: getEnv("SERVER_HOST", "localhost"),
		},
		Paths: PathConfig{
			VideoDir:     getEnv("VIDEO_DIR", ""),
			ThumbnailDir: getEnv("THUMBNAIL_DIR", "frontend/.thumbnails"),
			PerformerDir: getEnv("PERFORMER_DIR", "frontend/.performers"),
			OldLogsDir:   getEnv("OLD_LOGS_DIR", "old_logs"),
			LogFile:      getEnv("LOG_FILE", "app.log"),
		},
		Database: DatabaseConfig{
			Path:         getEnv("DB_PATH", "./video_organizer.db"),
			MaxOpenConns: getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns: getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
		},
		API: APIConfig{
			AdultDataLinkKey: getEnv("ADULTDATALINK_API_KEY", ""),
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "text"),
		},
		Features: FeatureConfig{
			EnableAutoScan:      getEnvAsBool("ENABLE_AUTO_SCAN", true),
			EnableMetadataFetch: getEnvAsBool("ENABLE_METADATA_FETCH", true),
		},
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	globalConfig = cfg
	return cfg, nil
}

// Get returns the global configuration instance
func Get() *Config {
	if globalConfig == nil {
		panic("configuration not loaded - call Load() first")
	}
	return globalConfig
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate required fields
	if c.Paths.VideoDir == "" {
		return fmt.Errorf("VIDEO_DIR is required")
	}

	// Check if video directory exists
	if _, err := os.Stat(c.Paths.VideoDir); os.IsNotExist(err) {
		return fmt.Errorf("video directory does not exist: %s", c.Paths.VideoDir)
	}

	// Ensure required directories exist or can be created
	dirs := []string{
		c.Paths.ThumbnailDir,
		c.Paths.PerformerDir,
		c.Paths.OldLogsDir,
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Validate API key if metadata fetch is enabled
	if c.Features.EnableMetadataFetch && c.API.AdultDataLinkKey == "" {
		return fmt.Errorf("ADULTDATALINK_API_KEY is required when ENABLE_METADATA_FETCH is true")
	}

	// Validate log level
	validLevels := []string{"debug", "info", "warning", "error"}
	if !contains(validLevels, strings.ToLower(c.Logging.Level)) {
		return fmt.Errorf("invalid LOG_LEVEL: must be one of %v", validLevels)
	}

	return nil
}

// GetAbsPath returns the absolute path for a given path
func (c *Config) GetAbsPath(path string) (string, error) {
	if filepath.IsAbs(path) {
		return path, nil
	}
	return filepath.Abs(path)
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Backward compatibility - these will be deprecated
var (
	VideoDir           string
	ThumbnailDir       string
	PerformerFoldersDir string
	OldLogsDir         string
	LogFile            string
)

// InitLegacyVars initializes the legacy global variables for backward compatibility
// Deprecated: Use the Config struct instead
func InitLegacyVars() {
	cfg := Get()
	VideoDir = cfg.Paths.VideoDir
	ThumbnailDir = cfg.Paths.ThumbnailDir
	PerformerFoldersDir = cfg.Paths.PerformerDir
	OldLogsDir = cfg.Paths.OldLogsDir
	LogFile = cfg.Paths.LogFile
}
