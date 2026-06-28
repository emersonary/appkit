package social

import "testing"

func TestResolveLinkedInModeShortAlwaysNative(t *testing.T) {
	input := PostInput{
		IntroText: "Summary hook.",
		BodyHTML:  "<p>Short plain post.</p>",
	}
	if got := resolveLinkedInMode(input); got != linkedInModeNative {
		t.Fatalf("short post: got %v want native", got)
	}
}

func TestResolveLinkedInModeShortWithImageStillNative(t *testing.T) {
	input := PostInput{
		IntroText: "Summary hook.",
		BodyHTML:  "<p>Short post with image.</p><img src=\"https://example.com/a.jpg\" alt=\"\">",
	}
	if got := resolveLinkedInMode(input); got != linkedInModeNative {
		t.Fatalf("short post with image: got %v want native", got)
	}
}

func TestResolveLinkedInModeRichContentUsesLink(t *testing.T) {
	body := "<p>" + repeat("Long paragraph. ", 200) + "</p><img src=\"https://example.com/a.jpg\" alt=\"\">"
	input := PostInput{
		IntroText: "Summary hook.",
		BodyHTML:  body,
	}
	if got := resolveLinkedInMode(input); got != linkedInModeLink {
		t.Fatalf("rich long post: got %v want link", got)
	}
}

func TestResolveLinkedInModeLongPlainUsesNative(t *testing.T) {
	body := "<p>" + repeat("Plain paragraph without media. ", 200) + "</p>"
	input := PostInput{
		IntroText: "Summary hook.",
		BodyHTML:  body,
	}
	if got := resolveLinkedInMode(input); got != linkedInModeNative {
		t.Fatalf("long plain post: got %v want native", got)
	}
}

func TestLinkedInNativeCaptionOmitsLinkLine(t *testing.T) {
	renderer, err := NewTemplateRenderer()
	if err != nil {
		t.Fatal(err)
	}
	input := PostInput{
		Title:       "Title",
		IntroText:   "Summary should not appear.",
		BodyHTML:    "<p>Full native body text.</p>",
		ArticleURL:  "https://example.com/r/abc123",
		SourceBrand: "Brand",
		SourceURL:   "https://example.com",
		Language:    "en",
		Hashtags:    "#go",
	}
	fields := applyLinkedInMode(BuildFields(input, PlatformConfig{}), input, linkedInModeNative)
	out, err := renderer.Render(PlatformLinkedIn, fields)
	if err != nil {
		t.Fatal(err)
	}
	want := "Full native body text.\n\n#go"
	if out != want {
		t.Fatalf("got %q want %q", out, want)
	}
}

func TestHtmlToPlainText(t *testing.T) {
	got := htmlToPlainText("<p>Hello <strong>world</strong>.</p><pre>code</pre>")
	want := "Hello world.\ncode"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func repeat(s string, n int) string {
	out := make([]byte, 0, len(s)*n)
	for i := 0; i < n; i++ {
		out = append(out, s...)
	}
	return string(out)
}
