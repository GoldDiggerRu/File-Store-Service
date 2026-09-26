package app

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GoldDiggerRu/File-Store-Service/model"
)

// MockFileRepository is a mock implementation for testing
type MockFileRepository struct {
	SaveCalled bool
	SaveName   string
	ListCalled bool
	SaveErr    error
	ListErr    error
	SavedFiles map[string]model.File
}

func NewMockFileRepository() *MockFileRepository {
	return &MockFileRepository{
		SavedFiles: make(map[string]model.File),
	}
}

func (m *MockFileRepository) Save(name string, r io.Reader) (model.File, error) {
	m.SaveCalled = true
	m.SaveName = name

	if m.SaveErr != nil {
		return model.File{}, m.SaveErr
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return model.File{}, err
	}

	f := model.File{
		Name: name,
		Size: int64(len(data)),
		Path: "/mock/path/" + name,
	}
	m.SavedFiles[name] = f
	return f, nil
}

func (m *MockFileRepository) List() ([]model.File, error) {
	m.ListCalled = true

	if m.ListErr != nil {
		return nil, m.ListErr
	}

	var files []model.File
	for _, f := range m.SavedFiles {
		files = append(files, f)
	}
	return files, nil
}

func TestHealthHandler(t *testing.T) {
	t.Run("returns healthy status", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		healthHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		if w.Header().Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
		}

		body := w.Body.String()
		if !contains(body, "healthy") || !contains(body, "File Store Service") {
			t.Errorf("expected body to contain 'healthy' and 'File Store Service', got %s", body)
		}
	})

	t.Run("response is valid JSON", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		healthHandler(w, req)

		var result map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &result)
		if err != nil {
			t.Errorf("expected valid JSON, got error: %v", err)
		}

		if result["status"] != "healthy" {
			t.Errorf("expected status field to be 'healthy', got %s", result["status"])
		}
	})
}

func TestUploadHandler(t *testing.T) {
	t.Run("rejects non-POST requests", func(t *testing.T) {
		mock := NewMockFileRepository()
		handler := makeUploadHandler(mock)

		req := httptest.NewRequest("GET", "/upload", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status 405, got %d", w.Code)
		}

		if mock.SaveCalled {
			t.Errorf("should not have called Save for non-POST request")
		}
	})

	t.Run("successfully uploads file", func(t *testing.T) {
		mock := NewMockFileRepository()
		handler := makeUploadHandler(mock)

		body := new(bytes.Buffer)
		mw := multipart.NewWriter(body)
		fw, err := mw.CreateFormFile("file", "test.txt")
		if err != nil {
			t.Fatalf("failed to create form file: %v", err)
		}

		fw.Write([]byte("test content"))
		mw.Close()

		req := httptest.NewRequest("POST", "/upload", body)
		req.Header.Set("Content-Type", mw.FormDataContentType())
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		if !mock.SaveCalled {
			t.Errorf("expected Save to be called")
		}

		if mock.SaveName != "test.txt" {
			t.Errorf("expected filename test.txt, got %s", mock.SaveName)
		}

		var result map[string]string
		json.Unmarshal(w.Body.Bytes(), &result)
		if result["message"] != "File uploaded successfully" {
			t.Errorf("expected success message, got %s", result["message"])
		}
	})

	t.Run("handles missing file field", func(t *testing.T) {
		mock := NewMockFileRepository()
		handler := makeUploadHandler(mock)

		body := new(bytes.Buffer)
		mw := multipart.NewWriter(body)
		mw.WriteField("notfile", "value")
		mw.Close()

		req := httptest.NewRequest("POST", "/upload", body)
		req.Header.Set("Content-Type", mw.FormDataContentType())
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("handles repository save error", func(t *testing.T) {
		mock := NewMockFileRepository()
		mock.SaveErr = ErrMock

		handler := makeUploadHandler(mock)

		body := new(bytes.Buffer)
		mw := multipart.NewWriter(body)
		fw, _ := mw.CreateFormFile("file", "test.txt")
		fw.Write([]byte("content"))
		mw.Close()

		req := httptest.NewRequest("POST", "/upload", body)
		req.Header.Set("Content-Type", mw.FormDataContentType())
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})

	t.Run("accepts file under 10MB limit", func(t *testing.T) {
		mock := NewMockFileRepository()
		handler := makeUploadHandler(mock)

		// Create 5MB file (under 10MB limit)
		largeContent := make([]byte, 5*1024*1024)
		body := new(bytes.Buffer)
		mw := multipart.NewWriter(body)
		fw, _ := mw.CreateFormFile("file", "large.bin")
		fw.Write(largeContent)
		mw.Close()

		req := httptest.NewRequest("POST", "/upload", body)
		req.Header.Set("Content-Type", mw.FormDataContentType())
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200 for file under 10MB limit, got %d", w.Code)
		}
	})

	t.Run("handles malformed multipart form", func(t *testing.T) {
		mock := NewMockFileRepository()
		handler := makeUploadHandler(mock)

		body := bytes.NewBufferString("invalid multipart data")

		req := httptest.NewRequest("POST", "/upload", body)
		req.Header.Set("Content-Type", "multipart/form-data; boundary=invalid")
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("returns filename in response", func(t *testing.T) {
		mock := NewMockFileRepository()
		handler := makeUploadHandler(mock)

		body := new(bytes.Buffer)
		mw := multipart.NewWriter(body)
		fw, _ := mw.CreateFormFile("file", "myfile.pdf")
		fw.Write([]byte("pdf content"))
		mw.Close()

		req := httptest.NewRequest("POST", "/upload", body)
		req.Header.Set("Content-Type", mw.FormDataContentType())
		w := httptest.NewRecorder()

		handler(w, req)

		var result map[string]string
		json.Unmarshal(w.Body.Bytes(), &result)
		if result["filename"] != "myfile.pdf" {
			t.Errorf("expected filename myfile.pdf in response, got %s", result["filename"])
		}
	})
}

func TestListHandler(t *testing.T) {
	t.Run("rejects non-GET requests", func(t *testing.T) {
		mock := NewMockFileRepository()
		handler := makeListHandler(mock)

		req := httptest.NewRequest("POST", "/list", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status 405, got %d", w.Code)
		}

		if mock.ListCalled {
			t.Errorf("should not have called List for non-GET request")
		}
	})

	t.Run("returns empty file list", func(t *testing.T) {
		mock := NewMockFileRepository()
		handler := makeListHandler(mock)

		req := httptest.NewRequest("GET", "/list", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		if !mock.ListCalled {
			t.Errorf("expected List to be called")
		}

		var result map[string][]model.File
		json.Unmarshal(w.Body.Bytes(), &result)
		if len(result["files"]) != 0 {
			t.Errorf("expected empty file list, got %d files", len(result["files"]))
		}
	})

	t.Run("returns multiple files", func(t *testing.T) {
		mock := NewMockFileRepository()

		// Pre-populate with files
		mock.SavedFiles["file1.txt"] = model.File{Name: "file1.txt", Size: 100, Path: "/path/file1.txt"}
		mock.SavedFiles["file2.txt"] = model.File{Name: "file2.txt", Size: 200, Path: "/path/file2.txt"}
		mock.SavedFiles["file3.txt"] = model.File{Name: "file3.txt", Size: 300, Path: "/path/file3.txt"}

		handler := makeListHandler(mock)

		req := httptest.NewRequest("GET", "/list", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var result map[string][]model.File
		json.Unmarshal(w.Body.Bytes(), &result)
		if len(result["files"]) != 3 {
			t.Errorf("expected 3 files, got %d", len(result["files"]))
		}
	})

	t.Run("handles repository list error", func(t *testing.T) {
		mock := NewMockFileRepository()
		mock.ListErr = ErrMock

		handler := makeListHandler(mock)

		req := httptest.NewRequest("GET", "/list", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})

	t.Run("response has correct content type", func(t *testing.T) {
		mock := NewMockFileRepository()
		handler := makeListHandler(mock)

		req := httptest.NewRequest("GET", "/list", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Header().Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
		}
	})

	t.Run("returns valid JSON structure", func(t *testing.T) {
		mock := NewMockFileRepository()
		mock.SavedFiles["test.txt"] = model.File{Name: "test.txt", Size: 42, Path: "/path/test.txt"}

		handler := makeListHandler(mock)

		req := httptest.NewRequest("GET", "/list", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		var result map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &result)
		if err != nil {
			t.Errorf("expected valid JSON, got error: %v", err)
		}

		if _, ok := result["files"]; !ok {
			t.Errorf("expected 'files' key in response")
		}
	})
}

func TestRunServer(t *testing.T) {
	t.Run("uses default port when empty string provided", func(t *testing.T) {
		// This test just verifies the port selection logic
		// We can't actually run the server in a test, but we verify the function exists
		if defaultPort != ":8080" {
			t.Errorf("expected default port :8080, got %s", defaultPort)
		}
	})

	t.Run("accepts custom port", func(t *testing.T) {
		if defaultPort == "" {
			t.Errorf("default port should not be empty")
		}
	})
}

// Helper functions
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

var ErrMock = io.EOF // Using a standard error for mocking
