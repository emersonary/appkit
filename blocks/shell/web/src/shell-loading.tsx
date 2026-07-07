"use client";

import {
  createContext,
  useCallback,
  useContext,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from "react";
import { usePathname, useSearchParams } from "next/navigation";

type ShellLoadingContextValue = {
  contentLoading: boolean;
  routeGeneration: number;
  startNavigation: () => void;
  setContentLoading: (loading: boolean) => void;
};

const ShellLoadingContext = createContext<ShellLoadingContextValue | null>(null);

export function ShellLoadingProvider({ children }: { children: ReactNode }) {
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const routeKey = useMemo(() => {
    const search = searchParams.toString();
    return search ? `${pathname}?${search}` : pathname;
  }, [pathname, searchParams]);
  const [contentLoading, setContentLoadingState] = useState(false);
  const [routeGeneration, setRouteGeneration] = useState(0);
  const routeGenerationRef = useRef(0);

  useLayoutEffect(() => {
    routeGenerationRef.current += 1;
    setRouteGeneration(routeGenerationRef.current);
    setContentLoadingState(true);
  }, [routeKey]);

  const setContentLoading = useCallback((loading: boolean) => {
    if (loading) {
      setContentLoadingState(true);
      return;
    }

    const generationAtRequest = routeGenerationRef.current;
    setContentLoadingState((current) => {
      if (generationAtRequest !== routeGenerationRef.current) {
        return current;
      }
      return false;
    });
  }, []);

  const startNavigation = useCallback(() => {
    setContentLoadingState(true);
  }, []);

  const value = useMemo(
    () => ({
      contentLoading,
      routeGeneration,
      startNavigation,
      setContentLoading,
    }),
    [contentLoading, routeGeneration, startNavigation, setContentLoading],
  );

  return <ShellLoadingContext.Provider value={value}>{children}</ShellLoadingContext.Provider>;
}

export function useShellLoading() {
  const context = useContext(ShellLoadingContext);
  if (!context) {
    throw new Error("useShellLoading must be used within ShellLoadingProvider");
  }
  return context;
}

/** Mirror a component's loading flag into the shell main panel. */
export function useSyncShellLoading(loading: boolean) {
  const { setContentLoading, routeGeneration } = useShellLoading();
  const generationRef = useRef(routeGeneration);

  useLayoutEffect(() => {
    generationRef.current = routeGeneration;
  }, [routeGeneration]);

  useLayoutEffect(() => {
    if (generationRef.current !== routeGeneration) {
      return;
    }
    setContentLoading(loading);
  }, [loading, routeGeneration, setContentLoading]);
}

/** Static pages with no async work should call this once mounted. */
export function useShellReady() {
  const { setContentLoading, routeGeneration } = useShellLoading();

  useLayoutEffect(() => {
    setContentLoading(false);
  }, [routeGeneration, setContentLoading]);
}

export type ShellContentLoadingProps = {
  label?: string;
  className?: string;
  labelClassName?: string;
};

export function ShellContentLoading({
  label = "Loading…",
  className = "appkit-shell-content-loading",
  labelClassName = "appkit-shell-content-loading__label",
}: ShellContentLoadingProps) {
  return (
    <div className={className} aria-live="polite" aria-busy="true">
      <p className={labelClassName}>{label}</p>
    </div>
  );
}
