package downloader

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeNineAnimeDownloader returns a 9anime downloader with a temp output dir.
func makeNineAnimeDownloader(t *testing.T) *NineAnimeDownloader {
	t.Helper()
	return NewNineAnimeDownloader(NineAnimeDownloadConfig{
		OutputDir: t.TempDir(),
		AnimeName: "TestAnime",
		Season:    1,
	})
}

// ---------------------------------------------------------------------------
// Constructor
// ---------------------------------------------------------------------------

func TestNewNineAnimeDownloader_Defaults(t *testing.T) {
	t.Parallel()
	d := NewNineAnimeDownloader(NineAnimeDownloadConfig{})
	require.NotNil(t, d)
	assert.NotEmpty(t, d.config.OutputDir)
	assert.Equal(t, "sub", d.config.AudioType)
	assert.Equal(t, "best", d.config.Quality)
	assert.Equal(t, 1, d.config.Season)
	assert.GreaterOrEqual(t, d.config.Concurrent, 1)
	assert.NotNil(t, d.client)
}

func TestNewNineAnimeDownloader_RespectsExplicit(t *testing.T) {
	t.Parallel()
	d := NewNineAnimeDownloader(NineAnimeDownloadConfig{
		OutputDir:  "/tmp/9a",
		AudioType:  "dub",
		Quality:    "1080p",
		Season:     3,
		Concurrent: 4,
	})
	assert.Equal(t, "/tmp/9a", d.config.OutputDir)
	assert.Equal(t, "dub", d.config.AudioType)
	assert.Equal(t, "1080p", d.config.Quality)
	assert.Equal(t, 3, d.config.Season)
	assert.Equal(t, 4, d.config.Concurrent)
}

// ---------------------------------------------------------------------------
// Public commands (nil + network error paths — client.GetEpisodes hits 9anime)
// ---------------------------------------------------------------------------

func TestNineAnimeDownloader_DownloadAllEpisodes_NilAnime(t *testing.T) {
	t.Parallel()
	d := makeNineAnimeDownloader(t)
	err := d.DownloadAllEpisodes(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "anime is nil")
}

func TestNineAnimeDownloader_DownloadSingleEpisode_NilAnime(t *testing.T) {
	t.Parallel()
	d := makeNineAnimeDownloader(t)
	err := d.DownloadSingleEpisode(nil, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "anime is nil")
}

func TestNineAnimeDownloader_DownloadEpisodeRange_NilAnime(t *testing.T) {
	t.Parallel()
	d := makeNineAnimeDownloader(t)
	err := d.DownloadEpisodeRange(nil, 1, 5)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "anime is nil")
}

func TestNineAnimeDownloader_DownloadEpisodeRange_InvertedRange(t *testing.T) {
	t.Parallel()
	d := makeNineAnimeDownloader(t)
	err := d.DownloadEpisodeRange(&models.Anime{Name: "X", URL: "abc"}, 9, 2)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be greater")
}

// ---------------------------------------------------------------------------
// Path-building helpers
// ---------------------------------------------------------------------------

func TestNineAnimeDownloader_BuildOutputDir(t *testing.T) {
	t.Parallel()
	d := makeNineAnimeDownloader(t)

	anime := &models.Anime{Name: "Naruto", Year: "2002"}
	got := d.buildOutputDir(anime)
	assert.Contains(t, got, "Naruto")
	assert.Contains(t, got, "Season 01")
	assert.True(t, strings.HasPrefix(got, d.config.OutputDir))
}

func TestNineAnimeDownloader_BuildOutputDir_FallbackName(t *testing.T) {
	t.Parallel()
	d := NewNineAnimeDownloader(NineAnimeDownloadConfig{OutputDir: t.TempDir(), Season: 2})
	got := d.buildOutputDir(&models.Anime{Name: "Bleach"})
	assert.Contains(t, got, "Bleach")
	assert.Contains(t, got, "Season 02")
}

func TestNineAnimeDownloader_EpisodeFilename(t *testing.T) {
	t.Parallel()
	d := makeNineAnimeDownloader(t)
	got := d.episodeFilename(&models.Anime{Name: "Naruto"}, 7)
	assert.Contains(t, got, "S01E07")
	assert.Contains(t, got, "TestAnime")
}

func TestNineAnimeDownloader_EpisodeFilename_EmptyNameFallback(t *testing.T) {
	t.Parallel()
	d := NewNineAnimeDownloader(NineAnimeDownloadConfig{OutputDir: t.TempDir(), Season: 1})
	// AnimeName empty + anime.Name empty → uses "9Anime_<URL>" fallback
	got := d.episodeFilename(&models.Anime{Name: "", URL: "abc123"}, 4)
	assert.Contains(t, got, "9Anime_abc123")
}

// ---------------------------------------------------------------------------
// Stream resolution — pinned (requires live 9anime endpoint)
// ---------------------------------------------------------------------------

func TestNineAnimeDownloader_ResolveStream_Pin(t *testing.T) {
	t.Parallel()
	d := makeNineAnimeDownloader(t)
	_ = d.resolveStream
}

// ---------------------------------------------------------------------------
// Subtitle prompt — exercise pre-configured branches
// ---------------------------------------------------------------------------

func TestNineAnimeDownloader_PromptSubtitleLanguage(t *testing.T) {
	t.Run("none_preconfigured", func(t *testing.T) {
		t.Parallel()
		d := NewNineAnimeDownloader(NineAnimeDownloadConfig{OutputDir: t.TempDir(), SubsLanguage: "none"})
		got := d.promptSubtitleLanguage([]scraper.NineAnimeSubtitleTrack{{Label: "English"}})
		assert.Equal(t, "", got)
		assert.True(t, d.subLangResolved)
	})

	t.Run("all_preconfigured", func(t *testing.T) {
		t.Parallel()
		d := NewNineAnimeDownloader(NineAnimeDownloadConfig{OutputDir: t.TempDir(), SubsLanguage: "all"})
		got := d.promptSubtitleLanguage([]scraper.NineAnimeSubtitleTrack{{Label: "English"}})
		assert.Equal(t, "__all__", got)
	})

	t.Run("exact_match", func(t *testing.T) {
		t.Parallel()
		d := NewNineAnimeDownloader(NineAnimeDownloadConfig{OutputDir: t.TempDir(), SubsLanguage: "Portuguese"})
		got := d.promptSubtitleLanguage([]scraper.NineAnimeSubtitleTrack{
			{Label: "English"},
			{Label: "Portuguese"},
		})
		assert.Equal(t, "Portuguese", got)
	})

	t.Run("cached_returns_immediately", func(t *testing.T) {
		t.Parallel()
		d := makeNineAnimeDownloader(t)
		d.subLangResolved = true
		d.chosenSubLang = "Spanish"
		got := d.promptSubtitleLanguage([]scraper.NineAnimeSubtitleTrack{{Label: "English"}})
		assert.Equal(t, "Spanish", got)
	})

	t.Run("empty_tracks", func(t *testing.T) {
		t.Parallel()
		d := makeNineAnimeDownloader(t)
		got := d.promptSubtitleLanguage(nil)
		assert.Equal(t, "", got)
	})
}

// ---------------------------------------------------------------------------
// Subtitle muxing — pinned (requires ffmpeg)
// ---------------------------------------------------------------------------

func TestNineAnimeDownloader_DownloadSubtitles_NoTracks(t *testing.T) {
	t.Parallel()
	d := makeNineAnimeDownloader(t)
	// Empty tracks → early return without any side effects.
	assert.NotPanics(t, func() {
		d.downloadSubtitles(nil, filepath.Join(d.config.OutputDir, "x.mp4"))
	})
}

// ---------------------------------------------------------------------------
// downloadFile — direct unit via httptest
// ---------------------------------------------------------------------------

func TestNineAnimeDownloader_DownloadFile_SuccessBlocked(t *testing.T) {
	t.Parallel()
	d := makeNineAnimeDownloader(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "subtitle-vtt-data")
	}))
	t.Cleanup(srv.Close)

	dest := filepath.Join(t.TempDir(), "sub.vtt")
	// SafeTransport blocks loopback → DownloadFile must surface that error.
	err := d.downloadFile(srv.URL+"/sub.vtt", dest)
	require.Error(t, err)
}

func TestNineAnimeDownloader_DownloadFile_BadURL(t *testing.T) {
	t.Parallel()
	d := makeNineAnimeDownloader(t)
	err := d.downloadFile("://invalid::", filepath.Join(t.TempDir(), "x"))
	require.Error(t, err)
}

func TestNineAnimeDownloader_DownloadFile_NonOK(t *testing.T) {
	t.Parallel()
	d := makeNineAnimeDownloader(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)
	err := d.downloadFile(srv.URL+"/missing.vtt", filepath.Join(t.TempDir(), "x.vtt"))
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// downloadEpisodeWithProgress — skip-existing branch
// ---------------------------------------------------------------------------

func TestNineAnimeDownloader_DownloadEpisodeWithProgress_SkipExisting(t *testing.T) {
	t.Parallel()
	d := makeNineAnimeDownloader(t)
	outDir := t.TempDir()
	ep := &scraper.NineAnimeEpisode{Number: 5, Title: "x", EpisodeID: "id"}

	target := filepath.Join(outDir, d.episodeFilename(&models.Anime{Name: "Anime"}, ep.Number))
	require.NoError(t, os.MkdirAll(filepath.Dir(target), 0o700))
	require.NoError(t, os.WriteFile(target, make([]byte, 2*1024*1024), 0o600))

	err := d.downloadEpisodeWithProgress(&models.Anime{Name: "Anime"}, ep, outDir)
	require.NoError(t, err, "existing episode must short-circuit without error")
}

// ---------------------------------------------------------------------------
// downloadBatchWithProgress — all-existing short circuit
// ---------------------------------------------------------------------------

func TestNineAnimeDownloader_DownloadBatchWithProgress_AllExisting(t *testing.T) {
	t.Parallel()
	d := makeNineAnimeDownloader(t)
	outDir := t.TempDir()
	episodes := []scraper.NineAnimeEpisode{{Number: 1, Title: "a"}, {Number: 2, Title: "b"}}
	a := &models.Anime{Name: "Anime"}
	for _, ep := range episodes {
		path := filepath.Join(outDir, d.episodeFilename(a, ep.Number))
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
		require.NoError(t, os.WriteFile(path, make([]byte, 2*1024*1024), 0o600))
	}
	err := d.downloadBatchWithProgress(a, episodes, outDir)
	require.NoError(t, err)
}

// ---------------------------------------------------------------------------
// downloadStream / downloadNativeHLS / downloadWithYtDlp — pinned
// ---------------------------------------------------------------------------

func TestNineAnimeDownloader_DownloadStream_Pin(t *testing.T) {
	t.Parallel()
	d := makeNineAnimeDownloader(t)
	_ = d.downloadStream
}

func TestNineAnimeDownloader_DownloadNativeHLS_Pin(t *testing.T) {
	t.Parallel()
	d := makeNineAnimeDownloader(t)
	_ = d.downloadNativeHLS
}

func TestNineAnimeDownloader_DownloadWithYtDlp_Pin(t *testing.T) {
	t.Parallel()
	d := makeNineAnimeDownloader(t)
	_ = d.downloadWithYtDlp
}

// ---------------------------------------------------------------------------
// Pure helpers
// ---------------------------------------------------------------------------

func TestIsRetryableDownloadError(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"timeout", errors.New("read timeout"), true},
		{"connection_reset", errors.New("connection reset by peer"), true},
		{"network_unreachable", errors.New("network is unreachable"), true},
		{"refused", errors.New("connection refused"), true},
		{"temporary", errors.New("temporary failure"), true},
		{"unrelated", errors.New("404 not found"), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, isRetryableDownloadError(tc.err))
		})
	}
}
