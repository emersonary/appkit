package ai

// TranslateRequest is input for plain text translation.
type TranslateRequest struct {
	Text           string
	SourceLanguage string // optional; empty means auto-detect
	TargetLanguage string // required ISO 639-1 code
}

// TranslateResponse is translated text metadata.
type TranslateResponse struct {
	Text           string
	SourceLanguage string
	TargetLanguage string
	Provider       string
	Model          string
}
