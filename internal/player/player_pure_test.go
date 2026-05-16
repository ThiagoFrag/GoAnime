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
