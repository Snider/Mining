package mining

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Repository defines a generic interface for data persistence.
// Implementations can store data in files, databases, etc.
type Repository[T any] interface {
	// Load reads data from the repository
	Load() (T, error)

	// Save writes data to the repository
	Save(data T) error

	// Update atomically loads, modifies, and saves data
	Update(fn func(*T) error) error
}

// FileRepository provides atomic file-based persistence for JSON data.
// It uses atomic writes (temp file + rename) to prevent corruption.
type FileRepository[T any] struct {
	mu       sync.RWMutex
	path     string
	defaults func() T
}

// FileRepositoryOption configures a FileRepository.
type FileRepositoryOption[T any] func(*FileRepository[T])

// WithDefaults sets the default value factory for when the file doesn't exist.
func WithDefaults[T any](fn func() T) FileRepositoryOption[T] {
	return func(r *FileRepository[T]) {
		r.defaults = fn
	}
}

// NewFileRepository creates a new file-based repository.
func NewFileRepository[T any](path string, opts ...FileRepositoryOption[T]) *FileRepository[T] {
	r := &FileRepository[T]{
		path: path,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Load reads and deserializes data from the file.
// Returns defaults if file doesn't exist.
func (r *FileRepository[T]) Load() (T, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result T

	data, err := os.ReadFile(r.path)
	if err != nil {
		if os.IsNotExist(err) {
			if r.defaults != nil {
				return r.defaults(), nil
			}
			return result, nil
		}
		return result, fmt.Errorf("failed to read file: %w", err)
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return result, fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return result, nil
}

// Save serializes and writes data to the file atomically.
func (r *FileRepository[T]) Save(data T) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.saveUnlocked(data)
}

// saveUnlocked saves data without acquiring the lock (caller must hold lock).
func (r *FileRepository[T]) saveUnlocked(data T) error {
	dir := filepath.Dir(r.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	return AtomicWriteFile(r.path, jsonData, 0600)
}

// Update atomically loads, modifies, and saves data.
// The modification function receives a pointer to the data.
func (r *FileRepository[T]) Update(fn func(*T) error) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Load current data
	var data T
	fileData, err := os.ReadFile(r.path)
	if err != nil {
		if os.IsNotExist(err) {
			if r.defaults != nil {
				data = r.defaults()
			}
		} else {
			return fmt.Errorf("failed to read file: %w", err)
		}
	} else {
		if err := json.Unmarshal(fileData, &data); err != nil {
			return fmt.Errorf("failed to unmarshal data: %w", err)
		}
	}

	// Apply modification
	if err := fn(&data); err != nil {
		return err
	}

	// Save atomically
	return r.saveUnlocked(data)
}

// Path returns the file path of this repository.
func (r *FileRepository[T]) Path() string {
	return r.path
}

// Exists returns true if the repository file exists.
func (r *FileRepository[T]) Exists() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, err := os.Stat(r.path)
	return err == nil
}

// Delete removes the repository file.
func (r *FileRepository[T]) Delete() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	err := os.Remove(r.path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
