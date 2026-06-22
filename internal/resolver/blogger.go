// Package resolver turns the intermediate URLs that scrapers return (Blogger
// embeds, AnimeFire video-JSON endpoints) into directly playable media URLs.
//
// It lives below both internal/scraper and internal/player in the import graph
// (it imports neither) so the public pkg/goanime can return playable URLs
// without the GUI having to reimplement extraction.
package resolver

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"

	g "github.com/enetx/g"
	"github.com/enetx/surf"
)

var (
	bloggerTokenRe = regexp.MustCompile(`token=([A-Za-z0-9_-]+)`)
	bloggerSidRe   = regexp.MustCompile(`"FdrFJe"\s*:\s*"([^"]+)"`)
	bloggerBhRe    = regexp.MustCompile(`"cfb2h"\s*:\s*"([^"]+)"`)
	bloggerAtRe    = regexp.MustCompile(`"SNlM0e"\s*:\s*"([^"]+)"`)
	googleVideoRe  = regexp.MustCompile(`https://[^"\\]+\.googlevideo\.com/[^"\\]+`)
)

// ErrBloggerVideoUnavailable means Google's batchexecute returned HTTP 200 with
// an empty payload — the token is dead (video deleted/region-blocked). Retrying
// can't recover it; callers should fall through to the next source.
var ErrBloggerVideoUnavailable = errors.New("blogger video unavailable upstream")

// IsBloggerURL reports whether u is a Blogger video embed needing extraction.
func IsBloggerURL(u string) bool {
	l := strings.ToLower(u)
	return strings.Contains(l, "blogger.com/video.g") || strings.Contains(l, "blogspot.com/video.g")
}

var (
	bloggerSessionClient     *surf.Client
	bloggerSessionClientOnce sync.Once
)

func getBloggerSessionClient() *surf.Client {
	bloggerSessionClientOnce.Do(func() {
		bloggerSessionClient = surf.NewClient().
			Builder().
			Impersonate().Chrome().
			Build().
			Unwrap()
	})
	return bloggerSessionClient
}

// ResolveBlogger extracts the direct googlevideo CDN URL from a Blogger
// "video.g?token=..." embed via Blogger's batchexecute API, using a Chrome-
// impersonating client so Google's CDN accepts the request.
func ResolveBlogger(bloggerURL string) (string, error) {
	tokenMatch := bloggerTokenRe.FindStringSubmatch(bloggerURL)
	if len(tokenMatch) < 2 {
		return "", fmt.Errorf("could not extract token from Blogger URL: %s", bloggerURL)
	}
	token := tokenMatch[1]

	client := getBloggerSessionClient()

	// Step 1: load the Blogger page to extract session params.
	result := client.Get(g.String(bloggerURL)).Do()
	if result.IsErr() {
		return "", fmt.Errorf("failed to load Blogger page: %w", result.Err())
	}
	resp := result.Ok()
	pageBody, err := io.ReadAll(io.LimitReader(resp.Body.Stream(), 10*1024*1024))
	if err != nil {
		return "", fmt.Errorf("failed to read Blogger page: %w", err)
	}
	pageText := string(pageBody)

	sidMatch := bloggerSidRe.FindStringSubmatch(pageText)
	bhMatch := bloggerBhRe.FindStringSubmatch(pageText)
	atMatch := bloggerAtRe.FindStringSubmatch(pageText)
	if len(sidMatch) < 2 || len(bhMatch) < 2 {
		return "", errors.New("failed to extract session params (FdrFJe/cfb2h) from Blogger page")
	}
	sid := sidMatch[1]
	bh := bhMatch[1]
	at := ""
	if len(atMatch) >= 2 {
		at = atMatch[1]
	}

	// Step 2: call batchexecute (WcwnYd RPC) to get the googlevideo URL.
	inner, err := json.Marshal([]any{token, "", 0})
	if err != nil {
		return "", fmt.Errorf("failed to marshal inner data: %w", err)
	}
	freq, err := json.Marshal([][]any{{[]any{"WcwnYd", string(inner), nil, "generic"}}})
	if err != nil {
		return "", fmt.Errorf("failed to marshal freq data: %w", err)
	}
	postData := "f.req=" + url.QueryEscape(string(freq))
	if at != "" {
		postData += "&at=" + url.QueryEscape(at)
	}

	batchURL := fmt.Sprintf(
		"https://www.blogger.com/_/BloggerVideoPlayerUi/data/batchexecute?rpcids=WcwnYd&source-path=%%2Fvideo.g&f.sid=%s&bl=%s&hl=en-US&_reqid=100001&rt=c",
		url.QueryEscape(sid), url.QueryEscape(bh),
	)

	batchResult := client.Post(g.String(batchURL)).
		SetHeaders("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8").
		AddHeaders("X-Same-Domain", "1").
		AddHeaders("Origin", "https://www.blogger.com").
		AddHeaders("Referer", bloggerURL).
		Body(postData).
		Do()
	if batchResult.IsErr() {
		return "", fmt.Errorf("batchexecute request failed: %w", batchResult.Err())
	}
	batchResp := batchResult.Ok()
	batchBody, err := io.ReadAll(io.LimitReader(batchResp.Body.Stream(), 5*1024*1024))
	if err != nil {
		return "", fmt.Errorf("failed to read batchexecute response: %w", err)
	}
	if int(batchResp.StatusCode) != http.StatusOK {
		return "", fmt.Errorf("batchexecute returned status %d", batchResp.StatusCode)
	}

	return parseBatchexecuteResponse(batchBody)
}

// parseBatchexecuteResponse extracts the best MP4 URL from a Google
// batchexecute (WcwnYd) response, preferring 720p (itag=22) over 360p.
func parseBatchexecuteResponse(body []byte) (string, error) {
	var videoURL string
	for _, line := range strings.Split(string(body), "\n") {
		if !strings.Contains(line, "wrb.fr") {
			continue
		}
		var outer []any
		if err := json.Unmarshal([]byte(line), &outer); err != nil {
			continue
		}
		for _, entry := range outer {
			arr, ok := entry.([]any)
			if !ok || len(arr) < 3 {
				continue
			}
			if fmt.Sprint(arr[0]) != "wrb.fr" || fmt.Sprint(arr[1]) != "WcwnYd" {
				continue
			}
			var data []any
			if err := json.Unmarshal(fmt.Append(nil, arr[2]), &data); err != nil {
				continue
			}
			var streams []any
			for _, elem := range data {
				if s, ok := elem.([]any); ok && len(s) > 0 {
					if _, isSlice := s[0].([]any); isSlice {
						streams = s
						break
					}
				}
			}
			if streams == nil {
				continue
			}
			var mp4URLs []string
			for _, s := range streams {
				stream, ok := s.([]any)
				if !ok || len(stream) < 1 {
					continue
				}
				u, ok := stream[0].(string)
				if !ok {
					continue
				}
				if strings.Contains(u, "mime=video%2Fmp4") || strings.Contains(u, "mime=video/mp4") {
					mp4URLs = append(mp4URLs, u)
				}
			}
			for _, u := range mp4URLs {
				if strings.Contains(u, "itag=22") {
					videoURL = u
					break
				}
			}
			if videoURL == "" && len(mp4URLs) > 0 {
				videoURL = mp4URLs[0]
			}
			if videoURL == "" && len(streams) > 0 {
				if first, ok := streams[0].([]any); ok && len(first) > 0 {
					if u, ok := first[0].(string); ok {
						videoURL = u
					}
				}
			}
			break
		}
		if videoURL != "" {
			break
		}
	}

	if videoURL == "" {
		if match := googleVideoRe.FindString(string(body)); match != "" {
			videoURL = match
		}
	}

	if videoURL == "" {
		stripped := bytes.TrimSpace(bytes.TrimPrefix(bytes.TrimSpace(body), []byte(")]}'")))
		if len(stripped) == 0 {
			return "", ErrBloggerVideoUnavailable
		}
		return "", errors.New("no video URL found in batchexecute response")
	}
	return videoURL, nil
}
