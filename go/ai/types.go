package ai

// DetectResponse is the detected language for input text.
type DetectResponse struct {
	Language   string
	Confidence float64
	Provider   string
	Model      string
}

// TranslateRequest is input for plain or HTML translation.
// When Text contains HTML tags, providers preserve markup automatically.
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

// ChatMessage is one turn in a chat request or response.
type ChatMessage struct {
	Role    string
	Content string
}

// ChatRequest is input for a single chat completion turn.
type ChatRequest struct {
	SystemPrompt string
	Messages     []ChatMessage
}

// ChatResponse is a chat completion result.
type ChatResponse struct {
	Message  ChatMessage
	Provider string
	Model    string
}
