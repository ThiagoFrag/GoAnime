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
	assert.NotNil(t, mm.sflixClient)
}

func TestMediaManager_SearchMoviesAndTV_ReturnsErrorOrSlice(t *testing.T) {
	t.Parallel()
	mm := NewMediaManager()
	_, _ = mm.SearchMoviesAndTV("zzz_x")
}

func TestMediaManager_GetTrendingMovies_NoPanic(t *testing.T) {
	t.Parallel()
	mm := NewMediaManager()
	_, _ = mm.GetTrendingMovies()
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
	_ = err
}

func TestMediaManager_SearchSFlixMoviesAndTV_ReturnsErrorOrSlice(t *testing.T) {
	t.Parallel()
	mm := NewMediaManager()
	_, _ = mm.SearchSFlixMoviesAndTV("zzz_x")
}

func TestMediaManager_GetSFlixTrendingMovies_NoPanic(t *testing.T) {
	t.Parallel()
	mm := NewMediaManager()
	_, _ = mm.GetSFlixTrendingMovies()
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
