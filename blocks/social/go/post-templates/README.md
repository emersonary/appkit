# Social post caption templates

Each `*.template.txt` file is **caption body only**. Lines starting with `#` are comments (never published).

## Replaceable fields

| Field | Source |
|-------|--------|
| `{{title}}` | Post title |
| `{{intro_text}}` | Summary / hook (translation) |
| `{{full_text}}` | Full post body as plain text (HTML stripped) |
| `{{article_url}}` | Tracking redirect URL (`/r/{code}`) |
| `{{link_line}}` | Language-aware CTA with brand + tracking URL (link mode) |
| `{{read_on_line}}` | Same CTA text, for native mode footer placement |
| `{{source_brand}}` | Project display name |
| `{{source_url}}` | Public site base URL |
| `{{hero_image_url}}` | First inline image or video thumbnail (metadata only) |
| `{{hashtags}}` | Optional hashtags (omit when empty) |

## Sessions

Templates may define **sessions**: conditional blocks rendered only when a filter matches.

```text
{{title}}

@session type=link
{{intro_text}}

{{link_line}}

{{hashtags}}
@end

@session type=native
{{full_text}}

{{hashtags}}

{{read_on_line}}
@end
```

- Content **outside** `@session` … `@end` is always rendered (global).
- `@session type=link` renders only when the render context has `type=link`.
- Empty placeholders are still removed after render.

## LinkedIn (`li`)

Mode is resolved in Go (`resolveLinkedInMode`); the template selects layout via sessions.

**Link mode** (long post with inline images, code blocks, or similar):

- Global: `{{title}}`
- Session `type=link`: intro + link line + hashtags
- LinkedIn API attaches the article link preview

**Native mode** (short posts, or long plain text without rich media):

- Global: `{{title}}`
- Session `type=native`: full body + hashtags + read-on line
- No article link attachment on LinkedIn

Rules:

- Short posts (≤1200 characters plain text) → always native
- Long posts with `<img>`, `<pre>`, `<code>`, `<video>`, or `<iframe>` → link
- Otherwise → native

Cover/header image is passed as publish metadata (`cover_image_url` / `hero_image_url`), not in caption text.

## Instagram / Facebook

Flat templates (no sessions): intro + link line + optional hashtags.

## Limits (approximate)

- LinkedIn: ~3000 characters
- Instagram: 2200 characters
- Threads: 500 characters per post
- TikTok: 2200 characters (keep short)
