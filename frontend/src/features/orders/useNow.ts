import { useCallback, useSyncExternalStore } from "react";

let currentTime = Date.now();
let timer: ReturnType<typeof setInterval> | null = null;
const listeners = new Set<() => void>();

function tick() {
  currentTime = Date.now();
  for (const listener of listeners) listener();
}

function subscribe(listener: () => void) {
  listeners.add(listener);

  // One interval shared by every countdown cell, regardless of row count.
  if (listeners.size === 1) {
    currentTime = Date.now();
    timer = setInterval(tick, 1_000);
  }

  return () => {
    listeners.delete(listener);
    if (listeners.size === 0 && timer !== null) {
      clearInterval(timer);
      timer = null;
    }
  };
}

function getSnapshot() {
  return currentTime;
}

export function useNow(): number {
  return useSyncExternalStore(subscribe, getSnapshot, getSnapshot);
}

/**
 * Subscribes to the shared clock but only re-renders when the boolean result
 * changes, so rows do not re-render on every tick.
 */
export function useIsExpired(expiresAt: string | null | undefined): boolean {
  const expiresAtMs = expiresAt ? Date.parse(expiresAt) : Number.NaN;
  const getExpiredSnapshot = useCallback(
    () => (Number.isFinite(expiresAtMs) ? currentTime >= expiresAtMs : false),
    [expiresAtMs],
  );

  return useSyncExternalStore(
    subscribe,
    getExpiredSnapshot,
    getExpiredSnapshot,
  );
}

export function formatDuration(remainingMs: number): string {
  const totalSeconds = Math.max(0, Math.ceil(remainingMs / 1_000));
  const hours = Math.floor(totalSeconds / 3_600);
  const minutes = Math.floor((totalSeconds % 3_600) / 60);
  const seconds = totalSeconds % 60;

  if (hours > 0) {
    return `${hours}h ${String(minutes).padStart(2, "0")}m`;
  }
  if (minutes > 0) {
    return `${minutes}m ${String(seconds).padStart(2, "0")}s`;
  }
  return `${seconds}s`;
}
