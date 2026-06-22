package resolver

import (
	"errors"
	"testing"
)

func TestIsBloggerURL(t *testing.T) {
	yes := []string{
		"https://www.blogger.com/video.g?token=ABC",
		"https://BLOGGER.com/video.g?token=x",
		"https://x.blogspot.com/video.g?token=y",
	}
	no := []string{
		"https://lightspeedst.net/s2/mp4/x/sd/1.mp4",
		"https://repackager.wixmp.com/x.m3u8",
		"",
	}
	for _, u := range yes {
		if !IsBloggerURL(u) {
			t.Errorf("IsBloggerURL(%q) = false, want true", u)
		}
	}
	for _, u := range no {
		if IsBloggerURL(u) {
			t.Errorf("IsBloggerURL(%q) = true, want false", u)
		}
	}
}

func TestIsAnimeFireVideoURL(t *testing.T) {
	if !IsAnimeFireVideoURL("https://animefire.io/video/slug/1") {
		t.Error("expected animefire video URL to match")
	}
	if IsAnimeFireVideoURL("https://lightspeedst.net/x.mp4") {
		t.Error("direct mp4 should not match")
	}
}

func TestPickQuality(t *testing.T) {
	data := []afSource{
		{Src: "low", Label: "360p"},
		{Src: "mid", Label: "720p"},
		{Src: "high", Label: "1080p"},
	}
	if got := pickQuality(data, "best"); got != "high" {
		t.Errorf("best = %q, want high", got)
	}
	if got := pickQuality(data, "worst"); got != "low" {
		t.Errorf("worst = %q, want low", got)
	}
	if got := pickQuality(data, ""); got != "high" {
		t.Errorf("default = %q, want high", got)
	}
}

func TestLabelN(t *testing.T) {
	cases := map[string]int{"360p": 360, "1080p": 1080, "HD": 0, "720": 720}
	for in, want := range cases {
		if got := labelN(in); got != want {
			t.Errorf("labelN(%q) = %d, want %d", in, got, want)
		}
	}
}

func TestParseBatchexecuteEmptyMeansUnavailable(t *testing.T) {
	// HTTP 200 with only the anti-hijacking prefix → token dead upstream.
	_, err := parseBatchexecuteResponse([]byte(")]}'\n\n"))
	if !errors.Is(err, ErrBloggerVideoUnavailable) {
		t.Errorf("empty body err = %v, want ErrBloggerVideoUnavailable", err)
	}
}

func TestParseBatchexecuteRegexFallback(t *testing.T) {
	body := []byte(`)]}'
[["wrb.fr","WcwnYd","[]"]]
garbage https://rr1---sn-x.googlevideo.com/videoplayback?itag=22&mime=video/mp4 more`)
	got, err := parseBatchexecuteResponse(body)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got == "" || got[:8] != "https://" {
		t.Errorf("expected googlevideo URL, got %q", got)
	}
}

func TestResolvePassthrough(t *testing.T) {
	direct, ref, err := Resolve("https://lightspeedst.net/x.mp4", "best")
	if err != nil || direct != "https://lightspeedst.net/x.mp4" || ref != "" {
		t.Errorf("passthrough failed: %q %q %v", direct, ref, err)
	}
}
