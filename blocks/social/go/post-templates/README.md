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

Caption layout is always `intro_text`, optional `link_line`, optional `hashtags` (same template file).

**Link mode** (long post with inline images, code blocks, or similar):

- `intro_text` = summary / hook
- `link_line` = language-aware CTA with tracking URL
- LinkedIn API attaches the article link preview

**Native mode** (short posts, or long plain text without rich media):

- `intro_text` = full post body as plain text (HTML stripped)
- `link_line` = empty (omitted from caption)
- No article link attachment on LinkedIn

Mode is chosen automatically in Go (`resolveLinkedInMode`); there is no per-post override.

- Short posts (≤1200 characters plain text) → always native
- Long posts with `<img>`, `<pre>`, `<code>`, `<video>`, or `<iframe>` → link
- Otherwise → native

Cover/header image is passed as publish metadata (`cover_image_url` / `hero_image_url`), not in caption text. Prefer post cover, then first inline image, then video thumbnail.

- Link post: LinkedIn auto-generates preview from `{{article_url}}` — ensure OG tags on the public article page.
- Full tracking URL in link mode is clickable on LinkedIn.

## Instagram / Facebook

- Prefer hero image in carousel or link preview; caption = intro + link line + optional hashtags.

## Limits (approximate)

- LinkedIn: ~3000 characters
- Instagram: 2200 characters
- Threads: 500 characters per post
- TikTok: 2200 characters (keep short)
