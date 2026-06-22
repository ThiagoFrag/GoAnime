package resolver

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// animeFireReferer is required by the lightspeedst CDN — it returns HTTP 401
// without it.
const animeFireReferer = "https://animefire.io/"

var httpClient = &http.Client{Timeout: 25 * time.Second}

type afSource struct {
	Src   string `json:"src"`
	Label string `json:"label"`
}

// IsAnimeFireVideoURL reports whether u is an AnimeFire JSON endpoint that must
// be resolved to a direct media URL.
func IsAnimeFireVideoURL(u string) bool {
	return strings.Contains(strings.ToLower(u), "animefire.io/video/")
}

// ResolveAnimeFireJSON fetches the AnimeFire "/video/<slug>" JSON endpoint and
// returns a direct .mp4 URL for the requested quality plus the Referer the CDN
// requires.
func ResolveAnimeFireJSON(videoAPIURL, quality string) (directURL, referer string, err error) {
	req, err := http.NewRequest(http.MethodGet, videoAPIURL, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Referer", animeFireReferer)

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", "", err
	}

	var parsed struct {
		Data []afSource `json:"data"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", "", errors.New("animefire: unexpected response")
	}
	if len(parsed.Data) == 0 {
		return "", "", errors.New("animefire: no sources available")
	}
	return pickQuality(parsed.Data, quality), animeFireReferer, nil
}

// pickQuality selects a source by label ("360p","720p"...). "worst" takes the
// lowest, anything else (incl. "best") takes the highest.
func pickQuality(data []afSource, quality string) string {
	best, worst := data[0].Src, data[0].Src
	bestN, worstN := labelN(data[0].Label), labelN(data[0].Label)
	for _, d := range data[1:] {
		n := labelN(d.Label)
		if n > bestN {
			bestN, best = n, d.Src
		}
		if n < worstN {
			worstN, worst = n, d.Src
		}
	}
	if strings.EqualFold(quality, "worst") {
		return worst
	}
	return best
}

func labelN(label string) int {
	var digits strings.Builder
	for _, r := range label {
		if r >= '0' && r <= '9' {
			digits.WriteRune(r)
		}
	}
	n, _ := strconv.Atoi(digits.String())
	return n
}
