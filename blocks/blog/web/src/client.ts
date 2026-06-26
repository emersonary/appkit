import { createClient, type Interceptor } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { BlogService } from "./gen/v1/blog_connect";
import { PageDirection, type PostSummary, type PublishedPostResponse } from "./gen/v1/blog_pb";

export type BlogClient = ReturnType<typeof createBlogClient>;

export type CreateBlogClientOptions = {
  baseUrl: string;
  getAccessToken?: () => string | null | Promise<string | null>;
};

export function createBlogClient({ baseUrl, getAccessToken }: CreateBlogClientOptions) {
  const interceptors: Interceptor[] = [];
  if (getAccessToken) {
    interceptors.push((next) => async (req) => {
      const token = await getAccessToken();
      if (token) {
        req.header.set("Authorization", `Bearer ${token}`);
      }
      return next(req);
    });
  }

  const transport = createConnectTransport({
    baseUrl,
    interceptors,
  });

  return createClient(BlogService, transport);
}

export type PublishedPost = PublishedPostResponse;

export type PublishedPostSummary = PostSummary;

export type ListPublishedPostsOptions = {
  pageSize?: number;
  page?: number;
  pageToken?: string;
  pageDirection?: "next" | "prev";
};

export type ListPublishedPostsResult = {
  posts: PostSummary[];
  nextPageToken?: string;
  prevPageToken?: string;
  totalCount: number;
  page: number;
};

export async function listPublishedPosts(
  client: BlogClient,
  projectId: string,
  language: string,
  options?: ListPublishedPostsOptions,
): Promise<ListPublishedPostsResult> {
  const response = await client.listPublishedPosts({
    projectId,
    language,
    pageSize: options?.pageSize ?? 0,
    page: options?.page ?? 0,
    pageToken: options?.pageToken ?? "",
    pageDirection:
      options?.pageDirection === "prev"
        ? PageDirection.PREV
        : PageDirection.UNSPECIFIED,
  });

  return {
    posts: response.posts ?? [],
    nextPageToken: response.nextPageToken || undefined,
    prevPageToken: response.prevPageToken || undefined,
    totalCount: response.totalCount ?? 0,
    page: response.page > 0 ? response.page : (options?.page ?? 1),
  };
}

export async function getPublishedPost(
  client: BlogClient,
  projectId: string,
  slug: string,
  language: string,
) {
  return client.getPublishedPost({ projectId, slug, language });
}
