import SparkMD5 from "spark-md5";

export function md5Hex(data: Uint8Array): string {
  const view =
    data.byteOffset === 0 && data.byteLength === data.buffer.byteLength
      ? data
      : data.slice();
  return SparkMD5.ArrayBuffer.hash(view.buffer as ArrayBuffer);
}
