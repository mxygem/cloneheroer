package watcher

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watcher monitors a directory for new image files.
type Watcher struct {
	watchDir     string
	processedDir string
	failedDir    string
	onNewFile    func(string) error
	watcher      *fsnotify.Watcher
	processed    map[string]bool
}

// NewWatcher creates a new file watcher.
func NewWatcher(watchDir, processedDir, failedDir string, onNewFile func(string) error) (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	w := &Watcher{
		watchDir:     watchDir,
		processedDir: processedDir,
		failedDir:    failedDir,
		onNewFile:    onNewFile,
		watcher:      watcher,
		processed:    make(map[string]bool),
	}

	return w, nil
}

// Start begins watching the directory for new files.
func (w *Watcher) Start(ctx context.Context) error {
	// Create directories if they don't exist
	if err := os.MkdirAll(w.watchDir, 0755); err != nil {
		return fmt.Errorf("failed to create watch directory: %w", err)
	}
	if w.processedDir != "" {
		if err := os.MkdirAll(w.processedDir, 0755); err != nil {
			return fmt.Errorf("failed to create processed directory: %w", err)
		}
	}
	if w.failedDir != "" {
		if err := os.MkdirAll(w.failedDir, 0755); err != nil {
			return fmt.Errorf("failed to create failed directory: %w", err)
		}
	}

	// Add watch directory
	if err := w.watcher.Add(w.watchDir); err != nil {
		return fmt.Errorf("failed to add watch directory: %w", err)
	}

	// Process existing files in the directory
	if err := w.processExistingFiles(); err != nil {
		log.Printf("warning: failed to process existing files: %v", err)
	}

	// Start watching for new files
	go w.watchLoop(ctx)

	return nil
}

// processExistingFiles processes all image files already in the watch directory.
func (w *Watcher) processExistingFiles() error {
	entries, err := os.ReadDir(w.watchDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !isImageFile(entry.Name()) {
			continue
		}

		filePath := filepath.Join(w.watchDir, entry.Name())
		if w.processed[filePath] {
			continue
		}

		// Wait a bit to ensure file is fully written
		time.Sleep(100 * time.Millisecond)

		if err := w.handleFile(filePath); err != nil {
			log.Printf("error processing existing file %s: %v", filePath, err)
		} else {
			w.processed[filePath] = true
		}
	}

	return nil
}

// watchLoop monitors for file system events.
func (w *Watcher) watchLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Create == fsnotify.Create {
				if isImageFile(event.Name) {
					// Wait a bit to ensure file is fully written
					time.Sleep(500 * time.Millisecond)
					if err := w.handleFile(event.Name); err != nil {
						log.Printf("error processing file %s: %v", event.Name, err)
					} else {
						w.processed[event.Name] = true
					}
				}
			}
		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("watcher error: %v", err)
		}
	}
}

// handleFile processes a single image file.
func (w *Watcher) handleFile(filePath string) error {
	if w.processed[filePath] {
		return nil
	}

	log.Printf("processing new file: %s", filePath)

	// Call the callback to process the file
	if err := w.onNewFile(filePath); err != nil {
		// Move to failed directory if configured
		if w.failedDir != "" {
			destPath := filepath.Join(w.failedDir, filepath.Base(filePath))
			if moveErr := os.Rename(filePath, destPath); moveErr != nil {
				log.Printf("failed to move file to failed directory: %v", moveErr)
			} else {
				log.Printf("moved failed file to: %s", destPath)
			}
		}
		return fmt.Errorf("failed to process file: %w", err)
	}

	// Move to processed directory if configured
	if w.processedDir != "" {
		destPath := filepath.Join(w.processedDir, filepath.Base(filePath))
		if err := os.Rename(filePath, destPath); err != nil {
			log.Printf("failed to move file to processed directory: %v", err)
		} else {
			log.Printf("moved processed file to: %s", destPath)
		}
	}

	return nil
}

// Close stops watching and releases resources.
func (w *Watcher) Close() error {
	return w.watcher.Close()
}

// isImageFile checks if a filename is an image file.
func isImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".webp"
}

