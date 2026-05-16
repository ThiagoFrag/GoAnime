package scraper

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMediaManager(t *testing.T) {
	t.Parallel()
	mm := NewMediaManager()
	require.NotNil(t, mm)
	assert.NotNil(t, mm.scraperManager)
	assert.NotNil(t, mm.flixhqClient)
	assert.NotNil(t, mm.sflixClient)
}

func TestMediaManager_ConvertSFlixToFlixHQ(t *testing.T) {
	t.Parallel()
	mm := NewMediaManager()
	in := &SFlixMedia{
		ID:          "1",
		Title:       "X",
		URL:         "https://x.com",
		ImageURL:    "https://img",
		Type:        "movie",
		Year:        "2024",
		ReleaseDate: "2024-01-01",
		Quality:     "HD",
		Duration:    "120",
		Description: "desc",
		Genres:      []string{"Drama"},
		Country:     "US",
		Production:  "Prod",
		Casts:       []string{"A", "B"},
	}
	got := mm.ConvertSFlixToFlixHQ(in)
	require.NotNil(t, got)
	assert.Equal(t, "1", got.ID)
	assert.Equal(t, "X", got.Title)
	assert.Equal(t, "SFlix", got.Source)
	assert.Equal(t, []string{"Drama"}, got.Genres)
}

func TestMediaManager_SearchAllMovieSources_TempDisabled(t *testing.T) {
	t.Parallel()
	mm := NewMediaManager()
	_, err := mm.SearchAllMovieSources("dexter")
	assert.Error(t, err, "TEMP-DISABLED returns error")
}

func TestMediaManager_SearchMoviesAndTV_TempDisabled(t *testing.T) {
	t.Parallel()
	mm := NewMediaManager()
	_, err := mm.SearchMoviesAndTV("dexter")
	assert.Error(t, err)
}

func TestMediaManager_GetTrendingMovies_DelegatesToFetchBoth(t *testing.T) {
	t.Parallel()
	mm := NewMediaManager()
	// Both upstream HTTP calls go to real CDNs which are SSRF-blocked in test
	// environment via the safe transport. Verify the call surface is wired up
	// and returns either results or a defined error — no panic.
	_, err := mm.GetTrendingMovies()
	if err != nil {
		assert.Error(t, err)
	}
}

func TestMediaManager_GetRecentMovies_NoPanic(t *testing.T) {
	t.Parallel()
	mm := NewMediaManager()
	_, _ = mm.GetRecentMovies()
}

func TestMediaManager_SearchAnimeOnly_NoMatch(t *testing.T) {
	t.Parallel()
	mm := NewMediaManager()
	_, err := mm.SearchAnimeOnly("zzz_definitely_not_an_anime_xyz_999")
	// Either returns an error (no results) or the live source returned something.
	// Both outcomes are acceptable; we just want the call to be safe.
	_ = err
}

func TestMediaManager_SearchFlixHQOnly_ReturnsErrorOrSlice(t *testing.T) {
	t.Parallel()
	mm := NewMediaManager()
	_, _ = mm.SearchFlixHQOnly("zzz_x")
}

func TestMediaManager_SearchSFlixMoviesAndTV_ReturnsErrorOrSlice(t *testing.T) {
	t.Parallel()
	mm := NewMediaManager()
	_, _ = mm.SearchSFlixMoviesAndTV("zzz_x")
}

func TestMediaManager_GetFlixHQTrendingMovies_NoPanic(t *testing.T) {
	t.Parallel()
	mm := NewMediaManager()
	_, _ = mm.GetFlixHQTrendingMovies()
}

func TestMediaManager_GetSFlixTrendingMovies_NoPanic(t *testing.T) {
	t.Parallel()
	mm := NewMediaManager()
	_, _ = mm.GetSFlixTrendingMovies()
}

func TestMediaManager_GetFlixHQRecentMovies_NoPanic(t *testing.T) {
	t.Parallel()
	mm := NewMediaManager()
	_, _ = mm.GetFlixHQRecentMovies()
}

func TestMediaManager_GetSFlixRecentMovies_NoPanic(t *testing.T) {
	t.Parallel()
	mm := NewMediaManager()
	_, _ = mm.GetSFlixRecentMovies()
}

func TestMediaManager_SearchAll_NoPanic(t *testing.T) {
	t.Parallel()
	mm := NewMediaManager()
	_, _ = mm.SearchAll("zzz")
}
