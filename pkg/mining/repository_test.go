package mining

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

type testData struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func TestFileRepository_Load(t *testing.T) {
	t.Run("NonExistentFile", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "nonexistent.json")
		repo := NewFileRepository[testData](path)

		data, err := repo.Load()
		if err != nil {
			t.Fatalf("Load should not error for non-existent file: %v", err)
		}
		if data.Name != "" || data.Value != 0 {
			t.Error("Expected zero value for non-existent file")
		}
	})

	t.Run("NonExistentFileWithDefaults", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "nonexistent.json")
		repo := NewFileRepository[testData](path, WithDefaults(func() testData {
			return testData{Name: "default", Value: 42}
		}))

		data, err := repo.Load()
		if err != nil {
			t.Fatalf("Load should not error: %v", err)
		}
		if data.Name != "default" || data.Value != 42 {
			t.Errorf("Expected default values, got %+v", data)
		}
	})

	t.Run("ExistingFile", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "test.json")

		// Write test data
		if err := os.WriteFile(path, []byte(`{"name":"test","value":123}`), 0600); err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		repo := NewFileRepository[testData](path)
		data, err := repo.Load()
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}
		if data.Name != "test" || data.Value != 123 {
			t.Errorf("Unexpected data: %+v", data)
		}
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "invalid.json")

		if err := os.WriteFile(path, []byte(`{invalid json}`), 0600); err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		repo := NewFileRepository[testData](path)
		_, err := repo.Load()
		if err == nil {
			t.Error("Expected error for invalid JSON")
		}
	})
}

func TestFileRepository_Save(t *testing.T) {
	t.Run("NewFile", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "subdir", "new.json")
		repo := NewFileRepository[testData](path)

		data := testData{Name: "saved", Value: 456}
		if err := repo.Save(data); err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		// Verify file was created
		if !repo.Exists() {
			t.Error("File should exist after save")
		}

		// Verify content
		loaded, err := repo.Load()
		if err != nil {
			t.Fatalf("Load after save failed: %v", err)
		}
		if loaded.Name != "saved" || loaded.Value != 456 {
			t.Errorf("Unexpected loaded data: %+v", loaded)
		}
	})

	t.Run("OverwriteExisting", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "existing.json")
		repo := NewFileRepository[testData](path)

		// Save initial data
		if err := repo.Save(testData{Name: "first", Value: 1}); err != nil {
			t.Fatalf("First save failed: %v", err)
		}

		// Overwrite
		if err := repo.Save(testData{Name: "second", Value: 2}); err != nil {
			t.Fatalf("Second save failed: %v", err)
		}

		// Verify overwrite
		loaded, err := repo.Load()
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}
		if loaded.Name != "second" || loaded.Value != 2 {
			t.Errorf("Expected overwritten data, got: %+v", loaded)
		}
	})
}

func TestFileRepository_Update(t *testing.T) {
	t.Run("UpdateExisting", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "update.json")
		repo := NewFileRepository[testData](path)

		// Save initial data
		if err := repo.Save(testData{Name: "initial", Value: 10}); err != nil {
			t.Fatalf("Initial save failed: %v", err)
		}

		// Update
		err := repo.Update(func(data *testData) error {
			data.Value += 5
			return nil
		})
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}

		// Verify update
		loaded, err := repo.Load()
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}
		if loaded.Value != 15 {
			t.Errorf("Expected value 15, got %d", loaded.Value)
		}
	})

	t.Run("UpdateNonExistentWithDefaults", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "new.json")
		repo := NewFileRepository[testData](path, WithDefaults(func() testData {
			return testData{Name: "default", Value: 100}
		}))

		err := repo.Update(func(data *testData) error {
			data.Value *= 2
			return nil
		})
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}

		// Verify update started from defaults
		loaded, err := repo.Load()
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}
		if loaded.Value != 200 {
			t.Errorf("Expected value 200, got %d", loaded.Value)
		}
	})

	t.Run("UpdateWithError", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "error.json")
		repo := NewFileRepository[testData](path)

		if err := repo.Save(testData{Name: "test", Value: 1}); err != nil {
			t.Fatalf("Initial save failed: %v", err)
		}

		// Update that returns error
		testErr := errors.New("update error")
		err := repo.Update(func(data *testData) error {
			data.Value = 999 // This change should not be saved
			return testErr
		})
		if err != testErr {
			t.Errorf("Expected test error, got: %v", err)
		}

		// Verify original data unchanged
		loaded, err := repo.Load()
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}
		if loaded.Value != 1 {
			t.Errorf("Expected value 1 (unchanged), got %d", loaded.Value)
		}
	})
}

func TestFileRepository_Delete(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "delete.json")
	repo := NewFileRepository[testData](path)

	// Save data
	if err := repo.Save(testData{Name: "temp", Value: 1}); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if !repo.Exists() {
		t.Error("File should exist after save")
	}

	// Delete
	if err := repo.Delete(); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	if repo.Exists() {
		t.Error("File should not exist after delete")
	}

	// Delete non-existent should not error
	if err := repo.Delete(); err != nil {
		t.Errorf("Delete non-existent should not error: %v", err)
	}
}

func TestFileRepository_Path(t *testing.T) {
	path := "/some/path/config.json"
	repo := NewFileRepository[testData](path)

	if repo.Path() != path {
		t.Errorf("Expected path %s, got %s", path, repo.Path())
	}
}

// Test with slice data
func TestFileRepository_SliceData(t *testing.T) {
	type item struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "items.json")
	repo := NewFileRepository[[]item](path, WithDefaults(func() []item {
		return []item{}
	}))

	// Save slice
	items := []item{
		{ID: "1", Name: "First"},
		{ID: "2", Name: "Second"},
	}
	if err := repo.Save(items); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Load and verify
	loaded, err := repo.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(loaded) != 2 {
		t.Errorf("Expected 2 items, got %d", len(loaded))
	}

	// Update slice
	err = repo.Update(func(data *[]item) error {
		*data = append(*data, item{ID: "3", Name: "Third"})
		return nil
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	loaded, _ = repo.Load()
	if len(loaded) != 3 {
		t.Errorf("Expected 3 items after update, got %d", len(loaded))
	}
}
