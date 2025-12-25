package watcher

import (
	"context"
	"fmt"
	"io"
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

// normalizePath normalizes a file path, handling spaces and ensuring it's absolute.
func normalizePath(path string) (string, error) {
	if path == "" {
		return "", nil
	}
	// Clean the path to remove any weird characters
	cleaned := filepath.Clean(path)
	// Convert to absolute path to handle relative paths and spaces properly
	abs, err := filepath.Abs(cleaned)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for %q: %w", path, err)
	}
	return abs, nil
}

// NewWatcher creates a new file watcher.
func NewWatcher(watchDir, processedDir, failedDir string, onNewFile func(string) error) (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	// Normalize all paths to handle spaces and ensure they're absolute
	normalizedWatchDir, err := normalizePath(watchDir)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize watch directory: %w", err)
	}

	var normalizedProcessedDir string
	if processedDir != "" {
		normalizedProcessedDir, err = normalizePath(processedDir)
		if err != nil {
			return nil, fmt.Errorf("failed to normalize processed directory: %w", err)
		}
	}

	var normalizedFailedDir string
	if failedDir != "" {
		normalizedFailedDir, err = normalizePath(failedDir)
		if err != nil {
			return nil, fmt.Errorf("failed to normalize failed directory: %w", err)
		}
	}

	w := &Watcher{
		watchDir:     normalizedWatchDir,
		processedDir: normalizedProcessedDir,
		failedDir:    normalizedFailedDir,
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

	// Add watch directory (already normalized in NewWatcher)
	log.Printf("adding watch directory: %q", w.watchDir)
	if err := w.watcher.Add(w.watchDir); err != nil {
		return fmt.Errorf("failed to add watch directory %q: %w", w.watchDir, err)
	}
	log.Printf("successfully watching directory: %q", w.watchDir)

	// Process existing files in the directory
	if err := w.processExistingFiles(); err != nil {
		log.Printf("warning: failed to process existing files: %v", err)
	}

	// Start watching for new files via fsnotify
	go w.watchLoop(ctx)

	// Start polling as a fallback (important for Windows mounts in WSL2)
	// fsnotify may not detect files created on the Windows side
	go w.pollLoop(ctx)

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

		loc := filepath.Join(w.watchDir, entry.Name())
		// Normalize the path to ensure consistent handling
		normalizedLoc, err := filepath.Abs(loc)
		if err != nil {
			log.Printf("warning: failed to normalize path %q: %v", loc, err)
			normalizedLoc = loc
		}

		log.Printf("processing existing file: %q", normalizedLoc)
		if w.processed[normalizedLoc] {
			log.Printf("skipping already processed file: %q", normalizedLoc)
			continue
		}

		// Wait a bit to ensure file is fully written
		time.Sleep(100 * time.Millisecond)

		if err := w.handleFile(normalizedLoc); err != nil {
			log.Printf("error processing existing file %s: %v", normalizedLoc, err)
		} else {
			w.processed[normalizedLoc] = true
		}
	}

	return nil
}

// pollLoop periodically checks for new files as a fallback for fsnotify.
// This is especially important for Windows mounts in WSL2 where fsnotify
// may not reliably detect files created on the Windows side.
func (w *Watcher) pollLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second) // Poll every 5 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Check for new files
			entries, err := os.ReadDir(w.watchDir)
			if err != nil {
				log.Printf("error reading watch directory during poll: %v", err)
				continue
			}

			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}
				if !isImageFile(entry.Name()) {
					continue
				}

				loc := filepath.Join(w.watchDir, entry.Name())
				normalizedLoc, err := filepath.Abs(loc)
				if err != nil {
					log.Printf("warning: failed to normalize path during poll %q: %v", loc, err)
					normalizedLoc = loc
				}

				// Skip if already processed
				if w.processed[normalizedLoc] {
					continue
				}

				// Process the file
				log.Printf("polling detected new file: %q", normalizedLoc)
				if err := w.handleFile(normalizedLoc); err != nil {
					log.Printf("error processing file from poll: %s: %v", normalizedLoc, err)
				} else {
					w.processed[normalizedLoc] = true
				}
			}
		}
	}
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
			log.Printf("file system event: %v (op: %v)", event, event.Op)
			if event.Op&fsnotify.Create == fsnotify.Create || event.Op&fsnotify.Write == fsnotify.Write {
				// Normalize the event path
				normalizedPath, err := filepath.Abs(event.Name)
				if err != nil {
					log.Printf("warning: failed to normalize event path %q: %v", event.Name, err)
					normalizedPath = event.Name
				}

				if isImageFile(normalizedPath) {
					// Check if already processed
					if w.processed[normalizedPath] {
						log.Printf("skipping already processed file: %q", normalizedPath)
						continue
					}

					// Wait a bit to ensure file is fully written (longer wait for new files)
					time.Sleep(500 * time.Millisecond)

					// Verify file exists and is readable before processing
					if _, err := os.Stat(normalizedPath); os.IsNotExist(err) {
						log.Printf("file no longer exists, skipping: %q", normalizedPath)
						continue
					}

					log.Printf("processing new file from watcher: %q", normalizedPath)
					if err := w.handleFile(normalizedPath); err != nil {
						log.Printf("error processing file %s: %v", normalizedPath, err)
					} else {
						w.processed[normalizedPath] = true
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

// moveFile moves a file from source to destination, verifying the move succeeded.
// On Windows/WSL2 mounts, os.Rename may fail across filesystems, so we use copy+delete as fallback.
func (w *Watcher) moveFile(src, destDir string) error {
	if destDir == "" {
		return fmt.Errorf("destination directory is empty")
	}

	// Normalize paths to ensure no double backslashes or other issues
	src = filepath.Clean(src)
	destDir = filepath.Clean(destDir)
	destPath := filepath.Join(destDir, filepath.Base(src))
	destPath = filepath.Clean(destPath)

	log.Printf("attempting to move file:")
	log.Printf("  source: %q", src)
	log.Printf("  destination: %q", destPath)

	// Ensure destination directory exists
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory %q: %w", destDir, err)
	}

	// Check if source file exists
	srcInfo, err := os.Stat(src)
	if os.IsNotExist(err) {
		return fmt.Errorf("source file does not exist: %q", src)
	}
	if err != nil {
		return fmt.Errorf("failed to stat source file %q: %w", src, err)
	}

	// Check if destination already exists
	if _, err := os.Stat(destPath); err == nil {
		log.Printf("warning: destination file already exists, removing: %q", destPath)
		if err := os.Remove(destPath); err != nil {
			return fmt.Errorf("failed to remove existing destination file %q: %w", destPath, err)
		}
	}

	// Try to rename (move) the file first (fastest method)
	renameErr := os.Rename(src, destPath)
	if renameErr == nil {
		// Rename succeeded, verify it
		time.Sleep(100 * time.Millisecond) // Brief pause for filesystem sync

		// Check if source still exists
		if _, err := os.Stat(src); !os.IsNotExist(err) {
			log.Printf("warning: source file still exists after rename, attempting removal: %q", src)
			if removeErr := os.Remove(src); removeErr != nil {
				log.Printf("warning: failed to remove source file after rename: %v", removeErr)
			}
		}

		// Verify destination exists
		if _, err := os.Stat(destPath); os.IsNotExist(err) {
			return fmt.Errorf("destination file does not exist after rename: %q", destPath)
		}

		log.Printf("successfully moved file via rename to: %q", destPath)
		return nil
	}

	// Rename failed (likely cross-filesystem), use copy+delete
	log.Printf("rename failed (likely cross-filesystem), using copy+delete: %v", renameErr)

	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file %q: %w", src, err)
	}
	defer srcFile.Close()

	// Create destination file
	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file %q: %w", destPath, err)
	}
	defer destFile.Close()

	// Copy file contents
	bytesCopied, err := io.Copy(destFile, srcFile)
	if err != nil {
		destFile.Close()
		os.Remove(destPath) // Clean up partial file
		return fmt.Errorf("failed to copy file: %w", err)
	}

	// Sync destination file to ensure it's written
	if err := destFile.Sync(); err != nil {
		log.Printf("warning: failed to sync destination file: %v", err)
	}

	// Close files before verifying
	srcFile.Close()
	destFile.Close()

	// Verify copy succeeded
	destInfo, err := os.Stat(destPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("destination file does not exist after copy: %q", destPath)
	}
	if err != nil {
		return fmt.Errorf("failed to stat destination file: %w", err)
	}

	if destInfo.Size() != srcInfo.Size() {
		os.Remove(destPath) // Clean up incorrect file
		return fmt.Errorf("file size mismatch: source %d bytes, destination %d bytes", srcInfo.Size(), destInfo.Size())
	}

	if bytesCopied != srcInfo.Size() {
		log.Printf("warning: copied %d bytes but expected %d bytes", bytesCopied, srcInfo.Size())
	}

	log.Printf("successfully copied file (%d bytes), now removing source", bytesCopied)

	// Remove source file
	if err := os.Remove(src); err != nil {
		log.Printf("error: failed to remove source file after copy: %v", err)
		// Don't fail here - the copy succeeded, we just have a duplicate
		// But log it so the user knows
		return fmt.Errorf("file copied successfully but failed to remove source: %w", err)
	}

	// Verify source is gone
	time.Sleep(100 * time.Millisecond)
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		log.Printf("warning: source file still exists after removal: %q", src)
	}

	log.Printf("successfully moved file via copy+delete to: %q", destPath)
	return nil
}

// handleFile processes a single image file.
func (w *Watcher) handleFile(loc string) error {
	// Normalize the path
	normalizedLoc, err := filepath.Abs(loc)
	if err != nil {
		log.Printf("warning: failed to normalize path %q: %v", loc, err)
		normalizedLoc = loc
	}

	if w.processed[normalizedLoc] {
		log.Printf("file already processed, skipping: %q", normalizedLoc)
		return nil
	}

	log.Printf("processing file: %q", normalizedLoc)

	// Verify file exists before processing
	if _, err := os.Stat(normalizedLoc); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %q", normalizedLoc)
	}

	// Call the callback to process the file
	if err := w.onNewFile(normalizedLoc); err != nil {
		// Move to failed directory if configured
		if w.failedDir != "" {
			if moveErr := w.moveFile(normalizedLoc, w.failedDir); moveErr != nil {
				log.Printf("error: failed to move file to failed directory: %v", moveErr)
			}
		}
		return fmt.Errorf("failed to process file: %w", err)
	}

	// Move to processed directory if configured
	if w.processedDir != "" {
		if moveErr := w.moveFile(normalizedLoc, w.processedDir); moveErr != nil {
			log.Printf("error: failed to move file to processed directory: %v", moveErr)
			// Don't return error here - file was processed successfully, just move failed
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
