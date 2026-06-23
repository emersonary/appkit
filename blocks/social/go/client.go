package social

import "context"

// PlatformClient formats and publishes content for one social network.
type PlatformClient interface {
	PlatformID() PlatformID
	DefaultDispatch() DispatchMode
	FormatPost(ctx context.Context, input PostInput) (FormattedPost, error)
	CreatePost(ctx context.Context, req CreatePostRequest) (CreatePostResult, error)
	GetPost(ctx context.Context, postID string) (PostInfo, error)
	GetAccountInfo(ctx context.Context) (AccountInfo, error)
}
