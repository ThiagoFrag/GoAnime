package resolver

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/enetx/surf"
)

// Google's CDN serves these blogger googlevideo URLs only to a Chrome TLS
// fingerprint, so a plain mpv/curl request gets HTTP 403. We run a tiny local
// proxy that fetches the bytes with a Chrome-impersonating client and re-streams
// them (forwarding Range requests for seeking) to whatever local player asks.

var bloggerProxy struct {
	mu       sync.Mutex
	server   *http.Server
	port     string
	videoURL string
}

// StopBloggerProxy shuts down any running proxy.
func StopBloggerProxy() {
	bloggerProxy.mu.Lock()
	defer bloggerProxy.mu.Unlock()
	if bloggerProxy.server != nil {
		_ = bloggerProxy.server.Close()
		bloggerProxy.server = nil
		bloggerProxy.port = ""
	}
}

// startBloggerProxy serves the given googlevideo URL through a local HTTP proxy
// using a Chrome-TLS client, and returns the local URL for a player to open.
// Only one proxy runs at a time (a new stream replaces the previous).
func startBloggerProxy(videoURL string) (string, error) {
	StopBloggerProxy()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", fmt.Errorf("failed to listen on a free port: %w", err)
	}
	port := strconv.Itoa(listener.Addr().(*net.TCPAddr).Port)

	// Chrome-TLS client with no request deadline (streaming would otherwise be
	// killed by surf's default 30s timeout).
	proxyClient := surf.NewClient().
		Builder().
		Impersonate().Chrome().
		NotFollowRedirects().
		Build().
		Unwrap().
		Std()
	proxyClient.Timeout = 0

	mux := http.NewServeMux()
	mux.HandleFunc("/blogger_proxy", func(w http.ResponseWriter, r *http.Request) {
		method := r.Method
		if method != http.MethodHead && method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		upReq, err := http.NewRequest(method, videoURL, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		if rng := r.Header.Get("Range"); rng != "" {
			upReq.Header.Set("Range", rng)
		}
		upResp, err := proxyClient.Do(upReq)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer func() { _ = upResp.Body.Close() }()
		for _, k := range []string{"Content-Type", "Content-Length", "Content-Range", "Accept-Ranges"} {
			if v := upResp.Header.Get(k); v != "" {
				w.Header().Set(k, v)
			}
		}
		w.WriteHeader(upResp.StatusCode)
		if method == http.MethodGet {
			_, _ = io.Copy(w, upResp.Body)
		}
	})

	srv := &http.Server{Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	bloggerProxy.mu.Lock()
	bloggerProxy.server = srv
	bloggerProxy.port = port
	bloggerProxy.videoURL = videoURL
	bloggerProxy.mu.Unlock()

	go func() {
		if err := srv.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			_ = err
		}
	}()

	proxyURL := fmt.Sprintf("http://127.0.0.1:%s/blogger_proxy", port)

	// Readiness poll: return as soon as the proxy answers (or after 3s anyway).
	probe := &http.Client{Timeout: 2 * time.Second}
	deadline := time.After(3 * time.Second)
	tick := time.NewTicker(40 * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case <-deadline:
			return proxyURL, nil
		case <-tick.C:
			if resp, e := probe.Head(proxyURL); e == nil {
				_ = resp.Body.Close()
				return proxyURL, nil
			}
		}
	}
}
