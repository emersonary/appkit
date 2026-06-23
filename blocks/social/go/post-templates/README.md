# Social post caption templates

Each `*.template.txt` file is **caption body only** — no comments or layout notes.
Documentation and platform tips live here, not in the template files.

## Replaceable fields

| Field | Source |
|-------|--------|
| `{{title}}` | Post title |
| `{{intro_text}}` | Summary / hook (translation) |
| `{{article_url}}` | Tracking redirect URL (`/r/{code}`) |
| `{{link_line}}` | Language-aware CTA with brand + tracking URL |
| `{{source_brand}}` | Project display name |
| `{{source_url}}` | Public site base URL |
| `{{hero_image_url}}` | First inline image or video thumbnail (metadata only) |
| `{{hashtags}}` | Optional hashtags (omit when empty) |

## LinkedIn (`li`)

- Link post: LinkedIn auto-generates preview from `{{article_url}}` — ensure OG tags on the public article page.
- Optional: upload hero image as a document-style image for higher feed presence (not in caption text).
- Intro = hook; `{{link_line}}` = clear CTA.
- Full tracking URL in the body is clickable on LinkedIn.

## Instagram / Facebook

- Prefer hero image in carousel or link preview; caption = intro + link line + optional hashtags.

## Limits (approximate)

- LinkedIn: ~3000 characters
- Instagram: 2200 characters
- Threads: 500 characters per post
- TikTok: 2200 characters (keep short)
