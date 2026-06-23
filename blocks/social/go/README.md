# Social publishing

Platform-agnostic library for formatting and publishing blog posts to Instagram, Facebook, Threads, Pinterest, TikTok, LinkedIn, and YouTube.

## Features

- **7 platform clients** with `FormatPost`, `CreatePost`, `GetPost`, and `GetAccountInfo`
- **Embedded caption templates** (`post-templates/*.template.txt`) with intro text + source brand link
- **Server vs client dispatch** — server calls platform APIs; client returns a `ClientPublishJob` for browser/mobile publishing
- **YAML config** with `access_token` or `access_token_env` placeholders

## Config

Copy `social.example.yaml` and set `enabled: true` after filling credentials:

```yaml
social:
  enabled: true
  config_path: config/social.yaml
```

Or inline in the posts API config (`internal/config.Config`):

```yaml
social:
  enabled: true
  default_dispatch: server
  platforms:
    li:
      driver: linkedin
```

Per-project credentials live under each feed entry in `tenants.feed[].metadata.social` (see posts `config.yaml`).

## Usage

```go
input := social.PostInput{
    Title:       "Guia secreto de Jeri",
    IntroText:   translation.Summary,
    ArticleURL:  baseURL + "/r/" + trackingCode,
    SourceBrand: "Emersonary",
    SourceURL:   "emersonary.com",
    HeroImageURL: heroURL,
    Hashtags:    "#jericoacoara",
}

results := postsApp.Social.PublishToPlatforms(ctx, social.PublishRequest{
    Input: input,
    PlatformIDs: []social.PlatformID{social.PlatformInstagram, social.PlatformFacebook},
})
```

## Platform APIs

| ID | Driver | API |
|----|--------|-----|
| ig | instagram | Meta Graph API — media + media_publish |
| fb | facebook | Meta Graph API — page feed |
| th | threads | Meta Graph API — threads |
| pi | pinterest | Pinterest API v5 — pins |
| tt | tiktok | TikTok Content Posting API |
| li | linkedin | LinkedIn REST posts |
| yt | youtube | YouTube Data API v3 (community → client dispatch by default) |

TikTok and YouTube default to **client dispatch** in the example config when video/community workflows need manual steps.
