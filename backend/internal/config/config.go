package config

import (
	"log"

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
	MaxImageWidth  int    `env:"MAX_oIMAGE_WIDTH" envDefault:"1920"`
	MaxImageHeight int    `env:"MAX_IMAGE_HEIGHT" envDefault:"1080"`
}

// Load parses environment variables into a Config struct.
func Load() Config {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	return cfg
}
