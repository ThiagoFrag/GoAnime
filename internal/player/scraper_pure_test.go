package player

import (
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestAbs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in, want int
	}{
		{0, 0},
		{5, 5},
		{-5, 5},
		{-1, 1},
		{1234, 1234},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, abs(tt.in))
	}
}

func TestExtractResolution(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		label string
		want  int
	}{
		{"1080p", "1080p", 1080},
		{"720p with prefix", "Quality 720p", 720},
		{"4k label", "2160p UHD", 2160},
		{"no number", "best", 0},
		{"empty", "", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, extractResolution(tt.label))
		})
	}
}

func TestIsNumericString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		s    string
		want bool
	}{
		{"int", "12", true},
		{"decimal", "12.5", true},
		{"text", "abc", false},
		{"mixed", "1a2", false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, isNumericString(tt.s))
		})
	}
}

func TestIsLikelyAllAnimeID(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		s    string
		want bool
	}{
		{"id-like", "hHjXnUTda", true},
		{"http rejected", "https://x/y", false},
		{"numeric rejected", "12345", false},
		{"too short", "abc", false},
		{"too long", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaa1", false},
		{"no letter", "1234567", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, isLikelyAllAnimeID(tt.s))
		})
	}
}

func TestDownloadFolderFormatter(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"video url", "https://animefire.io/video/anime-naruto-ep1", "anime-naruto-ep1"},
		{"no match", "https://example.com/", ""},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, DownloadFolderFormatter(tt.in))
		})
	}
}

func TestExtractEpisodeNumber(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"plain", "Episódio 5", "5"},
		{"two digit", "Episódio 42", "42"},
		{"movie tag", "Filme", "1"},
		{"ova", "OVA", "1"},
		{"special", "Special", "1"},
		{"naked numeric", "12", "12"},
		{"empty fallback", "", "1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, ExtractEpisodeNumber(tt.in))
		})
	}
}

func TestIsPlayableVideoURL(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{"mp4", "https://x.com/v.mp4", true},
		{"mp4 query", "https://x.com/v.mp4?t=1", true},
		{"m3u8", "https://x.com/m.m3u8", true},
		{"webm", "https://x.com/v.webm", true},
		{"source param", "https://x.com/play?source=video", true},
		{"html", "https://x.com/index.html", false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, isPlayableVideoURL(tt.url))
		})
	}
}

func TestIsAllAnimeSourcePlayer(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		anime *models.Anime
		want  bool
	}{
		{"nil", nil, false},
		{"source field", &models.Anime{Source: "AllAnime"}, true},
		{"url contains allanime", &models.Anime{URL: "https://allanime.to/x"}, true},
		{"short id", &models.Anime{URL: "hHjXnUTda"}, true},
		{"animedrive short rejected", &models.Anime{URL: "animesdrive"}, false},
		{"animefire", &models.Anime{Source: "AnimeFire"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, isAllAnimeSourcePlayer(tt.anime))
		})
	}
}

func TestIsAnimeDriveSourcePlayer(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		anime *models.Anime
		want  bool
	}{
		{"nil", nil, false},
		{"source", &models.Anime{Source: "AnimeDrive"}, true},
		{"name tag", &models.Anime{Name: "Naruto [AnimeDrive]"}, true},
		{"url", &models.Anime{URL: "https://animesdrive.blog/x"}, true},
		{"unrelated", &models.Anime{Source: "AllAnime"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, isAnimeDriveSourcePlayer(tt.anime))
		})
	}
}
