package social

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func testSocialConfig(t *testing.T) SocialConfig {
	t.Helper()
	os.Setenv("META_ACCESS_TOKEN", "meta-test-token")
	os.Setenv("PINTEREST_ACCESS_TOKEN", "pin-test-token")
	os.Setenv("TIKTOK_ACCESS_TOKEN", "tt-test-token")
	os.Setenv("LINKEDIN_ACCESS_TOKEN", "li-test-token")
	os.Setenv("YOUTUBE_ACCESS_TOKEN", "yt-test-token")
	t.Cleanup(func() {
		os.Unsetenv("META_ACCESS_TOKEN")
		os.Unsetenv("PINTEREST_ACCESS_TOKEN")
		os.Unsetenv("TIKTOK_ACCESS_TOKEN")
		os.Unsetenv("LINKEDIN_ACCESS_TOKEN")
		os.Unsetenv("YOUTUBE_ACCESS_TOKEN")
	})

	enabled := true
	return SocialConfig{
		Enabled:         true,
		DefaultDispatch: string(DispatchServer),
		Platforms: map[string]PlatformConfig{
			"ig": {
				Enabled:        &enabled,
				AccountID:      "ig-user",
				AccessTokenEnv: "META_ACCESS_TOKEN",
				APIBaseURL:     "https://graph.test",
			},
			"fb": {
				Enabled:        &enabled,
				AccountID:      "page-1",
				AccessTokenEnv: "META_ACCESS_TOKEN",
				APIBaseURL:     "https://graph.test",
			},
		},
	}
}

func samplePostInput() PostInput {
	return PostInput{
		Title:        "Guia secreto de Jeri",
		IntroText:    "Descobri um lugar incrível em Jericoacoara.",
		ArticleURL:   "https://emersonary.com/r/x7k2m9",
		SourceBrand:  "Emersonary",
		SourceURL:    "emersonary.com",
		Language:     "pt",
		Hashtags:     "#jericoacoara #viagem",
		HeroImageURL: "https://cdn.example/hero.jpg",
	}
}

func TestSocialConfigRequiresTokenWhenEnabled(t *testing.T) {
	cfg := testSocialConfig(t)
	os.Unsetenv("META_ACCESS_TOKEN")

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error for missing token")
	}
}

func TestSocialConfigValidateDefaultsSkipsCredentials(t *testing.T) {
	enabled := true
	cfg := SocialConfig{
		Enabled:         true,
		DefaultDispatch: string(DispatchServer),
		Platforms: map[string]PlatformConfig{
			"li": {
				Enabled: &enabled,
				Driver:  "linkedin",
			},
		},
	}
	if err := cfg.ValidateDefaults(); err != nil {
		t.Fatalf("ValidateDefaults: %v", err)
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected full Validate to require credentials")
	}
}

func TestTemplateRendererInstagram(t *testing.T) {
	renderer, err := NewTemplateRenderer()
	if err != nil {
		t.Fatal(err)
	}
	fields := BuildFields(samplePostInput(), PlatformConfig{})
	out, err := renderer.Render(PlatformInstagram, fields)
	if err != nil {
		t.Fatal(err)
	}
	if !containsAll(out, []string{
		"Descobri um lugar incrível",
		"Leia no Emersonary:",
		"https://emersonary.com/r/x7k2m9",
		"#jericoacoara",
	}) {
		t.Fatalf("unexpected caption: %q", out)
	}
	if containsAny(out, []string{"Replaceable fields:", "VISUAL LAYOUT", "MEDIA (recommendation)"}) {
		t.Fatalf("template comments leaked into caption: %q", out)
	}
}

func TestTemplateRendererLinkedIn(t *testing.T) {
	renderer, err := NewTemplateRenderer()
	if err != nil {
		t.Fatal(err)
	}
	input := samplePostInput()
	input.Language = "en"
	fields := BuildFields(input, PlatformConfig{})
	out, err := renderer.RenderWithContext(PlatformLinkedIn, fields, TemplateContext{
		Platform: PlatformLinkedIn,
		Filters:  map[string]string{"type": "link"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !containsAll(out, []string{
		"Descobri um lugar incrível",
		"Read on Emersonary:",
		"https://emersonary.com/r/x7k2m9",
		"#jericoacoara",
	}) {
		t.Fatalf("unexpected caption: %q", out)
	}
	if containsAny(out, []string{"Replaceable fields:", "Optional:", "LIMITS", "hero_image_url", "localhost"}) {
		t.Fatalf("template noise leaked into caption: %q", out)
	}
}

func TestTemplateRendererLinkedInEmptyHashtags(t *testing.T) {
	renderer, err := NewTemplateRenderer()
	if err != nil {
		t.Fatal(err)
	}
	fields := BuildFields(PostInput{
		Title:       "Title",
		IntroText:   "Hook line.",
		ArticleURL:  "https://example.com/r/abc123",
		SourceBrand: "Brand",
		SourceURL:   "https://example.com",
		Language:    "en",
	}, PlatformConfig{})
	out, err := renderer.RenderWithContext(PlatformLinkedIn, fields, TemplateContext{
		Platform: PlatformLinkedIn,
		Filters:  map[string]string{"type": "link"},
	})
	if err != nil {
		t.Fatal(err)
	}
	want := ToSansSerifBold("Title") + "\n\nHook line.\n\nRead on Brand: https://example.com/r/abc123"
	if out != want {
		t.Fatalf("got %q want %q", out, want)
	}
}

func TestInstagramCreatePostServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/media_publish"):
			_, _ = w.Write([]byte(`{"id":"published-1"}`))
		case r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/media"):
			_, _ = w.Write([]byte(`{"id":"container-1"}`))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	cfg := testSocialConfig(t)
	cfg.Platforms["ig"] = PlatformConfig{
		AccountID:      "ig-user",
		AccessTokenEnv: "META_ACCESS_TOKEN",
		APIBaseURL:     server.URL,
	}

	templates, err := NewTemplateRenderer()
	if err != nil {
		t.Fatal(err)
	}
	client, err := buildPlatformClient(PlatformInstagram, cfg.Platforms["ig"], templates, DispatchServer, server.Client())
	if err != nil {
		t.Fatal(err)
	}

	result, err := client.CreatePost(context.Background(), CreatePostRequest{Input: samplePostInput()})
	if err != nil {
		t.Fatal(err)
	}
	if result.PostID != "published-1" {
		t.Fatalf("expected published-1, got %q", result.PostID)
	}
}

func TestPublishClientDispatch(t *testing.T) {
	cfg := testSocialConfig(t)
	cfg.Platforms["tt"] = PlatformConfig{
		Dispatch:       string(DispatchClient),
		AccountID:      "open-1",
		AccessTokenEnv: "TIKTOK_ACCESS_TOKEN",
		APIBaseURL:     "https://tiktok.test",
	}

	templates, err := NewTemplateRenderer()
	if err != nil {
		t.Fatal(err)
	}
	client, err := buildPlatformClient(PlatformTikTok, cfg.Platforms["tt"], templates, DispatchServer, nil)
	if err != nil {
		t.Fatal(err)
	}

	result, err := client.CreatePost(context.Background(), CreatePostRequest{Input: samplePostInput()})
	if err != nil {
		t.Fatal(err)
	}
	if result.DispatchMode != DispatchClient || result.ClientJob == nil {
		t.Fatalf("expected client job, got %+v", result)
	}
}

func containsAll(s string, parts []string) bool {
	for _, p := range parts {
		if !strings.Contains(s, p) {
			return false
		}
	}
	return true
}

func containsAny(s string, parts []string) bool {
	for _, p := range parts {
		if strings.Contains(s, p) {
			return true
		}
	}
	return false
}
