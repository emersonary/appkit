package ai

import "testing"

func TestTranslationLooksLikeHTML(t *testing.T) {
	if translationLooksLikeHTML("Hello world") {
		t.Fatal("plain text should not match")
	}
	if !translationLooksLikeHTML("<p>Hello</p>") {
		t.Fatal("expected HTML tag")
	}
}
