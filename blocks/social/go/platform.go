package social

import "strings"

// PlatformID identifies a social network (matches posts.social_platforms).
type PlatformID string

const (
	PlatformInstagram PlatformID = "ig"
	PlatformFacebook  PlatformID = "fb"
	PlatformThreads   PlatformID = "th"
	PlatformPinterest PlatformID = "pi"
	PlatformTikTok    PlatformID = "tt"
	PlatformLinkedIn  PlatformID = "li"
	PlatformYouTube   PlatformID = "yt"
)

// DefaultPlatforms is the canonical catalog shared with the posts project.
var DefaultPlatforms = []PlatformID{
	PlatformInstagram,
	PlatformFacebook,
	PlatformThreads,
	PlatformPinterest,
	PlatformTikTok,
	PlatformLinkedIn,
	PlatformYouTube,
}

// PlatformMeta describes a platform for UI and logging.
type PlatformMeta struct {
	ID          PlatformID
	Name        string
	DefaultDriver string
}

var platformCatalog = map[PlatformID]PlatformMeta{
	PlatformInstagram: {ID: PlatformInstagram, Name: "Instagram", DefaultDriver: "instagram"},
	PlatformFacebook:  {ID: PlatformFacebook, Name: "Facebook", DefaultDriver: "facebook"},
	PlatformThreads:   {ID: PlatformThreads, Name: "Threads", DefaultDriver: "threads"},
	PlatformPinterest: {ID: PlatformPinterest, Name: "Pinterest", DefaultDriver: "pinterest"},
	PlatformTikTok:    {ID: PlatformTikTok, Name: "TikTok", DefaultDriver: "tiktok"},
	PlatformLinkedIn:  {ID: PlatformLinkedIn, Name: "LinkedIn", DefaultDriver: "linkedin"},
	PlatformYouTube:   {ID: PlatformYouTube, Name: "YouTube", DefaultDriver: "youtube"},
}

// ParsePlatformID validates a platform id string.
func ParsePlatformID(raw string) (PlatformID, error) {
	id := PlatformID(strings.TrimSpace(raw))
	if _, ok := platformCatalog[id]; !ok {
		return "", invalidConfigf("platform", "unknown platform id %q", raw)
	}
	return id, nil
}

// PlatformMetaFor returns metadata for a platform id.
func PlatformMetaFor(id PlatformID) (PlatformMeta, bool) {
	meta, ok := platformCatalog[id]
	return meta, ok
}
