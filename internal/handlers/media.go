// Package handlers provides HTTP handlers and flow controllers for media playback
package handlers

import (
	"context"
	"fmt"
	"strings"

	"charm.land/huh/v2"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/alvarorichard/Goanime/internal/tui"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/ktr0731/go-fuzzyfinder"
)

// MediaHandler handles media selection and playback operations
type MediaHandler struct {
	mediaManager *scraper.MediaManager
	provider     string
	quality      scraper.Quality
	subsLanguage string
}

// NewMediaHandler creates a new MediaHandler
func NewMediaHandler() *MediaHandler {
	return &MediaHandler{
		mediaManager: scraper.NewMediaManager(),
		provider:     "Vidcloud",
		quality:      scraper.Quality1080,
		subsLanguage: "english",
	}
}

// SetOptions sets playback options
func (mh *MediaHandler) SetOptions(provider, quality, subsLanguage string) {
	if provider != "" {
		mh.provider = provider
	}
	if quality != "" {
		mh.quality = scraper.Quality(quality)
	}
	if subsLanguage != "" {
		mh.subsLanguage = subsLanguage
	}
}

// SearchMedia searches for media based on content type
func (mh *MediaHandler) SearchMedia(query string, contentType models.MediaType) ([]*models.Anime, error) {
	switch contentType {
	case models.MediaTypeAnime:
		return mh.mediaManager.SearchAnimeOnly(query)
	case models.MediaTypeMovie, models.MediaTypeTV:
		media, err := mh.mediaManager.SearchMoviesAndTV(query)
		if err != nil {
			return nil, err
		}
		return scraper.ConvertSFlixToAnime(media), nil
	default:
		return mh.mediaManager.SearchAll(query)
	}
}

// SelectMediaType prompts user to select media type
func (mh *MediaHandler) SelectMediaType() (models.MediaType, error) {
	items := []string{"Anime", "Movies", "TV Shows", "Search All"}
	idx, err := tui.Find(items, func(i int) string {
		return items[i]
	}, fuzzyfinder.WithPromptString("Select content type: "))
	if err != nil {
		return "", err
	}

	switch idx {
	case 0:
		return models.MediaTypeAnime, nil
	case 1:
		return models.MediaTypeMovie, nil
	case 2:
		return models.MediaTypeTV, nil
	default:
		return "", nil // Search all
	}
}

// SelectMedia prompts user to select from search results
func (mh *MediaHandler) SelectMedia(results []*models.Anime) (*models.Anime, error) {
	if len(results) == 0 {
		return nil, fmt.Errorf("no results to select from")
	}

	idx, err := tui.Find(results, func(i int) string {
		r := results[i]
		typeTag := ""
		switch r.MediaType {
		case models.MediaTypeMovie:
			typeTag = "[Movie]"
		case models.MediaTypeTV:
			typeTag = "[TV]"
		case models.MediaTypeAnime:
			typeTag = "[Anime]"
		}
		year := ""
		if r.Year != "" {
			year = fmt.Sprintf(" (%s)", r.Year)
		}
		return fmt.Sprintf("%s %s%s - %s", typeTag, r.Name, year, r.Source)
	}, fuzzyfinder.WithPromptString("Select media: "))
	if err != nil {
		return nil, err
	}

	return results[idx], nil
}

// SelectSeason prompts user to select a TV season
func (mh *MediaHandler) SelectSeason(mediaID string) (*scraper.SFlixSeason, error) {
	seasons, err := mh.mediaManager.GetSFlixTVSeasons(mediaID)
	if err != nil {
		return nil, err
	}

	if len(seasons) == 0 {
		return nil, fmt.Errorf("no seasons found")
	}

	idx, err := tui.Find(seasons, func(i int) string {
		return seasons[i].Title
	}, fuzzyfinder.WithPromptString("Select season: "))
	if err != nil {
		return nil, err
	}

	return &seasons[idx], nil
}

// SelectEpisode prompts user to select a TV episode
func (mh *MediaHandler) SelectEpisode(seasonID string) (*scraper.SFlixEpisode, error) {
	episodes, err := mh.mediaManager.GetSFlixTVEpisodes(seasonID)
	if err != nil {
		return nil, err
	}

	if len(episodes) == 0 {
		return nil, fmt.Errorf("no episodes found")
	}

	idx, err := tui.Find(episodes, func(i int) string {
		return fmt.Sprintf("Episode %d: %s", episodes[i].Number, episodes[i].Title)
	}, fuzzyfinder.WithPromptString("Select episode: "))
	if err != nil {
		return nil, err
	}

	return &episodes[idx], nil
}

// GetStreamInfo gets streaming information for selected media
func (mh *MediaHandler) GetStreamInfo(media *models.Anime, episode *scraper.SFlixEpisode) (*scraper.SFlixStreamInfo, error) {
	return mh.GetStreamInfoWithContext(context.Background(), media, episode)
}

// GetStreamInfoWithContext gets streaming information with context support
func (mh *MediaHandler) GetStreamInfoWithContext(ctx context.Context, media *models.Anime, episode *scraper.SFlixEpisode) (*scraper.SFlixStreamInfo, error) {
	source := strings.ToLower(media.Source)

	if !strings.Contains(source, "sflix") {
		return nil, fmt.Errorf("media source %s does not support SFlix streaming", media.Source)
	}

	mediaID := extractIDFromURL(media.URL)
	if mediaID == "" {
		return nil, fmt.Errorf("could not extract media ID from URL: %s", media.URL)
	}

	if media.MediaType == models.MediaTypeMovie {
		return mh.mediaManager.GetSFlixMovieStreamInfo(mediaID, mh.provider, string(mh.quality), mh.subsLanguage)
	}

	if episode == nil {
		return nil, fmt.Errorf("episode is required for TV shows")
	}

	return mh.mediaManager.GetSFlixTVEpisodeStreamInfo(episode.DataID, mh.provider, string(mh.quality), mh.subsLanguage)
}

// GetAnimeStreamURL gets stream URL for anime content
func (mh *MediaHandler) GetAnimeStreamURL(anime *models.Anime, episodeNum string, mode string) (string, map[string]string, error) {
	return mh.mediaManager.GetAnimeStreamURL(anime, episodeNum, string(mh.quality), mode)
}

// InteractiveMediaFlow runs an interactive media selection and playback flow
func (mh *MediaHandler) InteractiveMediaFlow(query string) (*PlaybackInfo, error) {
	// Select media type if not already searching
	var contentType models.MediaType
	if query == "" {
		var err error
		contentType, err = mh.SelectMediaType()
		if err != nil {
			return nil, err
		}
	}

	if query == "" {
		var searchQuery string
		prompt := huh.NewInput().
			Title("Search").
			Value(&searchQuery)
		if err := tui.RunClean(prompt.Run); err != nil {
			return nil, err
		}
		query = searchQuery
	}

	results, err := mh.SearchMedia(query, contentType)
	if err != nil {
		return nil, err
	}

	util.Debug("Search results", "count", len(results))

	selected, err := mh.SelectMedia(results)
	if err != nil {
		return nil, err
	}

	playbackInfo := &PlaybackInfo{
		Title:     selected.Name,
		MediaType: selected.MediaType,
		Source:    selected.Source,
		ImageURL:  selected.ImageURL,
	}

	if strings.Contains(strings.ToLower(selected.Source), "sflix") {
		return mh.handleSFlixPlayback(selected, playbackInfo)
	}

	return mh.handleAnimePlayback(selected, playbackInfo)
}

func (mh *MediaHandler) handleSFlixPlayback(media *models.Anime, info *PlaybackInfo) (*PlaybackInfo, error) {
	mediaID := extractIDFromURL(media.URL)

	if media.MediaType == models.MediaTypeMovie {
		streamInfo, err := mh.mediaManager.GetSFlixMovieStreamInfo(mediaID, mh.provider, string(mh.quality), mh.subsLanguage)
		if err != nil {
			return nil, err
		}
		info.StreamURL = streamInfo.VideoURL
		info.Subtitles = convertSubtitles(streamInfo.Subtitles)
		info.Quality = string(mh.quality)
		return info, nil
	}

	// TV Show flow
	season, err := mh.SelectSeason(mediaID)
	if err != nil {
		return nil, err
	}
	info.Season = season.Title

	episode, err := mh.SelectEpisode(season.ID)
	if err != nil {
		return nil, err
	}
	info.Episode = episode.Title
	info.EpisodeNum = episode.Number

	streamInfo, err := mh.mediaManager.GetSFlixTVEpisodeStreamInfo(episode.DataID, mh.provider, string(mh.quality), mh.subsLanguage)
	if err != nil {
		return nil, err
	}
	info.StreamURL = streamInfo.VideoURL
	info.Subtitles = convertSubtitles(streamInfo.Subtitles)
	info.Quality = string(mh.quality)

	return info, nil
}

func (mh *MediaHandler) handleAnimePlayback(anime *models.Anime, info *PlaybackInfo) (*PlaybackInfo, error) {
	var episodeNum string
	prompt := huh.NewInput().
		Title("Episode number").
		Value(&episodeNum).
		Validate(func(v string) error {
			if len(v) == 0 {
				return fmt.Errorf("episode number is required")
			}
			return nil
		})

	if err := tui.RunClean(prompt.Run); err != nil {
		return nil, err
	}
	if episodeNum == "" {
		episodeNum = "1"
	}

	modeItems := []string{"Sub (Subtitled)", "Dub (English Dubbed)"}
	modeIdx, err := tui.Find(modeItems, func(i int) string {
		return modeItems[i]
	}, fuzzyfinder.WithPromptString("Select audio: "))
	if err != nil {
		return nil, err
	}

	mode := "sub"
	if modeIdx == 1 {
		mode = "dub"
	}

	streamURL, metadata, err := mh.GetAnimeStreamURL(anime, episodeNum, mode)
	if err != nil {
		return nil, err
	}

	info.StreamURL = streamURL
	info.Episode = fmt.Sprintf("Episode %s", episodeNum)
	info.Metadata = metadata

	return info, nil
}

// PlaybackInfo contains all information needed for playback
type PlaybackInfo struct {
	Title      string
	MediaType  models.MediaType
	Source     string
	Season     string
	Episode    string
	EpisodeNum int
	StreamURL  string
	Quality    string
	Subtitles  []models.Subtitle
	Referer    string
	ImageURL   string
	Metadata   map[string]string
}

// Helper functions

func extractIDFromURL(urlStr string) string {
	// Extract ID from URL like /movie/watch-movie-name-12345 or /tv/watch-show-name-12345
	parts := strings.Split(urlStr, "-")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

func convertSubtitles(subsIn []scraper.SFlixSubtitle) []models.Subtitle {
	var subs []models.Subtitle
	for _, s := range subsIn {
		subs = append(subs, models.Subtitle{
			URL:      s.URL,
			Language: s.Language,
			Label:    s.Label,
		})
	}
	return subs
}
