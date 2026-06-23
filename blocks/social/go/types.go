package social

// PostInput is the canonical blog post payload used to format platform posts.
type PostInput struct {
	Title        string
	IntroText    string
	ArticleURL   string
	SourceBrand  string
	SourceURL    string
	HeroImageURL string
	VideoURL     string
	Hashtags     string
	Language     string
	ContentKind  string // article | video
	TrackingCode string
}

// TemplateFields are replaceable template variables per platform.
type TemplateFields map[string]string

// FormattedPost is the rendered platform-specific content ready to publish.
type FormattedPost struct {
	PlatformID PlatformID
	Caption    string
	Title      string
	LinkURL    string
	MediaURLs  []string
	VideoURL   string
	Fields     TemplateFields
}

// CreatePostRequest is the publish payload after formatting.
type CreatePostRequest struct {
	Input        PostInput
	Formatted    FormattedPost
	DispatchMode DispatchMode
}

// ClientPublishJob is returned when dispatch is client-side (browser / mobile app).
type ClientPublishJob struct {
	PlatformID   PlatformID
	Caption      string
	Title        string
	LinkURL      string
	MediaURLs    []string
	VideoURL     string
	Instructions string
	Payload      map[string]any
}

// CreatePostResult is the outcome of CreatePost.
type CreatePostResult struct {
	DispatchMode DispatchMode
	PostID       string
	PublishedURL string
	ClientJob    *ClientPublishJob
}

// PostInfo is a published post snapshot from the platform API.
type PostInfo struct {
	PlatformID   PlatformID
	PostID       string
	PublishedURL string
	Caption      string
	Status       string
}

// AccountInfo is basic account metadata from the platform API.
type AccountInfo struct {
	PlatformID  PlatformID
	AccountID   string
	DisplayName string
	Username    string
	Followers   int64
}

// PublishRequest publishes one canonical post to selected platforms.
type PublishRequest struct {
	Input       PostInput
	PlatformIDs []PlatformID
	Dispatch    DispatchMode // empty = per-platform default
}

// PublishResult is the per-platform publish outcome.
type PublishResult struct {
	PlatformID PlatformID
	Result     CreatePostResult
	Err        error
}
