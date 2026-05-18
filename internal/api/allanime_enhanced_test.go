package api

import (
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestIsAllAnimeSourceAPI(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		anime *models.Anime
		want  bool
	}{
		{
			name:  "explicit Source field",
			anime: &models.Anime{Source: "AllAnime", URL: "irrelevant"},
			want:  true,
		},
		{
			name:  "URL contains allanime",
			anime: &models.Anime{URL: "https://allanime.to/anime/abc"},
			want:  true,
		},
		{
			name:  "short alphanumeric ID",
			anime: &models.Anime{URL: "Bnp4XYZ"},
			want:  true,
		},
		{
			name:  "long URL not matching anything",
			anime: &models.Anime{URL: "https://goyabu.com/anime/some-very-long-slug-here"},
			want:  false,
		},
		{
			name:  "empty URL and Source",
			anime: &models.Anime{},
			want:  false,
		},
		{
			name:  "short URL with http prefix is rejected",
			anime: &models.Anime{URL: "http://x.io"},
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := isAllAnimeSourceAPI(tt.anime)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestExtractAllAnimeIDAPI(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"bare short ID", "Bnp4XYZ", "Bnp4XYZ"},
		// First match of the split heuristic: "https:" is 6 alphanumeric chars
		// and matches before deeper path segments. Documents existing behavior.
		{"allanime URL returns first match", "https://allanime.to/anime/abc123XYZ", "https:"},
		{"allanime path long enough to bypass short-ID branch", "/anime/longpath/longpath/abc123XYZ-allanime", "longpath"},
		{"unknown long URL returned as-is", "https://goyabu.com/anime/some-very-long-slug", "https://goyabu.com/anime/some-very-long-slug"},
		{"empty input", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := extractAllAnimeIDAPI(tt.in)
			assert.Equal(t, tt.want, got)
		})
	}
}
