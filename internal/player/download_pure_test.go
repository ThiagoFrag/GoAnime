package player

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsAnimeFireVideoAPIURL(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{"animefire.io", "https://animefire.io/video/abc", true},
		{"animefire.plus", "https://animefire.plus/video/xyz", true},
		{"upper case", "https://ANIMEFIRE.IO/VIDEO/x", true},
		{"unrelated", "https://example.com/video/x", false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, isAnimeFireVideoAPIURL(tt.url))
		})
	}
}

func TestExtractRefererFromURL(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		url  string
		want string
	}{
		{"empty", "", ""},
		{"with path", "https://megacloud.tv/embed-2/abc?k=v", "https://megacloud.tv/"},
		{"scheme only no host", "http:///x", ""},
		{"http", "http://example.com/y", "http://example.com/"},
		{"bare path", "/just/a/path", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, extractRefererFromURL(tt.url))
		})
	}
}

func TestFileExists(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	t.Run("missing", func(t *testing.T) {
		t.Parallel()
		assert.False(t, fileExists(filepath.Join(dir, "missing")))
	})

	t.Run("present", func(t *testing.T) {
		t.Parallel()
		p := filepath.Join(dir, "x")
		require.NoError(t, os.WriteFile(p, []byte("y"), 0o600))
		assert.True(t, fileExists(p))
	})
}

func TestSafePartPath(t *testing.T) {
	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		got, err := safePartPath("/tmp/video.mp4", 1)
		require.NoError(t, err)
		assert.Equal(t, "/tmp/video.mp4.part1", got)
	})

	t.Run("subdir ok", func(t *testing.T) {
		t.Parallel()
		got, err := safePartPath("/tmp/sub/video.mp4", 7)
		require.NoError(t, err)
		assert.Contains(t, got, "video.mp4.part7")
	})
}

func TestIsBloggerProxyURL_Extra(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{"correct proxy", "http://127.0.0.1:8080/blogger_proxy/abc", true},
		{"loopback no token", "http://127.0.0.1:8080/", false},
		{"token no loopback", "http://example.com/blogger_proxy/", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, isBloggerProxyURL(tt.url))
		})
	}
}

// Cover the bare error-text branches of isUnsafeExtensionError & isRetryableError
// that the existing tests do not hit.
func TestIsUnsafeExtensionError_Extra(t *testing.T) {
	t.Parallel()
	assert.False(t, isUnsafeExtensionError(nil))
	assert.True(t, isUnsafeExtensionError(errors.New("file has unsafe extension")))
	assert.True(t, isUnsafeExtensionError(errors.New("file is unusual and will be skipped")))
	assert.False(t, isUnsafeExtensionError(errors.New("other failure")))
}

func TestIsRetryableError_Extra(t *testing.T) {
	t.Parallel()
	assert.False(t, isRetryableError(nil))
	assert.True(t, isRetryableError(errors.New("connection reset")))
	assert.True(t, isRetryableError(errors.New("network down")))
	assert.True(t, isRetryableError(errors.New("connection refused")))
	assert.True(t, isRetryableError(errors.New("temporary error")))
	assert.False(t, isRetryableError(errors.New("auth denied")))
}
