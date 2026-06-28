import type { ReactNode } from "react";
import { BlogPostHashtags } from "./BlogPostHashtags";
import { normalizeBlogContentKind } from "./contentKind";
import { isTwoColumnLayout } from "./layout";
import "./BlogPostView.css";

export type BlogPostViewProps = {
  language: string;
  title: string;
  bodyHtml: string;
  contentKind: string;
  layoutType: string;
  coverImageUrl?: string;
  videoUrl?: string;
  videoProvider?: string;
  publishAt?: string;
  variant?: "public" | "preview";
  breadcrumb?: ReactNode;
  backLink?: ReactNode;
  emptyMessage?: string;
  hashtagsSuffix?: string;
  /** Shown in the preview badge (e.g. a language label component). */
  previewLanguageLabel?: ReactNode;
  untitledPreviewLabel?: string;
};

function parseYoutubeId(url: string): string | null {
  try {
    if (url.includes("youtu.be/")) {
      return url.split("youtu.be/")[1]?.split(/[?&]/)[0] ?? null;
    }
    return new URL(url).searchParams.get("v");
  } catch {
    return null;
  }
}

function VideoEmbed({ url, provider }: { url: string; provider?: string }) {
  const youtubeId = provider === "youtube" ? parseYoutubeId(url) : null;

  if (youtubeId) {
    return (
      <div className="blog-video-embed">
        <iframe
          title="video"
          src={`https://www.youtube.com/embed/${youtubeId}`}
          style={{ width: "100%", height: "100%", border: 0 }}
          allowFullScreen
        />
      </div>
    );
  }

  return (
    <p className="blog-video-fallback">
      <a href={url} target="_blank" rel="noreferrer">
        {url}
      </a>
    </p>
  );
}

function BlogTextBody({ layoutType, bodyHtml }: { layoutType: string; bodyHtml: string }) {
  if (!bodyHtml) {
    return null;
  }

  if (isTwoColumnLayout(layoutType)) {
    return (
      <div
        className="blog-body blog-body--two-columns"
        dangerouslySetInnerHTML={{ __html: bodyHtml }}
      />
    );
  }

  return <div className="blog-body" dangerouslySetInnerHTML={{ __html: bodyHtml }} />;
}

export function BlogPostView({
  language,
  title,
  bodyHtml,
  contentKind,
  layoutType,
  coverImageUrl,
  videoUrl,
  videoProvider,
  publishAt,
  variant = "public",
  breadcrumb,
  backLink,
  emptyMessage,
  hashtagsSuffix,
  previewLanguageLabel,
  untitledPreviewLabel = "(untitled)",
}: BlogPostViewProps) {
  const isPreview = variant === "preview";
  const isVideo = normalizeBlogContentKind(contentKind) === "video";
  const videoLink = isVideo ? bodyHtml || videoUrl || "" : undefined;
  const hasContent = Boolean(title.trim() || bodyHtml.trim() || videoLink);
  const wideLayout = isTwoColumnLayout(layoutType);
  const contentWidth = wideLayout ? "960px" : "720px";

  if (emptyMessage && !hasContent) {
    return (
      <div className={`blog-post-view blog-post-view--${variant}`}>
        {isPreview && previewLanguageLabel ? (
          <p className="blog-post-view__badge">Preview — {previewLanguageLabel}</p>
        ) : null}
        <p className="blog-post-view__empty">{emptyMessage}</p>
      </div>
    );
  }

  const hero = (
    <div
      className={
        coverImageUrl
          ? "blog-post-view__hero blog-post-view__hero--cover"
          : isPreview
            ? "blog-post-view__hero"
            : "page-hero blog-post-view__public-hero"
      }
      style={coverImageUrl ? { backgroundImage: `url(${coverImageUrl})` } : undefined}
    >
      {coverImageUrl ? <div className="blog-post-view__hero-overlay" aria-hidden="true" /> : null}
      <div className="blog-post-view__content" style={{ maxWidth: contentWidth }}>
        {breadcrumb}
        {publishAt ? <time>{new Date(publishAt).toLocaleDateString(language)}</time> : null}
        <h1 className="blog-post-view__title">
          {title || (isPreview ? untitledPreviewLabel : "")}
        </h1>
      </div>
    </div>
  );

  const body = (
    <section className={isPreview ? "blog-post-view__body" : "section"}>
      <div
        className={`blog-detail blog-post-view__content${isPreview ? " blog-post-view__inner" : ""}`}
        style={{ maxWidth: contentWidth }}
      >
        {videoLink ? (
          <VideoEmbed url={videoLink} provider={videoProvider} />
        ) : (
          <BlogTextBody layoutType={layoutType} bodyHtml={bodyHtml} />
        )}
        {hashtagsSuffix?.trim() ? (
          <BlogPostHashtags hashtags={hashtagsSuffix} />
        ) : null}
        {backLink}
      </div>
    </section>
  );

  return (
    <div className={`blog-post-view blog-post-view--${variant}`}>
      {isPreview && previewLanguageLabel ? (
        <p className="blog-post-view__badge">Preview — {previewLanguageLabel}</p>
      ) : null}
      {hero}
      {body}
    </div>
  );
}
