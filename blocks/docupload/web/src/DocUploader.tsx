import { uploadDocument, type DocUploadApi } from "@emersonary/appkit-doc-upload-core";
import { useEffect, useRef, useState, type DragEvent } from "react";
import "./doc-uploader.css";

export type DocUploaderProps = {
  api: DocUploadApi;
  onComplete?: (fileId: string) => void;
  onError?: (error: Error) => void;
  accept?: string;
};

type Busy = "file" | "camera" | "capture" | null;

type Queued = {
  name: string;
  bytes: Uint8Array;
  contentType: string;
  source: string;
  fileUnixMs: number;
};

export function DocUploader({ api, onComplete, onError, accept }: DocUploaderProps) {
  const inputRef = useRef<HTMLInputElement>(null);
  const videoRef = useRef<HTMLVideoElement>(null);
  const streamRef = useRef<MediaStream | null>(null);
  const queueRef = useRef<Queued[]>([]);
  const runningRef = useRef(false);
  const [dragOver, setDragOver] = useState(false);
  const [busy, setBusy] = useState<Busy>(null);
  const [cameraOn, setCameraOn] = useState(false);
  const [activeName, setActiveName] = useState<string | null>(null);
  const [percent, setPercent] = useState<number | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    return () => stopStream(streamRef.current);
  }, []);

  useEffect(() => {
    const video = videoRef.current;
    const stream = streamRef.current;
    if (!cameraOn || !video || !stream) return;
    video.srcObject = stream;
    void video.play().catch(() => undefined);
  }, [cameraOn]);

  function reportError(err: unknown) {
    const error = err instanceof Error ? err : new Error("Upload failed");
    setError(error.message);
    onError?.(error);
  }

  async function pump() {
    if (runningRef.current) return;
    runningRef.current = true;
    try {
      while (queueRef.current.length > 0) {
        const item = queueRef.current.shift();
        if (!item) break;
        setActiveName(item.name);
        setPercent(0);
        const { fileId } = await uploadDocument(item, api, { onProgress: setPercent });
        onComplete?.(fileId);
      }
    } catch (err) {
      reportError(err);
    } finally {
      runningRef.current = false;
      setPercent(null);
      setActiveName(null);
      setBusy(null);
    }
  }

  function enqueue(items: Queued[]) {
    if (items.length === 0) return;
    setError(null);
    queueRef.current.push(...items);
    void pump();
  }

  async function addFiles(files: File[], source: string) {
    const queued: Queued[] = [];
    for (const file of files) {
      queued.push({
        name: file.name || source,
        bytes: new Uint8Array(await file.arrayBuffer()),
        contentType: file.type || "application/octet-stream",
        source,
        fileUnixMs: file.lastModified || Date.now(),
      });
    }
    enqueue(queued);
  }

  async function onDrop(event: DragEvent<HTMLElement>) {
    event.preventDefault();
    setDragOver(false);
    setBusy("file");
    try {
      await addFiles(Array.from(event.dataTransfer.files), "drop");
    } catch (err) {
      reportError(err);
      setBusy(null);
    }
  }

  async function onPick(files: FileList | null) {
    setBusy("file");
    try {
      await addFiles(Array.from(files ?? []), "drop");
    } catch (err) {
      reportError(err);
    } finally {
      if (!runningRef.current) setBusy(null);
    }
  }

  async function openCamera() {
    setError(null);
    setBusy("camera");
    try {
      const stream = await navigator.mediaDevices.getUserMedia({
        video: { facingMode: "environment" },
        audio: false,
      });
      streamRef.current = stream;
      setCameraOn(true);
    } catch (err) {
      reportError(err);
    } finally {
      setBusy(null);
    }
  }

  function closeCamera() {
    stopStream(streamRef.current);
    streamRef.current = null;
    setCameraOn(false);
  }

  function capture() {
    const video = videoRef.current;
    if (!video || video.videoWidth === 0) return;
    setBusy("capture");
    const canvas = document.createElement("canvas");
    canvas.width = video.videoWidth;
    canvas.height = video.videoHeight;
    const context = canvas.getContext("2d");
    if (!context) {
      setBusy(null);
      return;
    }
    context.drawImage(video, 0, 0);
    canvas.toBlob(async (blob) => {
      try {
        if (!blob) throw new Error("Could not read the photo");
        const name = `camera-${Date.now()}.jpg`;
        enqueue([
          {
            name,
            bytes: new Uint8Array(await blob.arrayBuffer()),
            contentType: "image/jpeg",
            source: "camera",
            fileUnixMs: Date.now(),
          },
        ]);
        closeCamera();
      } catch (err) {
        reportError(err);
        setBusy(null);
      }
    }, "image/jpeg", 0.92);
  }

  const uploading = percent !== null;

  return (
    <section
      className={`appkit-doc-uploader${dragOver ? " appkit-doc-uploader--active" : ""}`}
      onDragOver={(event) => {
        event.preventDefault();
        setDragOver(true);
      }}
      onDragLeave={() => setDragOver(false)}
      onDrop={(event) => void onDrop(event)}
    >
      <p className="appkit-doc-uploader__copy">
        Drop a file here, or choose one. You can also take a photo and send that picture.
      </p>
      <div className="appkit-doc-uploader__actions">
        <input
          ref={inputRef}
          className="appkit-doc-uploader__file"
          type="file"
          accept={accept}
          onChange={(event) => {
            void onPick(event.target.files);
            event.target.value = "";
          }}
        />
        <button
          type="button"
          className="appkit-doc-uploader__button"
          disabled={busy !== null || uploading}
          aria-busy={busy === "file"}
          onClick={() => inputRef.current?.click()}
        >
          {busy === "file" ? <span className="appkit-doc-uploader__spinner" aria-hidden="true" /> : null}
          Choose a file
        </button>
        <button
          type="button"
          className="appkit-doc-uploader__button appkit-doc-uploader__button--secondary"
          disabled={busy !== null || uploading || cameraOn}
          aria-busy={busy === "camera"}
          onClick={() => void openCamera()}
        >
          {busy === "camera" ? <span className="appkit-doc-uploader__spinner" aria-hidden="true" /> : null}
          Take a photo
        </button>
      </div>
      {cameraOn ? (
        <>
          <video ref={videoRef} className="appkit-doc-uploader__video" playsInline muted />
          <div className="appkit-doc-uploader__actions">
            <button
              type="button"
              className="appkit-doc-uploader__button"
              disabled={busy === "capture" || uploading}
              aria-busy={busy === "capture"}
              onClick={capture}
            >
              {busy === "capture" ? <span className="appkit-doc-uploader__spinner" aria-hidden="true" /> : null}
              Use this photo
            </button>
            <button type="button" className="appkit-doc-uploader__button appkit-doc-uploader__button--secondary" onClick={closeCamera}>
              Cancel
            </button>
          </div>
        </>
      ) : null}
      {activeName ? <p className="appkit-doc-uploader__status">Uploading {activeName}</p> : null}
      {error ? <p className="appkit-doc-uploader__error">{error}</p> : null}
      {percent !== null ? <Pie percent={percent} /> : null}
    </section>
  );
}

function Pie({ percent }: { percent: number }) {
  const radius = 18;
  const circumference = 2 * Math.PI * radius;
  const dash = (Math.max(0, Math.min(100, percent)) / 100) * circumference;
  return (
    <div
      className="appkit-doc-uploader__pie"
      role="progressbar"
      aria-valuemin={0}
      aria-valuemax={100}
      aria-valuenow={percent}
      aria-label="Upload progress"
    >
      <svg viewBox="0 0 48 48" aria-hidden="true">
        <circle className="appkit-doc-uploader__pie-track" cx="24" cy="24" r={radius} />
        <circle
          className="appkit-doc-uploader__pie-value"
          cx="24"
          cy="24"
          r={radius}
          strokeDasharray={`${dash} ${circumference - dash}`}
        />
      </svg>
      <span className="appkit-doc-uploader__pie-label">{percent}%</span>
    </div>
  );
}

function stopStream(stream: MediaStream | null) {
  stream?.getTracks().forEach((track) => track.stop());
}
