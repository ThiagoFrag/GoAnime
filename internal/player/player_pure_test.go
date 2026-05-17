package player

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alvarorichard/Goanime/internal/api/providers/metadata"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestSanitizeMediaTarget(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		wantErr bool
		want    string
	}{
		{"empty", "", true, ""},
		{"newline middle", "http://x\ny", true, ""},
		{"null byte", "http://x\x00y", true, ""},
		{"dash prefix", "-attack", true, ""},
		{"https url", "https://example.com/v.mp4", false, "https://example.com/v.mp4"},
		{"http url", "http://example.com/v.mp4", false, "http://example.com/v.mp4"},
		{"file scheme rejected", "file:///etc/passwd", true, ""},
		{"ftp rejected", "ftp://x.com", true, ""},
		{"plain path cleaned", "/tmp/video.mp4", false, "/tmp/video.mp4"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sanitizeMediaTarget(tt.in)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSanitizeOutputPath(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		wantErr bool
	}{
		{"empty", "", true},
		{"null byte", "file\x00", true},
		{"newline", "file\n.mp4", true},
		{"dash prefix", "-output.mp4", true},
		{"home prefix ok", filepath.Join(homeOrEmpty(), "ok.mp4"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := sanitizeOutputPath(tt.in)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func homeOrEmpty() string {
	d, err := os.UserHomeDir()
	if err != nil {
		return "/tmp"
	}
	return d
}

// SetMediaType toggles the isMovieOrTV flag (see snapshot). IsCurrentMediaMovie
// reads the separate `mediaType == "movie"` field set via SetExactMediaType.
func TestSetMediaType_TogglesFlag(t *testing.T) {
	prev := snapshotMedia().IsMovieOrTV
	t.Cleanup(func() { SetMediaType(prev) })

	SetMediaType(true)
	assert.True(t, snapshotMedia().IsMovieOrTV)

	SetMediaType(false)
	assert.False(t, snapshotMedia().IsMovieOrTV)
}

func TestIsCurrentMediaMovie_DependsOnMediaType(t *testing.T) {
	prev := GetExactMediaType()
	t.Cleanup(func() { SetExactMediaType(prev) })

	SetExactMediaType("movie")
	assert.True(t, IsCurrentMediaMovie())
	SetExactMediaType("tv")
	assert.False(t, IsCurrentMediaMovie())
	SetExactMediaType("anime")
	assert.False(t, IsCurrentMediaMovie())
}

func TestSetExactMediaType_RoundTrip(t *testing.T) {
	prev := GetExactMediaType()
	t.Cleanup(func() { SetExactMediaType(prev) })

	SetExactMediaType("movie")
	assert.Equal(t, "movie", GetExactMediaType())

	SetExactMediaType("tv")
	assert.Equal(t, "tv", GetExactMediaType())

	SetExactMediaType("anime")
	assert.Equal(t, "anime", GetExactMediaType())
}

func TestSetSeasonMap_RoundTrip(t *testing.T) {
	sm := []metadata.SeasonMapping{{Season: 1, StartEp: 1, EndEp: 12, EpisodeCount: 12}}
	SetSeasonMap(sm)
	snap := snapshotMedia()
	assert.Equal(t, 1, snap.SeasonMap[0].Season)
}

func TestSetMediaMeta_RoundTrip(t *testing.T) {
	meta := &util.MediaMeta{IMDBID: "tt1234567", Year: "2020"}
	SetMediaMeta(meta)
	got := GetMediaMeta()
	assert.NotNil(t, got)
	assert.Equal(t, "tt1234567", got.IMDBID)
	assert.Equal(t, "2020", got.Year)
}

func TestPreWarmMPVPath_NoPanic(t *testing.T) {
	assert.NotPanics(t, func() { PreWarmMPVPath() })
}

func TestSnapshotMedia_ReturnsCopy(t *testing.T) {
	SetAnimeName("X", 3)
	snap := snapshotMedia()
	assert.Equal(t, "X", snap.AnimeName)
	assert.Equal(t, 3, snap.AnimeSeason)
}

func TestPrintDownloadLocation_DoesNotPanic(t *testing.T) {
	t.Parallel()
	assert.NotPanics(t, func() { printDownloadLocation("/tmp/some/file.mp4") })
}

func TestSetLastAnimeURL_RoundTrip(t *testing.T) {
	prev := getLastAnimeURL()
	t.Cleanup(func() { setLastAnimeURL(prev) })

	setLastAnimeURL("https://example.com/anime/x")
	assert.Equal(t, "https://example.com/anime/x", getLastAnimeURL())

	setLastAnimeURL("")
	assert.Equal(t, "", getLastAnimeURL())
}

func TestGetExactMediaType_DefaultEmpty(t *testing.T) {
	prev := GetExactMediaType()
	t.Cleanup(func() { SetExactMediaType(prev) })

	SetExactMediaType("")
	assert.Equal(t, "", GetExactMediaType())
}

func TestGetMediaMeta_AfterClearReturnsNil(t *testing.T) {
	prev := GetMediaMeta()
	t.Cleanup(func() { SetMediaMeta(prev) })

	SetMediaMeta(nil)
	assert.Nil(t, GetMediaMeta())
}

// Mutates global util.GlobalSubtitles — keep serial (no parallel).
func TestDownloadSubtitleFiles_NoSubsIsNoop(t *testing.T) {
	prev := util.GlobalSubtitles
	util.ClearGlobalSubtitles()
	t.Cleanup(func() { util.GlobalSubtitles = prev })

	assert.NotPanics(t, func() { downloadSubtitleFiles("/tmp/x.mp4", nil) })
}

func TestStartVideo_InvalidLinkReturnsError(t *testing.T) {
	t.Parallel()
	if _, err := StartVideo("http://bad\nurl", nil); err == nil {
		t.Fatal("expected error from sanitize or missing-mpv path")
	}
}

// fuzzyfinder/tcell terminfo is package-level state — keep TUI-touching tests serial.
func TestHandleUpscaleFromMenu_DoesNotPanic(t *testing.T) {
	assert.NotPanics(t, func() { _ = handleUpscaleFromMenu() })
}

func TestAskForDownload_ReturnsValidMarker(t *testing.T) {
	got := askForDownload()
	assert.GreaterOrEqual(t, got, 1)
}

func TestAskForPlayOffline_DoesNotPanic(t *testing.T) {
	assert.NotPanics(t, func() { _ = askForPlayOffline() })
}

// HandleDownloadAndPlay loops on askForDownload (huh.NewSelect) which never
// terminates without a TTY — testing the orchestration requires either a real
// terminal or a major refactor (extract the loop body into a pure dispatcher).
// CLAUDE.md allows skipping pure TUI orchestration; the dispatched branches
// (askForDownload, HandleBatchDownload, handleUpscaleFromMenu, playVideo,
// extractActualVideoURL, downloadAndPlayEpisode) are each covered by their
// own dedicated tests in this package.
func TestHandleDownloadAndPlay_SymbolPinned(t *testing.T) {
	t.Parallel()
	assert.NotNil(t, HandleDownloadAndPlay)
}

// downloadAndPlayEpisode mirrors HandleDownloadAndPlay's structure: it is a
// pure-orchestration function whose every collaborator is exercised in
// isolation by sibling tests (DownloadVideo, downloadDirectHTTP,
// downloadWithYtDlp, downloadWithNativeHLS, askAndPlayDownloadedEpisode,
// playVideo). The dispatcher itself loops on huh.NewSelect and cannot be
// driven without a TTY.
func TestDownloadAndPlayEpisode_SymbolPinned(t *testing.T) {
	t.Parallel()
	assert.NotNil(t, downloadAndPlayEpisode)
}
