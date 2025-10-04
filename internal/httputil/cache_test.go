package httputil

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

// mockFileInfo implements os.FileInfo for testing
type mockFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

func (m mockFileInfo) Name() string       { return m.name }
func (m mockFileInfo) Size() int64        { return m.size }
func (m mockFileInfo) Mode() os.FileMode  { return m.mode }
func (m mockFileInfo) ModTime() time.Time { return m.modTime }
func (m mockFileInfo) IsDir() bool        { return m.isDir }
func (m mockFileInfo) Sys() any           { return nil }

// mockFile implements http.File for testing
type mockFile struct {
	*strings.Reader
	info os.FileInfo
}

func (m *mockFile) Close() error                             { return nil }
func (m *mockFile) Readdir(count int) ([]os.FileInfo, error) { return nil, nil }
func (m *mockFile) Stat() (os.FileInfo, error)               { return m.info, nil }

func newMockFile(name string, size int64, modTime time.Time) *mockFile {
	return &mockFile{
		Reader: strings.NewReader("mock file content"),
		info: mockFileInfo{
			name:    name,
			size:    size,
			mode:    0644,
			modTime: modTime,
			isDir:   false,
		},
	}
}

// mockFileSystem implements http.FileSystem for testing
type mockFileSystem struct {
	files map[string]*mockFile
	err   error
}

func (m *mockFileSystem) Open(name string) (http.File, error) {
	if m.err != nil {
		return nil, m.err
	}
	if file, exists := m.files[name]; exists {
		return file, nil
	}
	return nil, os.ErrNotExist
}

func newMockFileSystem() *mockFileSystem {
	return &mockFileSystem{
		files: make(map[string]*mockFile),
	}
}

func (m *mockFileSystem) addFile(path string, size int64, modTime time.Time) {
	m.files[path] = newMockFile(path, size, modTime)
}

func TestParseAndSortQuery(t *testing.T) {
	tests := []struct {
		name     string
		rawQuery string
		expected string
	}{
		{
			name:     "empty query",
			rawQuery: "",
			expected: "",
		},
		{
			name:     "single parameter",
			rawQuery: "param=value",
			expected: "param=value",
		},
		{
			name:     "multiple parameters sorted",
			rawQuery: "z=1&a=2&m=3",
			expected: "a=2&m=3&z=1",
		},
		{
			name:     "parameters with special characters",
			rawQuery: "search=hello%20world&sort=asc&filter=active",
			expected: "filter=active&search=hello%20world&sort=asc",
		},
		{
			name:     "duplicate parameters",
			rawQuery: "a=1&b=2&a=3",
			expected: "a=1&a=3&b=2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseAndSortQuery(tt.rawQuery)
			if result != tt.expected {
				t.Errorf("parseAndSortQuery(%q) = %q, want %q", tt.rawQuery, result, tt.expected)
			}
		})
	}
}

func TestGenerateETag(t *testing.T) {
	fixedTime := time.Date(2023, 10, 15, 12, 0, 0, 0, time.UTC)
	fileInfo := mockFileInfo{
		name:    "test.jpg",
		size:    1024,
		modTime: fixedTime,
	}

	tests := []struct {
		name         string
		info         os.FileInfo
		rawQuery     string
		expectPrefix string
	}{
		{
			name:         "no query parameters",
			info:         fileInfo,
			rawQuery:     "",
			expectPrefix: `"f`,
		},
		{
			name:         "with query parameters",
			info:         fileInfo,
			rawQuery:     "width=100&height=200",
			expectPrefix: `"f`,
		},
		{
			name:         "query parameters in different order should generate same etag",
			info:         fileInfo,
			rawQuery:     "height=200&width=100",
			expectPrefix: `"f`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			etag, err := generateETag(tt.info, tt.rawQuery)
			if err != nil {
				t.Errorf("generateETag() error = %v", err)
				return
			}
			if !strings.HasPrefix(etag, tt.expectPrefix) {
				t.Errorf("generateETag() = %q, want prefix %q", etag, tt.expectPrefix)
			}
			if !strings.HasSuffix(etag, `"`) {
				t.Errorf("generateETag() = %q, want suffix '\"'", etag)
			}
		})
	}

	// Test that same parameters in different order generate same ETag
	t.Run("query parameter order independence", func(t *testing.T) {
		etag1, err1 := generateETag(fileInfo, "width=100&height=200")
		etag2, err2 := generateETag(fileInfo, "height=200&width=100")

		if err1 != nil || err2 != nil {
			t.Errorf("generateETag() errors: %v, %v", err1, err2)
			return
		}
		if etag1 != etag2 {
			t.Errorf("generateETag() with reordered params: %q != %q", etag1, etag2)
		}
	})
}

func TestSetCacheControlHeaders(t *testing.T) {
	tests := []struct {
		name                 string
		path                 string
		query                string
		expectedCacheControl string
		expectVary           bool
	}{
		{
			name:                 "static file path",
			path:                 "/static/images/logo.png",
			query:                "",
			expectedCacheControl: "public, max-age=86400",
			expectVary:           false,
		},
		{
			name:                 "non-static path",
			path:                 "/api/products",
			query:                "",
			expectedCacheControl: "public, max-age=3600, stale-while-revalidate=86400",
			expectVary:           false,
		},
		{
			name:                 "static path with query",
			path:                 "/static/images/photo.jpg",
			query:                "width=100",
			expectedCacheControl: "public, max-age=86400",
			expectVary:           true,
		},
		{
			name:                 "non-static path with query",
			path:                 "/home",
			query:                "page=2",
			expectedCacheControl: "public, max-age=3600, stale-while-revalidate=86400",
			expectVary:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tt.path+"?"+tt.query, nil)

			setCacheControlHeaders(w, req)

			cacheControl := w.Header().Get("Cache-Control")
			if cacheControl != tt.expectedCacheControl {
				t.Errorf("setCacheControlHeaders() Cache-Control = %q, want %q", cacheControl, tt.expectedCacheControl)
			}

			vary := w.Header().Get("Vary")
			hasVary := vary != ""
			if hasVary != tt.expectVary {
				t.Errorf("setCacheControlHeaders() Vary header presence = %v, want %v", hasVary, tt.expectVary)
			}

			if tt.expectVary && vary != "Accept, Accept-Encoding" {
				t.Errorf("setCacheControlHeaders() Vary = %q, want %q", vary, "Accept, Accept-Encoding")
			}
		})
	}
}

func TestCacheHeaders(t *testing.T) {
	fixedTime := time.Date(2023, 10, 15, 12, 0, 0, 0, time.UTC)

	t.Run("successful cache miss", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/static/test.jpg", nil)

		fs := newMockFileSystem()
		fs.addFile("/test.jpg", 1024, fixedTime)

		cached, file, err := CacheHeaders(w, req, fs, "static/test.jpg")

		if err != nil {
			t.Errorf("CacheHeaders() error = %v", err)
		}
		if cached {
			t.Errorf("CacheHeaders() cached = %v, want false", cached)
		}
		if file == nil {
			t.Errorf("CacheHeaders() file = nil, want non-nil")
		}

		// Check headers are set
		if w.Header().Get("ETag") == "" {
			t.Errorf("CacheHeaders() ETag header not set")
		}
		if w.Header().Get("Last-Modified") == "" {
			t.Errorf("CacheHeaders() Last-Modified header not set")
		}
		if w.Header().Get("Cache-Control") == "" {
			t.Errorf("CacheHeaders() Cache-Control header not set")
		}

		file.Close()
	})

	t.Run("cache hit with If-None-Match", func(t *testing.T) {
		w := httptest.NewRecorder()
		fs := newMockFileSystem()
		fs.addFile("/test.jpg", 1024, fixedTime)

		// First call to get the ETag
		req1 := httptest.NewRequest(http.MethodGet, "/static/test.jpg", nil)
		_, file1, _ := CacheHeaders(w, req1, fs, "static/test.jpg")
		etag := w.Header().Get("ETag")
		file1.Close()

		// Second call with If-None-Match header
		w2 := httptest.NewRecorder()
		req2 := httptest.NewRequest(http.MethodGet, "/static/test.jpg", nil)
		req2.Header.Set("If-None-Match", etag)

		cached, file2, err := CacheHeaders(w2, req2, fs, "static/test.jpg")

		if err != nil {
			t.Errorf("CacheHeaders() error = %v", err)
		}
		if !cached {
			t.Errorf("CacheHeaders() cached = %v, want true", cached)
		}
		if w2.Code != http.StatusNotModified {
			t.Errorf("CacheHeaders() status = %v, want %v", w2.Code, http.StatusNotModified)
		}

		file2.Close()
	})

	t.Run("cache hit with If-Modified-Since", func(t *testing.T) {
		w := httptest.NewRecorder()
		fs := newMockFileSystem()
		fs.addFile("/test.jpg", 1024, fixedTime)

		// Request with If-Modified-Since set to future time
		req := httptest.NewRequest(http.MethodGet, "/static/test.jpg", nil)
		futureTime := fixedTime.Add(2 * time.Hour)
		req.Header.Set("If-Modified-Since", futureTime.UTC().Format(http.TimeFormat))

		cached, file, err := CacheHeaders(w, req, fs, "static/test.jpg")

		if err != nil {
			t.Errorf("CacheHeaders() error = %v", err)
		}
		if !cached {
			t.Errorf("CacheHeaders() cached = %v, want true", cached)
		}
		if w.Code != http.StatusNotModified {
			t.Errorf("CacheHeaders() status = %v, want %v", w.Code, http.StatusNotModified)
		}

		file.Close()
	})

	t.Run("file not found error", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/static/notfound.jpg", nil)

		fs := newMockFileSystem()

		cached, file, err := CacheHeaders(w, req, fs, "static/notfound.jpg")

		if err == nil {
			t.Errorf("CacheHeaders() error = nil, want non-nil")
		}
		if cached {
			t.Errorf("CacheHeaders() cached = %v, want false", cached)
		}
		if file != nil {
			t.Errorf("CacheHeaders() file = %v, want nil", file)
		}
	})

	t.Run("path trimming", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/static/test.jpg", nil)

		fs := newMockFileSystem()
		fs.addFile("/test.jpg", 1024, fixedTime)

		cached, file, err := CacheHeaders(w, req, fs, "static/test.jpg")

		if err != nil {
			t.Errorf("CacheHeaders() error = %v", err)
		}
		if cached {
			t.Errorf("CacheHeaders() cached = %v, want false", cached)
		}
		if file == nil {
			t.Errorf("CacheHeaders() file = nil, want non-nil")
		}

		file.Close()
	})

	t.Run("nil filesystem uses os.Open", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/static/nonexistent.jpg", nil)

		cached, file, err := CacheHeaders(w, req, nil, "static/nonexistent.jpg")

		// Should return an error since the file doesn't exist
		if err == nil {
			t.Errorf("CacheHeaders() with nil fs should return error for nonexistent file")
		}
		if cached {
			t.Errorf("CacheHeaders() cached = %v, want false", cached)
		}
		if file != nil {
			t.Errorf("CacheHeaders() file = %v, want nil", file)
		}
	})

	t.Run("with query parameters", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/static/test.jpg?width=100&height=200", nil)

		fs := newMockFileSystem()
		fs.addFile("/test.jpg", 1024, fixedTime)

		cached, file, err := CacheHeaders(w, req, fs, "static/test.jpg")

		if err != nil {
			t.Errorf("CacheHeaders() error = %v", err)
		}
		if cached {
			t.Errorf("CacheHeaders() cached = %v, want false", cached)
		}

		// Verify Vary header is set for requests with query params
		vary := w.Header().Get("Vary")
		if vary != "Accept, Accept-Encoding" {
			t.Errorf("CacheHeaders() Vary = %q, want %q", vary, "Accept, Accept-Encoding")
		}

		file.Close()
	})
}

// BenchmarkGenerateETag benchmarks the ETag generation function
func BenchmarkGenerateETag(b *testing.B) {
	fileInfo := mockFileInfo{
		name:    "test.jpg",
		size:    1024,
		modTime: time.Now(),
	}

	b.ResetTimer()
	for b.Loop() {
		if _, err := generateETag(fileInfo, "width=100&height=200&quality=80"); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParseAndSortQuery benchmarks the query parsing function
func BenchmarkParseAndSortQuery(b *testing.B) {
	query := "z=1&a=2&m=3&width=100&height=200&quality=80&format=webp"

	b.ResetTimer()
	for b.Loop() {
		parseAndSortQuery(query)
	}
}
