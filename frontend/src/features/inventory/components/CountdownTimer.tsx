import { memo } from "react";
import { formatCountdown } from "../../../utils/format";

interface CountdownTimerProps {
  remainingSeconds: number;
  isExpired: boolean;
}

export const CountdownTimer = memo(function CountdownTimer({ remainingSeconds, isExpired }: CountdownTimerProps) {
  const colorClass = isExpired
    ? "text-rose-400"
    : remainingSeconds <= 60
      ? "text-amber-300"
      : "text-emerald-300";

  return (
    <div className="flex items-baseline gap-2">
      <span className={`font-mono text-3xl font-bold tabular-nums ${colorClass}`} data-testid="countdown">
        {formatCountdown(remainingSeconds)}
      </span>
      <span className="text-xs text-slate-400">{isExpired ? "kedaluwarsa" : "sisa waktu"}</span>
    </div>
  );
});
