import { uploadDocument, type DocUploadApi } from "@emersonary/appkit-doc-upload-core";
import { useRef, useState } from "react";
import { ActivityIndicator, Pressable, StyleSheet, Text, View } from "react-native";
import { launchCamera, launchImageLibrary, type Asset } from "react-native-image-picker";

export type DocUploaderProps = {
  api: DocUploadApi;
  onComplete?: (fileId: string) => void;
  onError?: (error: Error) => void;
};

type Busy = "gallery" | "camera" | null;

export function DocUploader({ api, onComplete, onError }: DocUploaderProps) {
  const [busy, setBusy] = useState<Busy>(null);
  const [activeName, setActiveName] = useState<string | null>(null);
  const [percent, setPercent] = useState<number | null>(null);
  const [error, setError] = useState<string | null>(null);
  const running = useRef(false);

  function reportError(err: unknown) {
    const error = err instanceof Error ? err : new Error("Upload failed");
    setError(error.message);
    onError?.(error);
  }

  async function sendAsset(asset: Asset | undefined, source: "gallery" | "camera") {
    if (!asset) return;
    const bytes = await assetBytes(asset);
    const name = asset.fileName || `${source}-${Date.now()}.jpg`;
    setActiveName(name);
    setPercent(0);
    const { fileId } = await uploadDocument(
      {
        name,
        bytes,
        contentType: asset.type || "image/jpeg",
        source,
        fileUnixMs: asset.timestamp ? Date.parse(asset.timestamp) || Date.now() : Date.now(),
      },
      api,
      { onProgress: setPercent },
    );
    onComplete?.(fileId);
  }

  async function pick(source: "gallery" | "camera") {
    if (running.current) return;
    setError(null);
    setBusy(source);
    running.current = true;
    try {
      const result =
        source === "camera"
          ? await launchCamera({ mediaType: "photo", includeBase64: true, saveToPhotos: false })
          : await launchImageLibrary({ mediaType: "photo", includeBase64: true, selectionLimit: 1 });
      if (result.errorMessage) throw new Error(result.errorMessage);
      if (!result.didCancel) await sendAsset(result.assets?.[0], source);
    } catch (err) {
      reportError(err);
    } finally {
      running.current = false;
      setBusy(null);
      setPercent(null);
      setActiveName(null);
    }
  }

  return (
    <View style={styles.box}>
      <Text style={styles.copy}>Choose a photo from your gallery, or take one with the camera.</Text>
      <View style={styles.row}>
        <Pressable
          style={styles.button}
          disabled={busy !== null || percent !== null}
          accessibilityState={{ busy: busy === "gallery", disabled: busy !== null }}
          onPress={() => void pick("gallery")}
        >
          {busy === "gallery" ? <ActivityIndicator color="#ffffff" /> : null}
          <Text style={styles.buttonLabel}>Gallery</Text>
        </Pressable>
        <Pressable
          style={[styles.button, styles.secondary]}
          disabled={busy !== null || percent !== null}
          accessibilityState={{ busy: busy === "camera", disabled: busy !== null }}
          onPress={() => void pick("camera")}
        >
          {busy === "camera" ? <ActivityIndicator color="#1f6f8b" /> : null}
          <Text style={[styles.buttonLabel, styles.secondaryLabel]}>Camera</Text>
        </Pressable>
      </View>
      {activeName ? <Text style={styles.status}>Uploading {activeName}</Text> : null}
      {error ? <Text style={styles.error}>{error}</Text> : null}
      {percent !== null ? <Pie percent={percent} /> : null}
    </View>
  );
}

function Pie({ percent }: { percent: number }) {
  const clamped = Math.max(0, Math.min(100, percent));
  const turn = (clamped / 100) * 360;
  const right = clamped > 50;
  return (
    <View
      style={styles.pie}
      accessibilityRole="progressbar"
      accessibilityValue={{ min: 0, max: 100, now: clamped }}
    >
      <View style={styles.pieCircle}>
        <View style={[styles.hold, styles.holdRight]}>
          <View style={[styles.fill, { transform: [{ rotate: `${Math.min(turn, 180)}deg` }] }]} />
        </View>
        <View style={[styles.hold, styles.holdLeft]}>
          <View
            style={[
              styles.fill,
              { transform: [{ rotate: `${right ? turn - 180 : 0}deg` }] },
              !right ? styles.fillHidden : null,
            ]}
          />
        </View>
        <View style={styles.hole}>
          <Text style={styles.holeLabel}>{clamped}%</Text>
        </View>
      </View>
    </View>
  );
}

async function assetBytes(asset: Asset): Promise<Uint8Array> {
  if (asset.base64) return decodeBase64(asset.base64);
  if (!asset.uri) throw new Error("The selected photo has no data");
  const response = await fetch(asset.uri);
  return new Uint8Array(await response.arrayBuffer());
}

function decodeBase64(value: string): Uint8Array {
  const clean = value.replace(/\s/g, "");
  const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";
  const bytes: number[] = [];
  for (let i = 0; i < clean.length; i += 4) {
    const a = alphabet.indexOf(clean[i] ?? "A");
    const b = alphabet.indexOf(clean[i + 1] ?? "A");
    const c = clean[i + 2] === "=" ? -1 : alphabet.indexOf(clean[i + 2] ?? "");
    const d = clean[i + 3] === "=" ? -1 : alphabet.indexOf(clean[i + 3] ?? "");
    bytes.push((a << 2) | (b >> 4));
    if (c >= 0) bytes.push(((b & 15) << 4) | (c >> 2));
    if (d >= 0) bytes.push(((c & 3) << 6) | d);
  }
  return Uint8Array.from(bytes);
}

const styles = StyleSheet.create({
  box: {
    position: "relative",
    minHeight: 160,
    padding: 16,
    borderWidth: 1.5,
    borderStyle: "dashed",
    borderColor: "#8aa0b4",
    borderRadius: 12,
    backgroundColor: "#f7fafc",
  },
  copy: { color: "#1c2b36", marginBottom: 12 },
  row: { flexDirection: "row", gap: 8 },
  button: {
    flexDirection: "row",
    alignItems: "center",
    gap: 8,
    backgroundColor: "#1f6f8b",
    borderRadius: 8,
    paddingVertical: 10,
    paddingHorizontal: 14,
  },
  secondary: { backgroundColor: "#ffffff", borderWidth: 1, borderColor: "#1f6f8b" },
  buttonLabel: { color: "#ffffff", fontWeight: "600" },
  secondaryLabel: { color: "#1f6f8b" },
  status: { marginTop: 12, color: "#1c2b36" },
  error: { marginTop: 8, color: "#9b2335" },
  pie: { position: "absolute", right: 12, bottom: 12 },
  pieCircle: {
    width: 64,
    height: 64,
    borderRadius: 32,
    backgroundColor: "#d5e0e8",
    overflow: "hidden",
  },
  hold: { position: "absolute", width: 32, height: 64, overflow: "hidden" },
  holdRight: { right: 0 },
  holdLeft: { left: 0 },
  fill: {
    position: "absolute",
    width: 32,
    height: 64,
    backgroundColor: "#1f6f8b",
  },
  fillHidden: { backgroundColor: "transparent" },
  hole: {
    position: "absolute",
    top: 10,
    left: 10,
    width: 44,
    height: 44,
    borderRadius: 22,
    backgroundColor: "#f7fafc",
    alignItems: "center",
    justifyContent: "center",
  },
  holeLabel: { color: "#1c2b36", fontSize: 12, fontWeight: "700" },
});
