package config

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestHandler_GetConfig_MatchesBaseExampleJSON(t *testing.T) {
	cfg := loadYAMLFixture[BaseConfig](t, baseExampleYAML, LoadOptions{
		DefaultAppName: "fallback api",
	})

	mux := http.NewServeMux()
	NewHandler(func() any { return cfg }).RegisterRoutes(mux, "GET /config")

	req := httptest.NewRequest(http.MethodGet, "/config", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d want %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("content-type: got %q", ct)
	}

	got := normalizeJSON(t, rec.Body.Bytes())
	want := normalizeJSON(t, baseExampleJSON)

	if !bytes.Equal(got, want) {
		t.Fatalf("GET /config body mismatch\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func normalizeJSON(t *testing.T, raw []byte) []byte {
	t.Helper()

	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		t.Fatalf("unmarshal json: %v\nbody: %s", err, raw)
	}

	out, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}

	return out
}

func TestGenBaseExampleJSON(t *testing.T) {
	if os.Getenv("GEN_CONFIG_JSON") == "" {
		t.Skip("set GEN_CONFIG_JSON=1 to regenerate testdata/base.example.json")
	}

	cfg := loadYAMLFixture[BaseConfig](t, baseExampleYAML, LoadOptions{
		DefaultAppName: "fallback api",
	})

	body, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	body = append(body, '\n')

	if err := os.WriteFile("testdata/base.example.json", body, 0o644); err != nil {
		t.Fatal(err)
	}
}
