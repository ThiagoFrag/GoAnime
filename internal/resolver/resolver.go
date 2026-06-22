package resolver

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
	default:
		return rawURL, "", nil
	}
}
