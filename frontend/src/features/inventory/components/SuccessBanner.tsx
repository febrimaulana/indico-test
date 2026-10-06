import { memo, useCallback, useEffect } from "react";
import { formatIsoClock } from "../../../utils/format";

const AUTO_DISMISS_MS = 8_000;

interface SuccessBannerProps {
  reservationId: string;
  confirmedAt: string;
  onDismiss: () => void;
}

/** Shown after a successful Confirm Purchase; auto-dismisses after 8 seconds. */
export const SuccessBanner = memo(function SuccessBanner({ reservationId, confirmedAt, onDismiss }: SuccessBannerProps) {
  useEffect(() => {
    const timerId = window.setTimeout(onDismiss, AUTO_DISMISS_MS);
    return () => window.clearTimeout(timerId);
  }, [onDismiss]);

  const handleDismiss = useCallback(() => onDismiss(), [onDismiss]);

  return (
    <div
      role="status"
      className="animate-pop-in relative overflow-hidden rounded-2xl border border-emerald-500/40 bg-linear-to-r from-emerald-950/70 to-teal-950/50"
    >
      <div className="flex flex-wrap items-center gap-3 px-4 py-3.5">
        <span className="flex h-9 w-9 items-center justify-center rounded-full bg-emerald-500/20">
          <svg viewBox="0 0 20 20" className="h-5 w-5" fill="none" aria-hidden="true">
            <path
              className="animate-check"
              d="M4 10.5l4 4L16 6"
              stroke="#6ee7b7"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
        </span>
        <div className="min-w-0">
          <p className="text-sm font-semibold text-emerald-100">Pembelian berhasil dikonfirmasi!</p>
          <p className="mt-0.5 text-xs text-emerald-300/80">
            <span className="font-mono">{reservationId}</span> · {formatIsoClock(confirmedAt)}
          </p>
        </div>
        <button
          type="button"
          className="ml-auto rounded-lg px-2.5 py-1 text-xs font-medium text-emerald-300/80 transition-colors hover:bg-emerald-500/10 hover:text-emerald-200"
          onClick={handleDismiss}
          aria-label="Tutup notifikasi"
        >
          Tutup
        </button>
      </div>
      <div className="animate-shrink h-0.5 bg-emerald-400/60" />
    </div>
  );
});
