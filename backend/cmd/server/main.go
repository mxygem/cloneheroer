package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"cloneheroer/internal/config"
	"cloneheroer/internal/db"
	"cloneheroer/internal/parser"
	"cloneheroer/internal/server"
	"cloneheroer/internal/watcher"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("shutting down...")
		cancel()
	}()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	if cfg.MigrateOnStart {
		if err := runMigrations(cfg.DatabaseURL); err != nil {
			log.Fatalf("migration failed: %v", err)
		}
	}

	repo := db.NewRepo(pool)

	// Initialize parser with configurable image dimensions
	imgParser, err := parser.NewParser(cfg.MaxImageWidth, cfg.MaxImageHeight)
	if err != nil {
		log.Fatalf("failed to create parser: %v", err)
	}
	defer imgParser.Close()

	// Create file processor function
	processFile := func(filePath string) error {
		log.Printf("parsing image: %s", filePath)
		scoreData, err := imgParser.ParseImage(filePath)
		if err != nil {
			return fmt.Errorf("failed to parse image: %w", err)
		}

		log.Printf("creating score for: %s - %s", scoreData.Artist, scoreData.SongName)
		scoreID, err := repo.CreateScore(ctx, *scoreData)
		if err != nil {
			return fmt.Errorf("failed to create score: %w", err)
		}

		log.Printf("successfully created score with ID: %d", scoreID)
		return nil
	}

	// Initialize and start file watcher
	fileWatcher, err := watcher.NewWatcher(cfg.WatchDir, cfg.ProcessedDir, cfg.FailedDir, processFile)
	if err != nil {
		log.Fatalf("failed to create watcher: %v", err)
	}
	defer fileWatcher.Close()

	if err := fileWatcher.Start(ctx); err != nil {
		log.Fatalf("failed to start watcher: %v", err)
	}
	log.Printf("watching directory: %q", cfg.WatchDir)

	// Start HTTP server
	srv := server.New(repo)
	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("starting server on %s", addr)

	// Run server in a goroutine
	serverErr := make(chan error, 1)
	go func() {
		if err := srv.Start(addr); err != nil {
			serverErr <- err
		}
	}()

	// Wait for context cancellation or server error
	select {
	case <-ctx.Done():
		log.Println("context cancelled, shutting down")
	case err := <-serverErr:
		log.Fatalf("server error: %v", err)
	}
}

func runMigrations(databaseURL string) error {
	// Try to find migrations directory relative to the executable or working directory
	// First, try relative to backend directory (when running from project root)
	migrationPath := "file://backend/migrations"

	// If that doesn't work, try relative to current working directory
	wd, err := os.Getwd()
	if err == nil {
		// Check if we're in the backend directory
		if filepath.Base(wd) == "backend" {
			migrationPath = "file://migrations"
		} else {
			// Try absolute path
			migrationPath = fmt.Sprintf("file://%s/backend/migrations", wd)
		}
	}

	m, err := migrate.New(
		migrationPath,
		databaseURL,
	)
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}
