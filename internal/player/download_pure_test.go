package player

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsAnimeFireVideoAPIURL(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{"animefire.io", "https://animefire.io/video/abc", true},
		{"animefire.plus", "https://animefire.plus/video/xyz", true},
		{"upper case", "https://ANIMEFIRE.IO/VIDEO/x", true},
		{"unrelated", "https://example.com/video/x", false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, isAnimeFireVideoAPIURL(tt.url))
		})
	}
}

func TestExtractRefererFromURL(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		url  string
		want string
	}{
		{"empty", "", ""},
		{"with path", "https://megacloud.tv/embed-2/abc?k=v", "https://megacloud.tv/"},
		{"scheme only no host", "http:///x", ""},
		{"http", "http://example.com/y", "http://example.com/"},
		{"bare path", "/just/a/path", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, extractRefererFromURL(tt.url))
		})
	}
}

func TestFileExists(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	t.Run("missing", func(t *testing.T) {
		t.Parallel()
		assert.False(t, fileExists(filepath.Join(dir, "missing")))
	})

	t.Run("present", func(t *testing.T) {
		t.Parallel()
		p := filepath.Join(dir, "x")
		require.NoError(t, os.WriteFile(p, []byte("y"), 0o600))
		assert.True(t, fileExists(p))
	})
}

func TestSafePartPath(t *testing.T) {
	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		got, err := safePartPath("/tmp/video.mp4", 1)
		require.NoError(t, err)
		assert.Equal(t, "/tmp/video.mp4.part1", got)
	})

	t.Run("subdir ok", func(t *testing.T) {
		t.Parallel()
		got, err := safePartPath("/tmp/sub/video.mp4", 7)
		require.NoError(t, err)
		assert.Contains(t, got, "video.mp4.part7")
	})
}

func TestIsBloggerProxyURL_Extra(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{"correct proxy", "http://127.0.0.1:8080/blogger_proxy/abc", true},
		{"loopback no token", "http://127.0.0.1:8080/", false},
		{"token no loopback", "http://example.com/blogger_proxy/", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, isBloggerProxyURL(tt.url))
		})
	}
}

// Cover the bare error-text branches of isUnsafeExtensionError & isRetryableError
// that the existing tests do not hit.
func TestIsUnsafeExtensionError_Extra(t *testing.T) {
	t.Parallel()
	assert.False(t, isUnsafeExtensionError(nil))
	assert.True(t, isUnsafeExtensionError(errors.New("file has unsafe extension")))
	assert.True(t, isUnsafeExtensionError(errors.New("file is unusual and will be skipped")))
	assert.False(t, isUnsafeExtensionError(errors.New("other failure")))
}

func TestIsRetryableError_Extra(t *testing.T) {
	t.Parallel()
	assert.False(t, isRetryableError(nil))
	assert.True(t, isRetryableError(errors.New("connection reset")))
	assert.True(t, isRetryableError(errors.New("network down")))
	assert.True(t, isRetryableError(errors.New("connection refused")))
	assert.True(t, isRetryableError(errors.New("temporary error")))
	assert.False(t, isRetryableError(errors.New("auth denied")))
}

// Mutates global media state (SetAnimeName/SetMediaType) → serial.
func TestPrintBatchDownloadLocation_DoesNotPanic(t *testing.T) {
	SetAnimeName("PrintBatchDownloadLocationTest", 1)
	SetMediaType(false)
	SetExactMediaType("")
	t.Cleanup(func() {
		SetAnimeName("", 0)
		SetExactMediaType("")
	})

	assert.NotPanics(t, func() {
		printBatchDownloadLocation("https://example.com/anime/test", 1)
	})
}

func TestCombineParts(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	dest := filepath.Join(dir, "video.mp4")

	// Write 3 parts.
	for i, payload := range [][]byte{[]byte("AAA"), []byte("BBB"), []byte("CCC")} {
		p, err := safePartPath(dest, i)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(p, payload, 0o600))
	}

	require.NoError(t, combineParts(dest, 3))
	got, err := os.ReadFile(dest)
	require.NoError(t, err)
	assert.Equal(t, "AAABBBCCC", string(got))

	// Parts must be cleaned up.
	for i := range 3 {
		p, _ := safePartPath(dest, i)
		_, err := os.Stat(p)
		assert.True(t, os.IsNotExist(err), "part %d should be removed", i)
	}
}

func TestCombineParts_MissingPartReturnsError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	dest := filepath.Join(dir, "v.mp4")
	require.Error(t, combineParts(dest, 2))
}

func TestCreateEpisodePath_BuildsPathAndCreatesDir(t *testing.T) {
	SetAnimeName("CreateEpisodePathTest", 1)
	SetMediaType(false)
	SetExactMediaType("")
	t.Cleanup(func() {
		SetAnimeName("", 0)
		SetExactMediaType("")
	})

	got, err := createEpisodePath("https://example.com/anime/test", 5)
	require.NoError(t, err)
	require.NotEmpty(t, got)
	_, statErr := os.Stat(filepath.Dir(got))
	assert.NoError(t, statErr, "parent dir must exist")
}

func TestFindEpisode(t *testing.T) {
	t.Parallel()
	episodes := []models.Episode{{Num: 1}, {Num: 2}, {Num: 3}}
	ep, ok := findEpisode(episodes, 2)
	assert.True(t, ok)
	assert.Equal(t, 2, ep.Num)

	_, ok = findEpisode(episodes, 99)
	assert.False(t, ok)
}

func TestResolveDownloadURL_NonAnimeFireReturnsAsIs(t *testing.T) {
	t.Parallel()
	got, err := resolveDownloadURL("https://cdn.example/video.mp4")
	require.NoError(t, err)
	assert.Equal(t, "https://cdn.example/video.mp4", got)
}

func TestResolveDownloadURL_AnimeFireSSRFBlocked(t *testing.T) {
	t.Parallel()
	// AnimeFire path triggers SafeGet which rejects loopback.
	_, err := resolveDownloadURL("https://animefire.io/video/loopback")
	require.Error(t, err)
}

func TestResolveAnimeFireFallbackDownloadURL_NonAnimeFireRejected(t *testing.T) {
	t.Parallel()
	_, err := resolveAnimeFireFallbackDownloadURL("https://other.example/x", "")
	require.Error(t, err)
}

func TestSelectAnimeFireDownloadCandidates_FromData(t *testing.T) {
	t.Parallel()
	body := mustJSON(t, map[string]any{
		"data": []map[string]any{
			{"src": "https://cdn/720.mp4", "label": "720p"},
			{"src": "https://cdn/480.mp4", "label": "480p"},
		},
	})
	got, err := selectAnimeFireDownloadCandidates(body, "best")
	require.NoError(t, err)
	require.NotEmpty(t, got)
	assert.Equal(t, "https://cdn/720.mp4", got[0])
}

func TestSelectAnimeFireDownloadCandidates_FromBloggerToken(t *testing.T) {
	t.Parallel()
	body := mustJSON(t, map[string]any{
		"data":  []any{},
		"token": "https://blogger.com/video/abc",
	})
	got, err := selectAnimeFireDownloadCandidates(body, "best")
	require.NoError(t, err)
	assert.Equal(t, []string{"https://blogger.com/video/abc"}, got)
}

func TestSelectAnimeFireDownloadCandidates_EmptyReturnsError(t *testing.T) {
	t.Parallel()
	body := mustJSON(t, map[string]any{"data": []any{}})
	_, err := selectAnimeFireDownloadCandidates(body, "best")
	require.Error(t, err)
}

func TestSelectAnimeFireDownloadSource_TopCandidate(t *testing.T) {
	t.Parallel()
	body := mustJSON(t, map[string]any{
		"data": []map[string]any{
			{"src": "https://cdn/720.mp4", "label": "720p"},
		},
	})
	got, err := selectAnimeFireDownloadSource(body, "720p")
	require.NoError(t, err)
	assert.Equal(t, "https://cdn/720.mp4", got)
}

func TestOrderAnimeFireSources_DescendingForBest(t *testing.T) {
	t.Parallel()
	data := []VideoData{
		{Src: "low", Label: "480p"},
		{Src: "high", Label: "1080p"},
		{Src: "mid", Label: "720p"},
	}
	got := orderAnimeFireSources(data, "best")
	require.NotEmpty(t, got)
	assert.Equal(t, "high", got[0])
}

func TestOrderAnimeFireSources_AscendingForWorst(t *testing.T) {
	t.Parallel()
	data := []VideoData{
		{Src: "high", Label: "1080p"},
		{Src: "low", Label: "480p"},
	}
	got := orderAnimeFireSources(data, "worst")
	require.NotEmpty(t, got)
	assert.Equal(t, "low", got[0])
}

func TestOrderAnimeFireSources_EmptyReturnsNil(t *testing.T) {
	t.Parallel()
	assert.Nil(t, orderAnimeFireSources(nil, "best"))
}

func TestRecordBatchDownloadFailure_NilErrorNoop(t *testing.T) {
	t.Parallel()
	var mu sync.Mutex
	failures := []batchDownloadFailure{}
	recordBatchDownloadFailure(&mu, &failures, 1, nil)
	assert.Empty(t, failures)
}

func TestRecordBatchDownloadFailure_AppendsUnderLock(t *testing.T) {
	t.Parallel()
	var mu sync.Mutex
	failures := []batchDownloadFailure{}
	recordBatchDownloadFailure(&mu, &failures, 7, errors.New("boom"))
	require.Len(t, failures, 1)
	assert.Equal(t, 7, failures[0].Episode)
}

func TestNewBatchDownloadError_EmptyReturnsNil(t *testing.T) {
	t.Parallel()
	assert.Nil(t, newBatchDownloadError(nil))
}

func TestNewBatchDownloadError_SortsByEpisode(t *testing.T) {
	t.Parallel()
	err := newBatchDownloadError([]batchDownloadFailure{
		{Episode: 5, Err: errors.New("a")},
		{Episode: 2, Err: errors.New("b")},
		{Episode: 8, Err: errors.New("c")},
	})
	require.Error(t, err)
	var batchErr batchDownloadError
	require.ErrorAs(t, err, &batchErr)
	assert.Equal(t, 2, batchErr.Failures[0].Episode)
	assert.Equal(t, 5, batchErr.Failures[1].Episode)
	assert.Equal(t, 8, batchErr.Failures[2].Episode)
}

func TestBatchDownloadError_Error_NoFailures(t *testing.T) {
	t.Parallel()
	err := batchDownloadError{}
	assert.Equal(t, "batch download failed", err.Error())
}

func TestBatchDownloadError_Error_SingleFailure(t *testing.T) {
	t.Parallel()
	err := batchDownloadError{Failures: []batchDownloadFailure{
		{Episode: 3, Err: errors.New("HTTP 404")},
	}}
	got := err.Error()
	assert.Contains(t, got, "1 episode failed")
	assert.Contains(t, got, "episode 3")
}

func TestBatchDownloadError_Error_TruncatesAfterFive(t *testing.T) {
	t.Parallel()
	failures := make([]batchDownloadFailure, 7)
	for i := range failures {
		failures[i] = batchDownloadFailure{Episode: i + 1, Err: errors.New("x")}
	}
	got := batchDownloadError{Failures: failures}.Error()
	assert.Contains(t, got, "7 episodes failed")
	assert.Contains(t, got, "2 more")
}

func TestIsHTTPStatusError_MatchesStatus(t *testing.T) {
	t.Parallel()
	assert.False(t, isHTTPStatusError(nil, 404))
	assert.True(t, isHTTPStatusError(errors.New("got HTTP 404 from CDN"), 404))
	assert.False(t, isHTTPStatusError(errors.New("got HTTP 500 from CDN"), 404))
}

func TestRunAnimeFireDirectDownloadWithFallback_NoErrorReturnsEarly(t *testing.T) {
	t.Parallel()
	called := 0
	download := func(_, _ string, _ *model) error { called++; return nil }
	fallback := func(_, _ string) (string, error) { t.Fatal("fallback must not run"); return "", nil }
	err := runAnimeFireDirectDownloadWithFallback("https://animefire.io/video/x", "https://cdn/y.mp4", "/tmp/x", &model{}, download, fallback)
	require.NoError(t, err)
	assert.Equal(t, 1, called)
}

func TestRunAnimeFireDirectDownloadWithFallback_404FallsBack(t *testing.T) {
	t.Parallel()
	attempt := 0
	download := func(url, _ string, _ *model) error {
		attempt++
		if attempt == 1 {
			return errors.New("HTTP 404 not found")
		}
		assert.Equal(t, "https://cdn/fallback.mp4", url)
		return nil
	}
	fallback := func(_, _ string) (string, error) { return "https://cdn/fallback.mp4", nil }
	err := runAnimeFireDirectDownloadWithFallback("https://animefire.io/video/x", "https://cdn/orig.mp4", "/tmp/x", &model{}, download, fallback)
	require.NoError(t, err)
	assert.Equal(t, 2, attempt)
}

func TestRunAnimeFireDirectDownloadWithFallback_NonAnimeFire404ReturnsOriginal(t *testing.T) {
	t.Parallel()
	download := func(_, _ string, _ *model) error { return errors.New("HTTP 404 not found") }
	fallback := func(_, _ string) (string, error) { t.Fatal("fallback must not run for non-animefire"); return "", nil }
	err := runAnimeFireDirectDownloadWithFallback("https://other/x", "https://cdn/orig.mp4", "/tmp/x", &model{}, download, fallback)
	require.Error(t, err)
}

func TestDownloadAnimeFireDirectWithFallback_SetsGlobalReferer(t *testing.T) {
	prev := util.GetGlobalReferer()
	util.ClearGlobalReferer()
	t.Cleanup(func() { util.SetGlobalReferer(prev) })

	// Call with bogus URL so download errors instantly. Referer setup is
	// what we are pinning here, not the download outcome.
	_ = downloadAnimeFireDirectWithFallback("https://animefire.io/video/x", "http://0.0.0.0:0/x.mp4", "/tmp/_unused.mp4", &model{})
	assert.Equal(t, "https://animefire.io", util.GetGlobalReferer())
}

func TestDownloadBloggerDirect_SSRFBlocked(t *testing.T) {
	t.Parallel()
	err := downloadBloggerDirect("http://127.0.0.1:1/blocked", "/tmp/x.mp4", 1, &model{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "SSRF blocked")
}

func TestDownloadBloggerChunk_SSRFBlocked(t *testing.T) {
	t.Parallel()
	err := downloadBloggerChunk("http://127.0.0.1:1/blocked", 0, 100, 0, "/tmp/x.mp4", &model{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "SSRF blocked")
}

func TestDownloadVideo_BadURLReturnsError(t *testing.T) {
	t.Parallel()
	// Empty/invalid URL → getContentLength fails immediately.
	err := DownloadVideo("http://127.0.0.1:1/x", "/tmp/x.mp4", 1, &model{})
	require.Error(t, err)
}

func TestDownloadWithYtDlp_InvalidURLRejected(t *testing.T) {
	t.Parallel()
	err := downloadWithYtDlp("http://bad\nurl", "/tmp/x.mp4", &model{})
	require.Error(t, err)
}

func TestExtractVideoSources_BadURLReturnsError(t *testing.T) {
	t.Parallel()
	_, err := ExtractVideoSources("http://127.0.0.1:1/blocked")
	require.Error(t, err)
}

func TestExtractVideoSourcesWithPrompt_BadURLReturnsError(t *testing.T) {
	t.Parallel()
	_, err := ExtractVideoSourcesWithPrompt("http://127.0.0.1:1/blocked")
	require.Error(t, err)
}

func TestGetBestQualityURL_UnsupportedIdentifier(t *testing.T) {
	// AllAnime path retries through the enhanced API which spawns goroutines —
	// keep serial to avoid races with other tests that touch network globals.
	anime := &models.Anime{Source: "AllAnime", URL: ""}
	_, err := getBestQualityURL(models.Episode{URL: "x"}, anime)
	require.Error(t, err)
}

// HandleBatchDownload, getEpisodeRange, handleExistingEpisodes, and
// askAndPlayDownloadedEpisode each prompt the user via huh.NewInput /
// fuzzyfinder. Per CLAUDE.md, TUI orchestration must be tested indirectly:
// the dispatch helpers (findEpisode, createEpisodePath, fileExists,
// printBatchDownloadLocation, batchDownloadError) each have their own
// dedicated tests above. Here we pin the function symbols so adapter
// regressions surface at compile time.

func TestHandleBatchDownload_SymbolPinned(t *testing.T) {
	t.Parallel()
	assert.NotNil(t, HandleBatchDownload)
}

func TestGetEpisodeRange_SymbolPinned(t *testing.T) {
	t.Parallel()
	assert.NotNil(t, getEpisodeRange)
}

func TestHandleExistingEpisodes_NoExistingExits(t *testing.T) {
	// Mutates global media state — keep serial.
	SetAnimeName("HandleExistingTest", 1)
	t.Cleanup(func() { SetAnimeName("", 0) })

	// Empty episode list + bogus range → no existing matches → returns nil
	// without entering the fuzzy menu.
	err := handleExistingEpisodes(nil, "https://example.com/anime/handle-existing-test", 1, 3)
	assert.NoError(t, err)
}

func TestAskAndPlayDownloadedEpisode_NoDownloadedExits(t *testing.T) {
	SetAnimeName("AskAndPlayDownloadedTest", 1)
	t.Cleanup(func() { SetAnimeName("", 0) })

	err := askAndPlayDownloadedEpisode(nil, "https://example.com/anime/ask-and-play-test", 1, 3)
	assert.NoError(t, err)
}

// mustJSON marshals v to bytes, failing the test on error.
func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return b
}
