/** Format seconds as mm:ss for the countdown. */
export function formatCountdown(totalSeconds: number): string {
  const safeSeconds = Math.max(0, Math.floor(totalSeconds));
  const minutes = Math.floor(safeSeconds / 60);
  const seconds = safeSeconds % 60;
  return `${String(minutes).padStart(2, "0")}:${String(seconds).padStart(2, "0")}`;
}

/** Render an epoch millis timestamp as a local HH:mm:ss clock. */
export function formatClock(epochMillis: number | null): string {
  if (epochMillis === null) {
    return "-";
  }
  return new Date(epochMillis).toLocaleTimeString("id-ID", {
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });
}

/** Percentage of `part` against `total`, clamped to 0..100. */
export function stockPercentage(part: number, total: number): number {
  if (total <= 0) {
    return 0;
  }
  return Math.min(100, Math.max(0, Math.round((part / total) * 100)));
}
