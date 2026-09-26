package storage

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestNewFileRepo(t *testing.T) {
	t.Run("creates repo with existing directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		repo, err := NewFileRepo(tmpDir)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if repo.Dir != tmpDir {
			t.Errorf("expected Dir=%s, got %s", tmpDir, repo.Dir)
		}
	})

	t.Run("creates repo and directory if not exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		newDir := filepath.Join(tmpDir, "nested", "storage")

		repo, err := NewFileRepo(newDir)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if _, err := os.Stat(newDir); os.IsNotExist(err) {
			t.Errorf("expected directory to be created, but it doesn't exist")
		}
		if repo.Dir != newDir {
			t.Errorf("expected Dir=%s, got %s", newDir, repo.Dir)
		}
	})

	t.Run("returns error if directory creation fails", func(t *testing.T) {
		// Windows doesn't easily allow this test, but we can test with an invalid path
		// On some systems, a path with invalid characters will fail
		invalidPath := "/dev/null/cannot/create/here"
		if runtime.GOOS == "windows" {
			invalidPath = "CON:test" // Reserved name on Windows
		}

		_, err := NewFileRepo(invalidPath)
		if err == nil && runtime.GOOS != "windows" {
			t.Errorf("expected error for invalid path, got nil")
		}
	})
}

func TestFileRepoSave(t *testing.T) {
	tmpDir := t.TempDir()
	repo, _ := NewFileRepo(tmpDir)

	t.Run("saves file with content", func(t *testing.T) {
		content := []byte("test file content")
		reader := bytes.NewReader(content)

		saved, err := repo.Save("test.txt", reader)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if saved.Name != "test.txt" {
			t.Errorf("expected Name=test.txt, got %s", saved.Name)
		}
		if saved.Size != int64(len(content)) {
			t.Errorf("expected Size=%d, got %d", len(content), saved.Size)
		}

		// Verify file actually exists on disk
		data, err := os.ReadFile(saved.Path)
		if err != nil {
			t.Fatalf("expected file to exist, got error: %v", err)
		}
		if !bytes.Equal(data, content) {
			t.Errorf("expected file content %s, got %s", content, data)
		}
	})

	t.Run("saves empty file", func(t *testing.T) {
		reader := bytes.NewReader([]byte{})

		saved, err := repo.Save("empty.txt", reader)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if saved.Size != 0 {
			t.Errorf("expected Size=0, got %d", saved.Size)
		}
	})

	t.Run("saves large file", func(t *testing.T) {
		largeContent := make([]byte, 1024*1024) // 1MB
		for i := range largeContent {
			largeContent[i] = byte(i % 256)
		}
		reader := bytes.NewReader(largeContent)

		saved, err := repo.Save("large.bin", reader)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if saved.Size != int64(len(largeContent)) {
			t.Errorf("expected Size=%d, got %d", len(largeContent), saved.Size)
		}
	})

	t.Run("sets path correctly", func(t *testing.T) {
		reader := bytes.NewReader([]byte("content"))
		filename := "test.txt"

		saved, err := repo.Save(filename, reader)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		expectedPath := filepath.Join(tmpDir, filename)
		if saved.Path != expectedPath {
			t.Errorf("expected Path=%s, got %s", expectedPath, saved.Path)
		}
	})

	t.Run("overwrites existing file", func(t *testing.T) {
		filename := "overwrite.txt"

		// Save first file
		_, err := repo.Save(filename, bytes.NewReader([]byte("first content")))
		if err != nil {
			t.Fatalf("first save failed: %v", err)
		}

		// Save second file with same name
		content2 := []byte("second content")
		saved, err := repo.Save(filename, bytes.NewReader(content2))
		if err != nil {
			t.Fatalf("second save failed: %v", err)
		}

		if saved.Size != int64(len(content2)) {
			t.Errorf("expected Size=%d, got %d", len(content2), saved.Size)
		}

		data, _ := os.ReadFile(saved.Path)
		if !bytes.Equal(data, content2) {
			t.Errorf("expected file to contain second content")
		}
	})

	t.Run("handles special characters in filename", func(t *testing.T) {
		filenames := []string{
			"file with spaces.txt",
			"file-with-dashes.txt",
			"file_with_underscores.txt",
		}

		for _, filename := range filenames {
			reader := bytes.NewReader([]byte("content"))
			saved, err := repo.Save(filename, reader)
			if err != nil {
				t.Errorf("failed to save %s: %v", filename, err)
				continue
			}

			if saved.Name != filename {
				t.Errorf("expected Name=%s, got %s", filename, saved.Name)
			}

			if _, err := os.Stat(saved.Path); os.IsNotExist(err) {
				t.Errorf("expected file %s to exist", saved.Path)
			}
		}
	})
}

func TestFileRepoList(t *testing.T) {
	tmpDir := t.TempDir()
	repo, _ := NewFileRepo(tmpDir)

	t.Run("returns empty list for empty directory", func(t *testing.T) {
		files, err := repo.List()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(files) != 0 {
			t.Errorf("expected empty list, got %d files", len(files))
		}
	})

	t.Run("lists single file", func(t *testing.T) {
		repo.Save("file1.txt", bytes.NewReader([]byte("content1")))

		files, err := repo.List()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(files) != 1 {
			t.Errorf("expected 1 file, got %d", len(files))
		}
		if files[0].Name != "file1.txt" {
			t.Errorf("expected file1.txt, got %s", files[0].Name)
		}
		if files[0].Size != 8 {
			t.Errorf("expected size 8, got %d", files[0].Size)
		}
	})

	t.Run("lists multiple files", func(t *testing.T) {
		tmpDir2 := t.TempDir()
		repo2, _ := NewFileRepo(tmpDir2)

		repo2.Save("file1.txt", bytes.NewReader([]byte("content1")))
		repo2.Save("file2.txt", bytes.NewReader([]byte("content2")))
		repo2.Save("file3.txt", bytes.NewReader([]byte("content3")))

		files, err := repo2.List()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(files) != 3 {
			t.Errorf("expected 3 files, got %d", len(files))
		}

		// Verify all files are present
		names := make(map[string]bool)
		for _, f := range files {
			names[f.Name] = true
		}

		for i := 1; i <= 3; i++ {
			filename := "file" + string(rune('0'+i)) + ".txt"
			if !names[filename] {
				t.Errorf("expected %s to be listed", filename)
			}
		}
	})

	t.Run("ignores directories", func(t *testing.T) {
		tmpDir3 := t.TempDir()
		repo3, _ := NewFileRepo(tmpDir3)

		repo3.Save("file.txt", bytes.NewReader([]byte("content")))
		os.Mkdir(filepath.Join(tmpDir3, "subdir"), os.ModePerm)

		files, err := repo3.List()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(files) != 1 {
			t.Errorf("expected 1 file (directories should be ignored), got %d", len(files))
		}
	})

	t.Run("handles errors gracefully and returns partial results", func(t *testing.T) {
		tmpDir4 := t.TempDir()
		repo4, _ := NewFileRepo(tmpDir4)

		// Create some valid files
		repo4.Save("file1.txt", bytes.NewReader([]byte("content1")))
		repo4.Save("file2.txt", bytes.NewReader([]byte("content2")))

		files, err := repo4.List()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Should still return files despite any errors
		if len(files) < 2 {
			t.Errorf("expected at least 2 files, got %d", len(files))
		}
	})

	t.Run("returns correct path for each file", func(t *testing.T) {
		tmpDir5 := t.TempDir()
		repo5, _ := NewFileRepo(tmpDir5)

		repo5.Save("test.txt", bytes.NewReader([]byte("content")))

		files, _ := repo5.List()
		if len(files) > 0 {
			expectedPath := filepath.Join(tmpDir5, "test.txt")
			if files[0].Path != expectedPath {
				t.Errorf("expected Path=%s, got %s", expectedPath, files[0].Path)
			}
		}
	})
}

func TestFileRepoSaveAndList(t *testing.T) {
	t.Run("integration: save multiple files and list them", func(t *testing.T) {
		tmpDir := t.TempDir()
		repo, _ := NewFileRepo(tmpDir)

		testFiles := []struct {
			name    string
			content string
		}{
			{"readme.md", "# README"},
			{"config.json", `{"debug": true}`},
			{"data.bin", "\x00\x01\x02\x03"},
		}

		for _, tf := range testFiles {
			_, err := repo.Save(tf.name, bytes.NewReader([]byte(tf.content)))
			if err != nil {
				t.Fatalf("failed to save %s: %v", tf.name, err)
			}
		}

		files, err := repo.List()
		if err != nil {
			t.Fatalf("failed to list: %v", err)
		}

		if len(files) != len(testFiles) {
			t.Errorf("expected %d files, got %d", len(testFiles), len(files))
		}

		for _, tf := range testFiles {
			found := false
			for _, f := range files {
				if f.Name == tf.name {
					found = true
					if f.Size != int64(len(tf.content)) {
						t.Errorf("%s: expected size %d, got %d", tf.name, len(tf.content), f.Size)
					}
					break
				}
			}
			if !found {
				t.Errorf("expected to find %s in list", tf.name)
			}
		}
	})
}
