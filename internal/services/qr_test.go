package services

import (
	"context"
	"testing"

	"github.com/VictoriaMetrics/fastcache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewQRService(t *testing.T) {
	cache := fastcache.New(1024 * 1024)
	svc := NewQRService(cache)

	assert.NotNil(t, svc)
	assert.NotNil(t, svc.cache)
}

func TestGenerateQR(t *testing.T) {
	cache := fastcache.New(1024 * 1024)
	svc := NewQRService(cache)
	ctx := context.Background()

	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "generates QR with URL",
			content: "http://localhost:8080/cpoints/redeem?code=CP-ABC-123-DEF",
			wantErr: false,
		},
		{
			name:    "generates QR with short content",
			content: "http://example.com",
			wantErr: false,
		},
		{
			name:    "generates QR with empty content",
			content: "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qrBytes, err := svc.GenerateQR(ctx, tt.content)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, qrBytes)
			assert.True(t, len(qrBytes) > 0, "QR bytes should not be empty")
		})
	}
}

func TestGenerateQR_Cache(t *testing.T) {
	cache := fastcache.New(1024 * 1024)
	svc := NewQRService(cache)
	ctx := context.Background()

	content := "http://localhost:8080/cpoints/redeem?code=CP-TEST-123-ABC"

	qrBytes1, err := svc.GenerateQR(ctx, content)
	require.NoError(t, err)

	qrBytes2, err := svc.GenerateQR(ctx, content)
	require.NoError(t, err)

	assert.Equal(t, qrBytes1, qrBytes2, "cached QR should be identical")
}

func TestGenerateQRBase64(t *testing.T) {
	cache := fastcache.New(1024 * 1024)
	svc := NewQRService(cache)
	ctx := context.Background()

	content := "http://localhost:8080/cpoints/redeem?code=CP-BASE-789-GHI"

	qrBase64, err := svc.GenerateQRBase64(ctx, content)
	require.NoError(t, err)
	assert.Contains(t, qrBase64, "data:image/png;base64,")

	expectedPrefix := "data:image/png;base64,"
	actualData := qrBase64[len(expectedPrefix):]
	assert.NotEmpty(t, actualData, "base64 data should not be empty")
}

func TestGenerateQR_DifferentContent(t *testing.T) {
	cache := fastcache.New(1024 * 1024)
	svc := NewQRService(cache)
	ctx := context.Background()

	qrBytes1, err := svc.GenerateQR(ctx, "http://example.com/path1")
	require.NoError(t, err)

	qrBytes2, err := svc.GenerateQR(ctx, "http://example.com/path2")
	require.NoError(t, err)

	assert.NotEqual(t, qrBytes1, qrBytes2, "different content should produce different QR")
}

func BenchmarkGenerateQR(b *testing.B) {
	cache := fastcache.New(1024 * 1024)
	svc := NewQRService(cache)
	ctx := context.Background()
	content := "http://localhost:8080/cpoints/redeem?code=CP-TEST-123-ABC"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.GenerateQR(ctx, content)
	}
}

func BenchmarkGenerateQR_Cached(b *testing.B) {
	cache := fastcache.New(1024 * 1024)
	svc := NewQRService(cache)
	ctx := context.Background()
	content := "http://localhost:8080/cpoints/redeem?code=CP-TEST-123-ABC"

	_, _ = svc.GenerateQR(ctx, content)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.GenerateQR(ctx, content)
	}
}
