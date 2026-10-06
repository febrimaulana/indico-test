import { useEffect, useState } from "react";

export interface CountdownState {
  remainingSeconds: number;
  isExpired: boolean;
}

function secondsUntil(expiresAt: string | null): number {
  if (!expiresAt) {
    return 0;
  }
  const diffMillis = new Date(expiresAt).getTime() - Date.now();
  return diffMillis > 0 ? Math.ceil(diffMillis / 1000) : 0;
}

/**
 * Live countdown derived from the server's expires_at (UTC). The interval is
 * always cleared on unmount / expiry change so no timer leaks.
 */
export function useCountdown(expiresAt: string | null): CountdownState {
  const [remainingSeconds, setRemainingSeconds] = useState<number>(() => secondsUntil(expiresAt));

  useEffect(() => {
    if (!expiresAt) {
      setRemainingSeconds(0);
      return;
    }

    setRemainingSeconds(secondsUntil(expiresAt));
    const intervalId = window.setInterval(() => {
      setRemainingSeconds(secondsUntil(expiresAt));
    }, 1_000);

    return () => window.clearInterval(intervalId);
  }, [expiresAt]);

  return {
    remainingSeconds,
    isExpired: expiresAt !== null && remainingSeconds <= 0,
  };
}
