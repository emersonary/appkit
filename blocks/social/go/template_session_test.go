package social

import (
	"testing"
)

func TestParseTemplateGlobalAndSessions(t *testing.T) {
	raw := `# comment
{{title}}

@session type=link
{{intro_text}}
@end

@session type=native
{{full_text}}
@end
`
	parsed, err := parseTemplate(raw)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Global != "{{title}}" {
		t.Fatalf("global = %q", parsed.Global)
	}
	if len(parsed.Sessions) != 2 {
		t.Fatalf("sessions = %d", len(parsed.Sessions))
	}
	if parsed.Sessions[0].Filters["type"] != "link" {
		t.Fatalf("first session type = %q", parsed.Sessions[0].Filters["type"])
	}
}

func TestComposeTemplateBodyMatchesLinkedInLinkSession(t *testing.T) {
	parsed, err := parseTemplate(`{{title_bold}}

@session type=link
{{intro_text}}

{{link_line}}
@end

@session type=native
{{full_text}}
@end
`)
	if err != nil {
		t.Fatal(err)
	}
	body := composeTemplateBody(parsed, TemplateContext{Filters: map[string]string{"type": "link"}})
	if body != "{{title_bold}}\n\n{{intro_text}}\n\n{{link_line}}" {
		t.Fatalf("got %q", body)
	}
}

func TestRenderLinkedInNativeSession(t *testing.T) {
	renderer, err := NewTemplateRenderer()
	if err != nil {
		t.Fatal(err)
	}
	fields := TemplateFields{
		"title_bold":   ToSansSerifBold("My title"),
		"full_text":    "Full native body text.",
		"hashtags":     "#go",
		"read_on_line": "Read on Brand: https://example.com/r/abc123",
	}
	out, err := renderer.RenderWithContext(PlatformLinkedIn, fields, TemplateContext{
		Platform: PlatformLinkedIn,
		Filters:  map[string]string{"type": "native"},
	})
	if err != nil {
		t.Fatal(err)
	}
	want := ToSansSerifBold("My title") + "\n\nFull native body text.\n\n#go\n\nRead on Brand: https://example.com/r/abc123"
	if out != want {
		t.Fatalf("got %q want %q", out, want)
	}
}

func TestParseTemplateRejectsNestedSession(t *testing.T) {
	_, err := parseTemplate("@session type=link\n@session type=native\n@end\n@end\n")
	if err == nil {
		t.Fatal("expected nested session error")
	}
}
