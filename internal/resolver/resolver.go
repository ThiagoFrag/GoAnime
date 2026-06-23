package resolver

import (
	"net/url"
	"strings"
)

// Resolve turns a scraper's raw stream URL into a directly-playable one.
// It returns (directURL, referer, error); referer is "" when none is required.
// Unknown URLs pass through unchanged so it is safe to call on every source.
func Resolve(rawURL, quality string) (directURL, referer string, err error) {
	switch {
	case IsBloggerURL(rawURL):
		// Blogger googlevideo URLs are TLS-fingerprint locked (403 to mpv), so
		// serve them through a local Chrome-TLS proxy and hand back its URL.
		gv, err := ResolveBlogger(rawURL)
		if err != nil {
			return "", "", err
		}
		local, err := startBloggerProxy(gv)
		if err != nil {
			return "", "", err
		}
		return local, "", nil
	case IsAnimeFireVideoURL(rawURL):
		return ResolveAnimeFireJSON(rawURL, quality)
	case IsAniVideoURL(rawURL):
		// api.anivideo.fun/videohls.php?d=<hls_url> — extract the d= param.
		return resolveAniVideo(rawURL)
	default:
		return rawURL, "", nil
	}
}

// IsAniVideoURL reports whether u is an anivideo.fun HLS wrapper URL.
func IsAniVideoURL(u string) bool {
	return strings.Contains(strings.ToLower(u), "anivideo.fun/videohls.php")
}

// resolveAniVideo extracts the direct HLS URL from the anivideo.fun wrapper.
func resolveAniVideo(rawURL string) (directURL, referer string, err error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", "", err
	}
	d := parsed.Query().Get("d")
	if d == "" {
		return "", "", nil // no d= param, pass through as-is
	}
	return d, "", nil
}
