import { useEffect, useMemo, useState } from "react";
import { createBlogClient, getPublishedPost, type PublishedPost } from "../client";

export type UsePublishedPostOptions = {
  baseUrl: string;
  projectId: string;
  slug: string;
  language: string;
  getAccessToken?: () => string | null | Promise<string | null>;
};

export type UsePublishedPostResult = {
  post: PublishedPost | null;
  loading: boolean;
  error: Error | null;
  refresh: () => Promise<void>;
};

export function usePublishedPost({
  baseUrl,
  projectId,
  slug,
  language,
  getAccessToken,
}: UsePublishedPostOptions): UsePublishedPostResult {
  const client = useMemo(
    () => createBlogClient({ baseUrl, getAccessToken }),
    [baseUrl, getAccessToken],
  );

  const [post, setPost] = useState<PublishedPost | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  const refresh = async () => {
    if (!projectId || !slug || !language) {
      setPost(null);
      setLoading(false);
      return;
    }

    setLoading(true);
    setError(null);
    try {
      const response = await getPublishedPost(client, projectId, slug, language);
      setPost(response);
    } catch (err) {
      setError(err instanceof Error ? err : new Error("failed to load published post"));
      setPost(null);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void refresh();
  }, [client, projectId, slug, language]);

  return { post, loading, error, refresh };
}
