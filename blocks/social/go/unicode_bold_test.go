package social

import "testing"

func TestToSansSerifBold(t *testing.T) {
	got := ToSansSerifBold("Hello World 2026!")
	want := "𝗛𝗲𝗹𝗹𝗼 𝗪𝗼𝗿𝗹𝗱 𝟮𝟬𝟮𝟲!"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestToSansSerifBoldPreservesNonASCII(t *testing.T) {
	got := ToSansSerifBold("Café — naïve")
	want := "𝗖𝗮𝗳é — 𝗻𝗮ï𝘃𝗲"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
