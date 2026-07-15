"use client";

import type { ReactNode } from "react";
import { ShellContentLoading, type ShellContentLoadingProps } from "./shell-loading";
import { useShellLoading } from "./shell-loading";

export type ShellContentPanelProps = {
  children: ReactNode;
  className?: string;
  loading?: ShellContentLoadingProps;
};

export function ShellContentPanel({
  children,
  className = "appkit-shell-content-panel",
  loading,
}: ShellContentPanelProps) {
  const { contentLoading } = useShellLoading();

  return (
    <div className={className}>
      {contentLoading ? <ShellContentLoading {...loading} /> : null}
      <div
        className={
          contentLoading
            ? "appkit-shell-content-panel__body appkit-shell-content-panel__body--loading"
            : "appkit-shell-content-panel__body"
        }
        aria-hidden={contentLoading}
      >
        {children}
      </div>
    </div>
  );
}
