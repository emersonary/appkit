import { createClient, type Interceptor } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import type { DocUploadApi, StartUploadInput } from "@emersonary/appkit-doc-upload-core";
import { DocUploadService } from "./gen/v1/docupload_connect";

export type CreateDocUploadApiOptions = {
  baseUrl?: string;
  readAccessToken?: () => string | null;
};

export function createDocUploadApi(options: CreateDocUploadApiOptions = {}): DocUploadApi {
  const auth: Interceptor = (next) => async (req) => {
    const token = options.readAccessToken?.();
    if (token) req.header.set("Authorization", `Bearer ${token}`);
    return next(req);
  };
  const client = createClient(
    DocUploadService,
    createConnectTransport({
      baseUrl: options.baseUrl ?? "",
      interceptors: [auth],
    }),
  );

  return {
    async startUpload(input: StartUploadInput) {
      const response = await client.startUpload({
        fileId: input.fileId,
        fileName: input.fileName,
        fileUnixMs: BigInt(input.fileUnixMs),
        originalSize: BigInt(input.originalSize),
        compressedMd5: input.compressedMd5,
        compressedSize: BigInt(input.compressedSize),
        expiresUnixMs: BigInt(input.expiresUnixMs),
        contentType: input.contentType,
        source: input.source,
      });
      return {
        blockSize: response.blockSize,
        pendingBlocks: response.pendingBlocks,
        maxThreads: response.maxThreads,
        expiresUnixMs: Number(response.expiresUnixMs),
      };
    },
    async uploadBlock(input) {
      const response = await client.uploadBlock({
        fileId: input.fileId,
        index: input.index,
        payload: input.payload,
        blockMd5: input.blockMd5,
      });
      return { success: response.success, pendingBlocks: response.pendingBlocks };
    },
    async cancelUpload(fileId) {
      await client.cancelUpload({ fileId });
    },
  };
}
