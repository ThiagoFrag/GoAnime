package downloader

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeMovieDownloader returns a movie downloader pinned to a temp output dir
// so all path-relative tests are isolated.
func makeMovieDownloader(t *testing.T) *MovieDownloader {
	t.Helper()
	md := NewMovieDownloaderWithConfig(MovieDownloadConfig{
		OutputDir:    t.TempDir(),
		Quality:      scraper.Quality1080,
		SubsLanguage: "english",
		Provider:     "Vidcloud",
	})
	return md
}

// ---------------------------------------------------------------------------
// Constructors
// ---------------------------------------------------------------------------

func TestNewMovieDownloader_Defaults(t *testing.T) {
	t.Parallel()
	md := NewMovieDownloader()
	require.NotNil(t, md)
	assert.Equal(t, scraper.Quality1080, md.config.Quality)
	assert.Equal(t, "english", md.config.SubsLanguage)
	assert.Equal(t, "Vidcloud", md.config.Provider)
	assert.NotEmpty(t, md.config.OutputDir)
	assert.NotNil(t, md.mediaManager)
}

func TestNewMovieDownloaderWithConfig_FillsDefaults(t *testing.T) {
	t.Parallel()
	md := NewMovieDownloaderWithConfig(MovieDownloadConfig{OutputDir: "/tmp/movies"})
	require.NotNil(t, md)
	assert.Equal(t, "/tmp/movies", md.config.OutputDir)
	assert.Equal(t, scraper.Quality1080, md.config.Quality)
	assert.Equal(t, "english", md.config.SubsLanguage)
	assert.Equal(t, "Vidcloud", md.config.Provider)
}

func TestNewMovieDownloaderWithConfig_FullyEmptyConfig(t *testing.T) {
	t.Parallel()
	md := NewMovieDownloaderWithConfig(MovieDownloadConfig{})
	require.NotNil(t, md)
	assert.NotEmpty(t, md.config.OutputDir)
}

// ---------------------------------------------------------------------------
// Top-level public commands — nil-input & error paths
// ---------------------------------------------------------------------------

func TestDownloadMovie_NilMedia(t *testing.T) {
	t.Parallel()
	md := makeMovieDownloader(t)
	err := md.DownloadMovie(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "media is nil")
}

func TestDownloadMovie_BadMediaID(t *testing.T) {
	t.Parallel()
	md := makeMovieDownloader(t)
	// URL without dashes → extractMediaIDFromURL returns the whole URL but
	// only if there are no dashes — actually it returns last part of split.
	// Use an empty URL so the helper returns empty.
	err := md.DownloadMovie(&models.Anime{Name: "X", URL: ""})
	require.Error(t, err)
}

func TestDownloadTVEpisode_NilMedia(t *testing.T) {
	t.Parallel()
	md := makeMovieDownloader(t)
	err := md.DownloadTVEpisode(nil, 1, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "media is nil")
}

func TestDownloadTVEpisodeRange_InvertedRange(t *testing.T) {
	t.Parallel()
	md := makeMovieDownloader(t)
	err := md.DownloadTVEpisodeRange(&models.Anime{Name: "X", URL: "abc-123"}, 1, 9, 2)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be greater")
}

func TestDownloadAllSeasons_NilMedia(t *testing.T) {
	t.Parallel()
	md := makeMovieDownloader(t)
	err := md.DownloadAllSeasons(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "media is nil")
}

func TestDownloadAllSeasons_EmptyURL(t *testing.T) {
	t.Parallel()
	md := makeMovieDownloader(t)
	err := md.DownloadAllSeasons(&models.Anime{Name: "X", URL: ""})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "media ID")
}

// ---------------------------------------------------------------------------
// Stream fetch helpers — pinned (require live MediaManager scrapes)
// ---------------------------------------------------------------------------

func TestGetSFlixMovieStream_Pin(t *testing.T) {
	t.Parallel()
	md := makeMovieDownloader(t)
	_ = md.getSFlixMovieStream
}

func TestGetSFlixEpisodeStream_Pin(t *testing.T) {
	t.Parallel()
	md := makeMovieDownloader(t)
	_ = md.getSFlixEpisodeStream
}

// ---------------------------------------------------------------------------
// Download driver functions — pinned (drive bubbletea program)
// ---------------------------------------------------------------------------

func TestDownloadMovieWithProgress_Pin(t *testing.T) {
	t.Parallel()
	md := makeMovieDownloader(t)
	_ = md.downloadMovieWithProgress
}

func TestMovieDownloader_DownloadHTTPWithProgress_Pin(t *testing.T) {
	t.Parallel()
	md := makeMovieDownloader(t)
	_ = md.downloadHTTPWithProgress
}

func TestMovieDownloader_DownloadM3U8WithYtDlp_Pin(t *testing.T) {
	t.Parallel()
	md := makeMovieDownloader(t)
	_ = md.downloadM3U8WithYtDlp
}

func TestMovieDownloader_DownloadM3U8WithYtDlpDirect_Pin(t *testing.T) {
	t.Parallel()
	md := makeMovieDownloader(t)
	_ = md.downloadM3U8WithYtDlpDirect
}

func TestMovieDownloader_DownloadM3U8WithNativeHLS_Pin(t *testing.T) {
	t.Parallel()
	md := makeMovieDownloader(t)
	_ = md.downloadM3U8WithNativeHLS
}

// ---------------------------------------------------------------------------
// Pure helpers
// ---------------------------------------------------------------------------

func TestExtractRefererFromStreamURL(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"https_with_path", "https://megacloud.tv/embed-2/abc?k=v", "https://megacloud.tv/"},
		{"http_with_query", "http://foo.example/path?x=1", "http://foo.example/"},
		{"empty", "", ""},
		{"missing_scheme", "//example.com/x", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, extractRefererFromStreamURL(tc.in))
		})
	}
}

func TestMovieDownloader_GetContentLength(t *testing.T) {
	t.Parallel()
	md := makeMovieDownloader(t)

	t.Run("hls_estimate", func(t *testing.T) {
		t.Parallel()
		n, err := md.getContentLength("https://cdn.example.test/master.m3u8")
		require.NoError(t, err)
		assert.Equal(t, int64(500*1024*1024), n)
	})

	t.Run("head_failure_fallback", func(t *testing.T) {
		t.Parallel()
		// Unresolvable host → falls through to 500MB fallback OR error.
		// Current implementation returns error on HEAD failure for non-HLS URLs.
		_, err := md.getContentLength("https://nope.invalid.test/file.mp4")
		require.Error(t, err)
	})
}

func TestMovieDownloader_FileExists(t *testing.T) {
	t.Parallel()
	md := makeMovieDownloader(t)
	assert.False(t, md.fileExists(filepath.Join(md.config.OutputDir, "missing.mp4")))
	p := filepath.Join(md.config.OutputDir, "exists.mp4")
	require.NoError(t, writeTempFile(p, "x"))
	assert.True(t, md.fileExists(p))
}

func TestMovieDownloader_SanitizeDestPath(t *testing.T) {
	t.Parallel()
	md := makeMovieDownloader(t)

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		_, err := md.sanitizeDestPath("")
		require.Error(t, err)
	})

	t.Run("within", func(t *testing.T) {
		t.Parallel()
		p := filepath.Join(md.config.OutputDir, "Movie", "out.mp4")
		got, err := md.sanitizeDestPath(p)
		require.NoError(t, err)
		assert.Equal(t, p, got)
	})

	t.Run("escape", func(t *testing.T) {
		t.Parallel()
		_, err := md.sanitizeDestPath(filepath.Join(md.config.OutputDir, "..", "..", "etc"))
		require.Error(t, err)
	})
}

// ---------------------------------------------------------------------------
// TUI prompts (read stdin) — call with closed stdin so Scanln returns EOF.
// ---------------------------------------------------------------------------

func TestMovieDownloader_PromptPlayExisting(t *testing.T) {
	md := makeMovieDownloader(t)
	withClosedStdin(t)
	err := md.promptPlayExisting("/tmp/x", "Title")
	assert.NoError(t, err)
}

func TestMovieDownloader_PromptPlayDownloaded(t *testing.T) {
	md := makeMovieDownloader(t)
	withClosedStdin(t)
	err := md.promptPlayDownloaded("/tmp/x", "Title")
	assert.NoError(t, err)
}

func TestMovieDownloader_PlayMovie_Pin(t *testing.T) {
	t.Parallel()
	md := makeMovieDownloader(t)
	_ = md.playMovie // calls player.StartVideo → needs mpv
}

// ---------------------------------------------------------------------------
// Helper functions
// ---------------------------------------------------------------------------

func TestExtractMediaIDFromURL(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"flixhq_id", "https://flixhq.to/movie/watch-the-matrix-1234", "1234"},
		{"single_dash", "abc-99", "99"},
		{"no_dashes", "abcde", "abcde"},
		{"empty", "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, extractMediaIDFromURL(tc.in))
		})
	}
}

// ---------------------------------------------------------------------------
// Bubble Tea movieProgressModel
// ---------------------------------------------------------------------------

func TestMovieTickCmd_ReturnsMovieTickMsg(t *testing.T) {
	t.Parallel()
	cmd := movieTickCmd()
	require.NotNil(t, cmd)
	done := make(chan tea.Msg, 1)
	go func() { done <- cmd() }()
	select {
	case msg := <-done:
		_, ok := msg.(movieTickMsg)
		assert.True(t, ok, "expected movieTickMsg, got %T", msg)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("movieTickCmd timed out")
	}
}

func TestMovieProgressModel_Init(t *testing.T) {
	t.Parallel()
	m := &movieProgressModel{}
	assert.NotNil(t, m.Init())
}

func TestMovieProgressModel_Update(t *testing.T) {
	t.Parallel()

	t.Run("status_msg", func(t *testing.T) {
		t.Parallel()
		m := &movieProgressModel{progress: progress.New(progress.WithDefaultBlend())}
		_, cmd := m.Update(movieStatusMsg("downloading"))
		assert.Equal(t, "downloading", m.status)
		assert.Nil(t, cmd)
	})

	t.Run("progress_msg", func(t *testing.T) {
		t.Parallel()
		m := &movieProgressModel{progress: progress.New(progress.WithDefaultBlend())}
		_, _ = m.Update(movieProgressMsg{received: 30, totalBytes: 60, peakPct: 0})
		assert.Equal(t, int64(30), m.received)
		assert.Equal(t, int64(60), m.totalBytes)
		assert.Greater(t, m.peakPct, 0.0)
	})

	t.Run("tick_done_quits", func(t *testing.T) {
		t.Parallel()
		m := &movieProgressModel{progress: progress.New(progress.WithDefaultBlend()), done: true}
		_, cmd := m.Update(movieTickMsg(time.Now()))
		require.NotNil(t, cmd)
		_, ok := cmd().(tea.QuitMsg)
		assert.True(t, ok)
	})
}

func TestMovieProgressModel_View(t *testing.T) {
	t.Parallel()

	m := &movieProgressModel{progress: progress.New(progress.WithDefaultBlend())}
	v := m.View()
	assert.Contains(t, fmt.Sprintf("%v", v), "downloading")

	m2 := &movieProgressModel{progress: progress.New(progress.WithDefaultBlend()), title: "MyMovie", status: "Going"}
	v2 := m2.View()
	s2 := fmt.Sprintf("%v", v2)
	assert.Contains(t, s2, "MyMovie")
	assert.Contains(t, s2, "Going")
}

// ---------------------------------------------------------------------------
// Local helpers
// ---------------------------------------------------------------------------

func writeTempFile(path, content string) error {
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return err
		}
	}
	return os.WriteFile(path, []byte(content), 0o600)
}
