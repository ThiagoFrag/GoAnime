package scraper

import "strings"

// Shared media-scraper enum types. These are used by SFlix (and previously
// FlixHQ) and any future movie/TV scrapers.

// MediaType represents the type of media (movie or TV show)
type MediaType string

const (
	MediaTypeMovie MediaType = "movie"
	MediaTypeTV    MediaType = "tv"
)

// Quality represents video quality levels
type Quality string

const (
	QualityAuto Quality = "auto"
	Quality360  Quality = "360"
	Quality480  Quality = "480"
	Quality720  Quality = "720"
	Quality1080 Quality = "1080"
	QualityBest Quality = "best"
)

// StreamType represents the type of stream
type StreamType string

const (
	StreamTypeHLS StreamType = "hls"
	StreamTypeMP4 StreamType = "mp4"
)

// ServerName represents known streaming servers
type ServerName string

const (
	ServerVidcloud  ServerName = "Vidcloud"
	ServerUpCloud   ServerName = "UpCloud"
	ServerVoe       ServerName = "Voe"
	ServerMixDrop   ServerName = "MixDrop"
	ServerFilelions ServerName = "Filelions"
)

// DefaultServerPriority defines the preferred server order
var DefaultServerPriority = []ServerName{
	ServerVidcloud,
	ServerUpCloud,
	ServerVoe,
	ServerMixDrop,
	ServerFilelions,
}

// ExtractMediaPath extracts the media path from a full SFlix URL.
// e.g. "https://sflix.to/tv/watch-dexter-39448" → "tv/watch-dexter-39448"
func ExtractMediaPath(fullURL string) string {
	for _, base := range []string{"https://sflix.to/", "http://sflix.to/"} {
		if after, ok := strings.CutPrefix(fullURL, base); ok {
			return after
		}
	}
	if strings.HasPrefix(fullURL, "movie/") || strings.HasPrefix(fullURL, "tv/") {
		return fullURL
	}
	if _, after, ok := strings.Cut(fullURL, "://"); ok {
		rest := after
		if _, after, ok := strings.Cut(rest, "/"); ok {
			return after
		}
	}
	return fullURL
}

// parseQuality normalizes a free-form quality string into a Quality enum.
func parseQuality(q string) Quality {
	q = strings.ToLower(strings.TrimSpace(q))
	switch {
	case q == "auto" || q == "":
		return QualityAuto
	case strings.Contains(q, "360"):
		return Quality360
	case strings.Contains(q, "480"):
		return Quality480
	case strings.Contains(q, "720"):
		return Quality720
	case strings.Contains(q, "1080"):
		return Quality1080
	case q == "best":
		return QualityBest
	default:
		return Quality(q)
	}
}

// qualityToInt converts Quality to int for comparison
func qualityToInt(q Quality) int {
	switch q {
	case Quality360:
		return 360
	case Quality480:
		return 480
	case Quality720:
		return 720
	case Quality1080:
		return 1080
	case QualityBest:
		return 9999
	default:
		return 0
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
