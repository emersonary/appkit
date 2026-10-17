import { gzipSync } from "fflate";

/** gzip with a zero mtime so the same bytes always compress the same way. */
export function gzipBytes(data: Uint8Array): Uint8Array {
  return gzipSync(data, { mtime: 0 });
}
