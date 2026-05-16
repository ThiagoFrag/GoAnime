package playback

import (
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsAllAnimeSource(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		anime *models.Anime
		want  bool
	}{
		{"source field", &models.Anime{Source: "AllAnime"}, true},
		{"url contains", &models.Anime{URL: "https://allanime.to/x"}, true},
		{"short id", &models.Anime{URL: "hHjXnUTda"}, true},
		{"http url unrelated", &models.Anime{URL: "https://animefire.io/x"}, false},
		{"empty", &models.Anime{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, isAllAnimeSource(tt.anime))
		})
	}
}

func TestExtractAllAnimeID(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		url  string
		want string
	}{
		{"short id", "hHjXnUTda", "hHjXnUTda"},
		{"allanime url returns first long alphanumeric segment", "https://allanime.to/anime/abc123XYZ/title", "https:"},
		{"non-allanime", "https://example.com/x", "https://example.com/x"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, extractAllAnimeID(tt.url))
		})
	}
}

func TestAllAnimeNavigator_GetNextEpisode(t *testing.T) {
	t.Parallel()
	nav := &AllAnimeNavigator{animeID: "x", episodes: []string{"1", "2", "3"}}

	next, err := nav.GetNextEpisode("1")
	require.NoError(t, err)
	assert.Equal(t, "2", next)

	_, err = nav.GetNextEpisode("3")
	assert.Error(t, err)

	_, err = nav.GetNextEpisode("not-a-number")
	assert.Error(t, err)
}

func TestAllAnimeNavigator_GetPreviousEpisode(t *testing.T) {
	t.Parallel()
	nav := &AllAnimeNavigator{animeID: "x", episodes: []string{"1", "2", "3"}}

	prev, err := nav.GetPreviousEpisode("2")
	require.NoError(t, err)
	assert.Equal(t, "1", prev)

	_, err = nav.GetPreviousEpisode("1")
	assert.Error(t, err)

	_, err = nav.GetPreviousEpisode("bad")
	assert.Error(t, err)
}

func TestAllAnimeNavigator_GetTotalEpisodes(t *testing.T) {
	t.Parallel()
	nav := &AllAnimeNavigator{episodes: []string{"a", "b", "c", "d"}}
	assert.Equal(t, 4, nav.GetTotalEpisodes())
}

func TestAllAnimeNavigator_ListAllEpisodes(t *testing.T) {
	t.Parallel()
	nav := &AllAnimeNavigator{episodes: []string{"a", "b", "c"}}
	list := nav.ListAllEpisodes()
	assert.Equal(t, []string{"1", "2", "3"}, list)
}

func TestNewAllAnimeNavigator_RejectsNonAllAnime(t *testing.T) {
	t.Parallel()
	_, err := NewAllAnimeNavigator(&models.Anime{Source: "AnimeFire", URL: "https://animefire.io/x"})
	assert.Error(t, err)
}

func TestHandleAllAnimeEpisodeNavigation_InvalidDirection(t *testing.T) {
	t.Parallel()
	anime := &models.Anime{Source: "AllAnime", URL: "shortid12"}
	// Seed the navigator cache so we don't hit the network.
	nav := &AllAnimeNavigator{animeID: "shortid12", episodes: []string{"1", "2"}}
	navigatorCacheMu.Lock()
	navigatorCache["shortid12"] = nav
	navigatorCacheMu.Unlock()
	t.Cleanup(func() {
		navigatorCacheMu.Lock()
		delete(navigatorCache, "shortid12")
		navigatorCacheMu.Unlock()
	})

	_, err := HandleAllAnimeEpisodeNavigation(anime, "1", "sideways")
	assert.Error(t, err)
}
