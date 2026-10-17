import { gzipBytes } from "./gzip";
import { md5Hex } from "./md5";

const defaultRetries = 5;
const maxBackoffMs = 8000;

export type DocFile = {
  name: string;
  bytes: Uint8Array;
  contentType?: string;
  source?: string;
  fileUnixMs?: number;
  expiresUnixMs?: number;
};

export type StartUploadInput = {
  fileId: string;
  fileName: string;
  fileUnixMs: number;
  originalSize: number;
  compressedMd5: string;
  compressedSize: number;
  expiresUnixMs: number;
  contentType: string;
  source: string;
};

export type StartResult = {
  blockSize: number;
  pendingBlocks: number[];
  maxThreads: number;
  expiresUnixMs: number;
};

export type BlockResult = {
  success: boolean;
  pendingBlocks: number[];
};

export type DocUploadApi = {
  startUpload(input: StartUploadInput): Promise<StartResult>;
  uploadBlock(input: {
    fileId: string;
    index: number;
    payload: Uint8Array;
    blockMd5: string;
  }): Promise<BlockResult>;
  cancelUpload(fileId: string): Promise<void>;
};

export type UploadOptions = {
  onProgress?: (percent: number) => void;
  signal?: AbortSignal;
  retries?: number;
};

export function splitBlocks(data: Uint8Array, size: number): Uint8Array[] {
  if (size <= 0 || data.length === 0) return [];
  const out: Uint8Array[] = [];
  for (let i = 0; i < data.length; i += size) {
    out.push(data.subarray(i, Math.min(i + size, data.length)));
  }
  return out;
}

export async function uploadDocument(
  file: DocFile,
  api: DocUploadApi,
  options: UploadOptions = {},
): Promise<{ fileId: string }> {
  throwIfAborted(options.signal);
  const compressed = gzipBytes(file.bytes);
  const fileId = md5Hex(file.bytes);
  const started = await api.startUpload({
    fileId,
    fileName: file.name,
    fileUnixMs: file.fileUnixMs ?? Date.now(),
    originalSize: file.bytes.byteLength,
    compressedMd5: md5Hex(compressed),
    compressedSize: compressed.byteLength,
    expiresUnixMs: file.expiresUnixMs ?? 0,
    contentType: file.contentType || "application/octet-stream",
    source: file.source || "drop",
  });
  const blocks = splitBlocks(compressed, started.blockSize);
  const total = blocks.length;
  let done = Math.max(0, total - started.pendingBlocks.length);
  const report = () => {
    if (!options.onProgress || total === 0) return;
    options.onProgress(Math.min(100, Math.round((done / total) * 100)));
  };
  report();
  await mapPool(started.pendingBlocks, Math.max(1, started.maxThreads), async (index) => {
    throwIfAborted(options.signal);
    const payload = blocks[index];
    if (!payload) throw new Error(`block ${index} is outside the compressed file`);
    await sendBlock(api, fileId, index, payload, options);
    done += 1;
    report();
  });
  options.onProgress?.(100);
  return { fileId };
}

async function sendBlock(
  api: DocUploadApi,
  fileId: string,
  index: number,
  payload: Uint8Array,
  options: UploadOptions,
) {
  const retries = options.retries ?? defaultRetries;
  const blockMd5 = md5Hex(payload);
  let last: unknown;
  for (let attempt = 0; attempt < retries; attempt++) {
    throwIfAborted(options.signal);
    if (attempt > 0) await pause(attempt - 1, options.signal);
    try {
      const result = await api.uploadBlock({ fileId, index, payload, blockMd5 });
      if (result.success) return;
      last = new Error(`block ${index} was rejected`);
    } catch (err) {
      last = err;
    }
  }
  throw last instanceof Error ? last : new Error(`block ${index} failed`);
}

async function mapPool(indexes: number[], limit: number, worker: (index: number) => Promise<void>) {
  if (indexes.length === 0) return;
  let cursor = 0;
  let failure: unknown;
  const runners = Array.from({ length: Math.min(limit, indexes.length) }, async () => {
    while (failure === undefined) {
      const current = cursor;
      cursor += 1;
      if (current >= indexes.length) return;
      try {
        await worker(indexes[current]!);
      } catch (err) {
        failure ??= err;
        return;
      }
    }
  });
  await Promise.all(runners);
  if (failure !== undefined) throw failure;
}

function pause(attempt: number, signal?: AbortSignal) {
  const delay = Math.min(maxBackoffMs, 200 * 2 ** attempt);
  return new Promise<void>((resolve, reject) => {
    const timer = setTimeout(resolve, delay);
    signal?.addEventListener(
      "abort",
      () => {
        clearTimeout(timer);
        reject(abortError(signal));
      },
      { once: true },
    );
  });
}

function throwIfAborted(signal?: AbortSignal) {
  if (signal?.aborted) throw abortError(signal);
}

function abortError(signal: AbortSignal) {
  return signal.reason instanceof Error ? signal.reason : new Error("upload aborted");
}
