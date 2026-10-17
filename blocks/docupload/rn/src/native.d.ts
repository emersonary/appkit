declare module "react-native" {
  import type { ComponentType, ReactNode } from "react";

  type Style = Record<string, unknown>;

  export const View: ComponentType<{ style?: unknown; children?: ReactNode; accessibilityRole?: string; accessibilityValue?: { min?: number; max?: number; now?: number } }>;
  export const Text: ComponentType<{ style?: unknown; children?: ReactNode }>;
  export const Pressable: ComponentType<{
    style?: unknown;
    disabled?: boolean;
    accessibilityState?: { busy?: boolean; disabled?: boolean };
    onPress?: () => void;
    children?: ReactNode;
  }>;
  export const ActivityIndicator: ComponentType<{ color?: string }>;
  export const StyleSheet: {
    create<T extends Record<string, Style>>(styles: T): T;
    absoluteFillObject: Style;
  };
}

declare module "react-native-image-picker" {
  export type Asset = {
    uri?: string;
    base64?: string;
    fileName?: string;
    type?: string;
    timestamp?: string;
    fileSize?: number;
  };

  export type PickerResponse = { assets?: Asset[]; didCancel?: boolean; errorMessage?: string };

  export function launchCamera(options: {
    mediaType: "photo";
    includeBase64?: boolean;
    saveToPhotos?: boolean;
  }): Promise<PickerResponse>;

  export function launchImageLibrary(options: {
    mediaType: "photo";
    includeBase64?: boolean;
    selectionLimit?: number;
  }): Promise<PickerResponse>;
}
