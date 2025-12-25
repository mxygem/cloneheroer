package config

import (
	"log"
	"path/filepath"
	"strings"

	"github.com/caarlos0/env/v11"
)

// Config holds runtime configuration loaded from environment variables.
type Config struct {
	WatchDir       string `env:"WATCH_DIR,required"`
	DatabaseURL    string `env:"DATABASE_URL,required"`
	Port           int    `env:"PORT" envDefault:"3000"`
	LogLevel       string `env:"LOG_LEVEL" envDefault:"info"`
	MigrateOnStart bool   `env:"MIGRATE_ON_START" envDefault:"true"`
	ProcessedDir   string `env:"PROCESSED_DIR" envDefault:""`
	FailedDir      string `env:"FAILED_DIR" envDefault:""`
	MaxImageWidth  int    `env:"MAX_IMAGE_WIDTH" envDefault:"1920"`
	MaxImageHeight int    `env:"MAX_IMAGE_HEIGHT" envDefault:"1080"`
}

// normalizePath normalizes a file path, handling spaces and ensuring it's absolute.
func normalizePath(path string) string {
	if path == "" {
		return ""
	}

	// First, replace any escaped backslashes or double backslashes with single forward slashes
	// This handles cases where environment variables have "Clone\ Hero" or "Clone\\ Hero"
	normalized := strings.ReplaceAll(path, "\\", "/")
	normalized = strings.ReplaceAll(normalized, "//", "/")

	// Clean the path to remove any remaining weird characters
	cleaned := filepath.Clean(normalized)

	// Convert to absolute path to handle relative paths and spaces properly
	abs, err := filepath.Abs(cleaned)
	if err != nil {
		log.Printf("warning: failed to normalize path %q: %v, using cleaned path", path, err)
		return cleaned
	}

	return abs
}

// Load parses environment variables into a Config struct.
func Load() Config {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Normalize all directory paths to handle spaces and ensure they're absolute
	originalWatchDir := cfg.WatchDir
	cfg.WatchDir = normalizePath(cfg.WatchDir)
	if cfg.WatchDir != originalWatchDir {
		log.Printf("normalized WATCH_DIR: %q -> %q", originalWatchDir, cfg.WatchDir)
	}

	if cfg.ProcessedDir != "" {
		originalProcessedDir := cfg.ProcessedDir
		cfg.ProcessedDir = normalizePath(cfg.ProcessedDir)
		if cfg.ProcessedDir != originalProcessedDir {
			log.Printf("normalized PROCESSED_DIR: %q -> %q", originalProcessedDir, cfg.ProcessedDir)
		}
	}

	if cfg.FailedDir != "" {
		originalFailedDir := cfg.FailedDir
		cfg.FailedDir = normalizePath(cfg.FailedDir)
		if cfg.FailedDir != originalFailedDir {
			log.Printf("normalized FAILED_DIR: %q -> %q", originalFailedDir, cfg.FailedDir)
		}
	}

	return cfg
}
