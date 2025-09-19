package utils

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type BasicForm struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Address string `json:"address"`
}

type SliceForm struct {
	Tags     []string `json:"tags"`
	Name     string   `json:"name"`
	Keywords []string `json:"keywords"`
}

type MixedForm struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Categories  []string `json:"categories"`
	IsActive    string   `json:"is_active"`
}

func TestFormToStruct(t *testing.T) {
	tests := []struct {
		name     string
		formData url.Values
		target   any
		expected any
		wantErr  bool
	}{
		{
			name: "basic form with sanitization",
			formData: url.Values{
				"name":    []string{"  John Doe  "},
				"email":   []string{" john@example.com\n"},
				"address": []string{"\tMain Street\t\t123  "},
			},
			target: &BasicForm{},
			expected: &BasicForm{
				Name:    "John Doe",
				Email:   "john@example.com",
				Address: "Main Street 123",
			},
		},
		{
			name: "form with slice fields",
			formData: url.Values{
				"name":     []string{"  Product Name  "},
				"tags":     []string{" tag1 ", "\ttag2\n", "  tag3  "},
				"keywords": []string{"keyword1", "  keyword2  "},
			},
			target: &SliceForm{},
			expected: &SliceForm{
				Name:     "Product Name",
				Tags:     []string{"tag1", "tag2", "tag3"},
				Keywords: []string{"keyword1", "keyword2"},
			},
		},
		{
			name: "mixed form types",
			formData: url.Values{
				"title":       []string{"\n\nAwesome Title\n\n"},
				"description": []string{"  Multiple  \t  spaces   here  "},
				"categories":  []string{" cat1 ", "cat2\n", "\tcat3  "},
				"is_active":   []string{"true"},
			},
			target: &MixedForm{},
			expected: &MixedForm{
				Title:       "Awesome Title",
				Description: "Multiple spaces here",
				Categories:  []string{"cat1", "cat2", "cat3"},
				IsActive:    "true",
			},
		},
		{
			name: "empty and whitespace values",
			formData: url.Values{
				"name":    []string{""},
				"email":   []string{"   "},
				"address": []string{"\n\t\r"},
			},
			target: &BasicForm{},
			expected: &BasicForm{
				Name:    "",
				Email:   "",
				Address: "",
			},
		},
		{
			name: "single value treated as slice",
			formData: url.Values{
				"name": []string{"Single Name"},
				"tags": []string{"single-tag"},
			},
			target: &SliceForm{},
			expected: &SliceForm{
				Name: "Single Name",
				Tags: []string{"single-tag"},
			},
		},
		{
			name: "multiple values for non-slice field",
			formData: url.Values{
				"name":    []string{"First", "Second", "Third"},
				"email":   []string{"test@example.com"},
				"address": []string{"Address 1", "Address 2"},
			},
			target: &BasicForm{},
			expected: &BasicForm{
				Name:    "First",
				Email:   "test@example.com",
				Address: "Address 1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := tt.formData.Encode()
			req := &http.Request{
				Method: http.MethodPost,
				Header: make(http.Header),
				Body:   io.NopCloser(strings.NewReader(body)),
			}
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			err := FormToStruct(req, tt.target)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, tt.target)
			}
		})
	}
}

func TestFormToStructSanitization(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "trim whitespace",
			input:    "  hello world  ",
			expected: "hello world",
		},
		{
			name:     "normalize multiple spaces",
			input:    "hello    world",
			expected: "hello world",
		},
		{
			name:     "convert tabs to spaces",
			input:    "hello\tworld",
			expected: "hello world",
		},
		{
			name:     "convert newlines to spaces",
			input:    "hello\nworld\r\ntest",
			expected: "hello world test",
		},
		{
			name:     "complex whitespace normalization",
			input:    "\t  hello  \n\n  world  \t  ",
			expected: "hello world",
		},
		{
			name:     "empty after sanitization",
			input:    "   \n\t\r   ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formData := url.Values{
				"name": []string{tt.input},
			}

			body := formData.Encode()
			req := &http.Request{
				Method: http.MethodPost,
				Header: make(http.Header),
				Body:   io.NopCloser(strings.NewReader(body)),
			}
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			var result BasicForm
			err := FormToStruct(req, &result)
			require.NoError(t, err)
			require.Equal(t, tt.expected, result.Name)
		})
	}
}

func TestFormToStructErrors(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *http.Request
		target  any
		wantErr bool
	}{
		{
			name: "parse form error",
			setup: func() *http.Request {
				req := &http.Request{
					Method: http.MethodPost,
					Header: make(http.Header),
					Body:   io.NopCloser(strings.NewReader("invalid%form%data")),
				}
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				return req
			},
			target:  &BasicForm{},
			wantErr: true,
		},
		{
			name: "invalid target type",
			setup: func() *http.Request {
				formData := url.Values{"name": []string{"test"}}
				body := formData.Encode()
				req := &http.Request{
					Method: http.MethodPost,
					Header: make(http.Header),
					Body:   io.NopCloser(strings.NewReader(body)),
				}
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				return req
			},
			target:  "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.setup()
			err := FormToStruct(req, tt.target)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func BenchmarkFormToStruct(b *testing.B) {
	formData := url.Values{
		"title":       []string{"  Sample Title  "},
		"description": []string{" sample description\n"},
		"categories":  []string{"\tcat1\t", "cat2", "  cat3  "},
		"is_active":   []string{"true"},
	}

	body := formData.Encode()

	b.ResetTimer()
	for b.Loop() {
		req := &http.Request{
			Method: http.MethodPost,
			Header: make(http.Header),
			Body:   io.NopCloser(strings.NewReader(body)),
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		var result MixedForm
		_ = FormToStruct(req, &result)
	}
}

func BenchmarkSanitizeString(b *testing.B) {
	input := "\t  hello  \n\n  world  \t  with    multiple   spaces"

	b.ResetTimer()
	for b.Loop() {
		_ = SanitizeString(input)
	}
}
