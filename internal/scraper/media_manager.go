// Package scraper provides unified media handling for anime, movies, and TV shows
package scraper

import (
	"context"
	"fmt"
	"strings"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
)

// MediaManager provides a unified interface for all media types.
// Movies/TV are served via the SFlix client only.
type MediaManager struct {
	scraperManager *ScraperManager
	sflixClient    *SFlixClient
}

// NewMediaManager creates a new MediaManager
func NewMediaManager() *MediaManager {
	sm := NewScraperManager()

	var sflixClient *SFlixClient
	if adapter, ok := sm.scrapers[SFlixType].(*SFlixAdapter); ok {
		sflixClient = adapter.client
	} else {
		sflixClient = NewSFlixClient()
	}

	return &MediaManager{
		scraperManager: sm,
		sflixClient:    sflixClient,
	}
}

// SearchAll searches across all sources (anime + movies/TV)
func (mm *MediaManager) SearchAll(query string) ([]*models.Anime, error) {
	return mm.scraperManager.SearchAnime(query, nil)
}

// SearchAnimeOnly searches only anime sources concurrently
func (mm *MediaManager) SearchAnimeOnly(query string) ([]*models.Anime, error) {
	type sourceResult struct {
		results []*models.Anime
		err     error
	}

	ch := make(chan sourceResult, 2)

	go func() {
		t := AllAnimeType
		results, err := mm.scraperManager.SearchAnime(query, &t)
		ch <- sourceResult{results: results, err: err}
	}()

	go func() {
		t := AnimefireType
		results, err := mm.scraperManager.SearchAnime(query, &t)
		ch <- sourceResult{results: results, err: err}
	}()

	var allResults []*models.Anime
	for range 2 {
		res := <-ch
		if res.err == nil {
			allResults = append(allResults, res.results...)
		}
	}

	if len(allResults) == 0 {
		return nil, fmt.Errorf("no anime found with name: %s", query)
	}

	return allResults, nil
}


// SearchMoviesAndTV searches SFlix for movies and TV shows.
func (mm *MediaManager) SearchMoviesAndTV(query string) ([]*SFlixMedia, error) {
	return mm.sflixClient.SearchMedia(query)
}

// SearchSFlixMoviesAndTV searches SFlix for movies and TV shows.
func (mm *MediaManager) SearchSFlixMoviesAndTV(query string) ([]*SFlixMedia, error) {
	return mm.sflixClient.SearchMedia(query)
}

// GetTrendingMovies gets trending movies from SFlix.
func (mm *MediaManager) GetTrendingMovies() ([]*SFlixMedia, error) {
	return mm.sflixClient.GetTrending()
}

// GetSFlixTrendingMovies gets trending movies from SFlix.
func (mm *MediaManager) GetSFlixTrendingMovies() ([]*SFlixMedia, error) {
	return mm.sflixClient.GetTrending()
}

// GetRecentMovies gets recent movies from SFlix.
func (mm *MediaManager) GetRecentMovies() ([]*SFlixMedia, error) {
	return mm.sflixClient.GetRecentMovies()
}

// GetSFlixRecentMovies gets recent movies from SFlix.
func (mm *MediaManager) GetSFlixRecentMovies() ([]*SFlixMedia, error) {
	return mm.sflixClient.GetRecentMovies()
}

// GetRecentTV gets recent TV shows from SFlix.
func (mm *MediaManager) GetRecentTV() ([]*SFlixMedia, error) {
	return mm.sflixClient.GetRecentTV()
}

// GetSFlixRecentTV gets recent TV shows from SFlix.
func (mm *MediaManager) GetSFlixRecentTV() ([]*SFlixMedia, error) {
	return mm.sflixClient.GetRecentTV()
}

// GetSFlixTVSeasons gets all seasons for a TV show from SFlix.
func (mm *MediaManager) GetSFlixTVSeasons(mediaID string) ([]SFlixSeason, error) {
	return mm.sflixClient.GetSeasons(mediaID)
}

// GetSFlixTVEpisodes gets all episodes for a season from SFlix.
func (mm *MediaManager) GetSFlixTVEpisodes(seasonID string) ([]SFlixEpisode, error) {
	return mm.sflixClient.GetEpisodes(seasonID)
}

// GetSFlixMovieStreamInfo gets stream information for a movie from SFlix.
func (mm *MediaManager) GetSFlixMovieStreamInfo(mediaID, provider, quality, subsLanguage string) (*SFlixStreamInfo, error) {
	if provider == "" {
		provider = "Vidcloud"
	}
	if quality == "" {
		quality = "1080"
	}
	if subsLanguage == "" {
		subsLanguage = "english"
	}

	episodeID, err := mm.sflixClient.GetMovieServerID(mediaID, provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get movie server: %w", err)
	}

	embedLink, err := mm.sflixClient.GetEmbedLink(episodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get embed link: %w", err)
	}

	return mm.sflixClient.ExtractStreamInfo(embedLink, quality, subsLanguage)
}

// GetSFlixTVEpisodeStreamInfo gets stream information for a TV episode from SFlix.
func (mm *MediaManager) GetSFlixTVEpisodeStreamInfo(dataID, provider, quality, subsLanguage string) (*SFlixStreamInfo, error) {
	if provider == "" {
		provider = "Vidcloud"
	}
	if quality == "" {
		quality = "1080"
	}
	if subsLanguage == "" {
		subsLanguage = "english"
	}

	episodeID, err := mm.sflixClient.GetEpisodeServerID(dataID, provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get episode server: %w", err)
	}

	embedLink, err := mm.sflixClient.GetEmbedLink(episodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get embed link: %w", err)
	}

	return mm.sflixClient.ExtractStreamInfo(embedLink, quality, subsLanguage)
}

// GetAnimeStreamURL gets stream URL for anime episodes
func (mm *MediaManager) GetAnimeStreamURL(anime *models.Anime, episodeNum string, quality, mode string) (string, map[string]string, error) {
	source := strings.ToLower(anime.Source)

	util.Debug("Getting stream URL", "source", source, "anime", anime.Name, "episode", episodeNum)

	switch {
	case strings.Contains(source, "allanime"):
		scraper, err := mm.scraperManager.GetScraper(AllAnimeType)
		if err != nil {
			return "", nil, err
		}
		return scraper.GetStreamURL(anime.URL, episodeNum, quality, mode)

	case strings.Contains(source, "animefire"):
		scraper, err := mm.scraperManager.GetScraper(AnimefireType)
		if err != nil {
			return "", nil, err
		}
		return scraper.GetStreamURL(anime.URL, episodeNum, quality, mode)

	case strings.Contains(source, "animedrive"):
		scraper, err := mm.scraperManager.GetScraper(AnimeDriveType)
		if err != nil {
			return "", nil, err
		}
		return scraper.GetStreamURL(anime.URL, episodeNum, quality, mode)

	default:
		return "", nil, fmt.Errorf("unknown source: %s", anime.Source)
	}
}

// ConvertSFlixToAnime converts SFlix media list to Anime models for unified handling
func ConvertSFlixToAnime(media []*SFlixMedia) []*models.Anime {
	var animes []*models.Anime
	for _, m := range media {
		anime := m.ToAnimeModel()
		if m.Type == MediaTypeMovie {
			anime.MediaType = models.MediaTypeMovie
		} else {
			anime.MediaType = models.MediaTypeTV
		}
		anime.Year = m.Year
		animes = append(animes, anime)
	}
	return animes
}

// ConvertSFlixEpisodesToEpisodes converts SFlix episodes to Episode models
func ConvertSFlixEpisodesToEpisodes(episodes []SFlixEpisode) []models.Episode {
	var eps []models.Episode
	for _, e := range episodes {
		eps = append(eps, e.ToEpisodeModel())
	}
	return eps
}

// GetScraperManager returns the underlying scraper manager for advanced usage
func (mm *MediaManager) GetScraperManager() *ScraperManager {
	return mm.scraperManager
}

// GetSFlixClient returns the SFlix client for direct access
func (mm *MediaManager) GetSFlixClient() *SFlixClient {
	return mm.sflixClient
}

// GetSFlixMovieInfo gets detailed info for a movie or TV show from SFlix.
func (mm *MediaManager) GetSFlixMovieInfo(id string) (*SFlixMedia, error) {
	return mm.sflixClient.GetInfo(id)
}

// GetSFlixMovieInfoWithContext gets detailed info with context support.
func (mm *MediaManager) GetSFlixMovieInfoWithContext(ctx context.Context, id string) (*SFlixMedia, error) {
	return mm.sflixClient.GetInfoWithContext(ctx, id)
}

// GetSFlixServers gets available streaming servers from SFlix.
func (mm *MediaManager) GetSFlixServers(episodeID string, isMovie bool) ([]SFlixServer, error) {
	return mm.sflixClient.GetServers(episodeID, isMovie)
}
