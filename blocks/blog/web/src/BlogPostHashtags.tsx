export type BlogPostHashtagsProps = {
  hashtags?: string;
  className?: string;
};

function parseHashtags(raw: string): string[] {
  return raw.split(/\s+/).map((tag) => tag.trim()).filter(Boolean);
}

export function BlogPostHashtags({
  hashtags,
  className = "blog-post-view__hashtags",
}: BlogPostHashtagsProps) {
  const text = hashtags?.trim();
  if (!text) {
    return null;
  }

  const tags = parseHashtags(text);
  if (tags.length === 0) {
    return null;
  }

  return (
    <div className={className} aria-label="Hashtags">
      {tags.map((tag, index) => (
        <span key={`${tag}-${index}`} className="blog-post-hashtag">
          {tag}
        </span>
      ))}
    </div>
  );
}
