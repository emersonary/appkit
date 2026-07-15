"use client";

import { useMemo } from "react";
import { ResourceEditScreen, type ResourceEditScreenProps } from "./ResourceEditScreen";
import type { ResourceViewEditHandlers } from "./handlers";
import {
  type ResourceEditEndpointPaths,
  type ResourceEndpointHttp,
} from "./resource-endpoints";

export type ResourceEditEndpointsProps = Omit<ResourceEditScreenProps, "editHandlers"> & {
  http: Pick<ResourceEndpointHttp, "getEdit" | "patchEdit">;
  endpoints: ResourceEditEndpointPaths;
  uploadImage?: (file: File) => Promise<string>;
};

/**
 * Edit-only resource screen driven by GET + PATCH endpoint URLs.
 * Host supplies `http` transport (auth / protobuf / JSON).
 */
export function ResourceEditEndpoints({
  http,
  endpoints,
  uploadImage,
  ...screenProps
}: ResourceEditEndpointsProps) {
  const editHandlers = useMemo<ResourceViewEditHandlers>(
    () => ({
      onLoad: async () => http.getEdit(endpoints.get),
      onSubmit: async (values) => http.patchEdit(endpoints.patch, values),
      onUploadImage: uploadImage,
    }),
    [endpoints.get, endpoints.patch, http, uploadImage],
  );

  return <ResourceEditScreen {...screenProps} editHandlers={editHandlers} />;
}
