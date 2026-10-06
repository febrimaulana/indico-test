import { memo, useMemo } from "react";
import { formatCountdown } from "../../../utils/format";

const RING_RADIUS = 62;
const RING_CIRCUMFERENCE = 2 * Math.PI * RING_RADIUS;
const TICK_COUNT = 12;

interface CountdownRingProps {
  remainingSeconds: number;
  totalSeconds: number;
  isExpired: boolean;
}

function buildTicks(): { x1: number; y1: number; x2: number; y2: number }[] {
  const ticks: { x1: number; y1: number; x2: number; y2: number }[] = [];
  for (let index = 0; index < TICK_COUNT; index += 1) {
    const angle = (index / TICK_COUNT) * Math.PI * 2;
    const cos = Math.cos(angle);
    const sin = Math.sin(angle);
    ticks.push({
      x1: 80 + cos * 72,
      y1: 80 + sin * 72,
      x2: 80 + cos * 76,
      y2: 80 + sin * 76,
    });
  }
  return ticks;
}

const TICKS = buildTicks();

/** Circular 5-minute countdown: green → amber (≤60s, with glow) → rose (expired). */
export const CountdownRing = memo(function CountdownRing({
  remainingSeconds,
  totalSeconds,
  isExpired,
}: CountdownRingProps) {
  const progress = useMemo(() => {
    if (totalSeconds <= 0) {
      return 0;
    }
    return Math.min(1, Math.max(0, remainingSeconds / totalSeconds));
  }, [remainingSeconds, totalSeconds]);

  const ringOffset = useMemo(() => RING_CIRCUMFERENCE * (1 - progress), [progress]);

  const isWarning = !isExpired && remainingSeconds <= 60;
  const strokeColor = isExpired ? "#fb7185" : isWarning ? "#fbbf24" : "#34d399";

  return (
    <div className="relative h-44 w-44" role="timer" aria-label="Sisa waktu reservasi">
      <svg viewBox="0 0 160 160" className="h-full w-full -rotate-90">
        <g className="stroke-slate-700" strokeWidth="2" strokeLinecap="round">
          {TICKS.map((tick, index) => (
            <line key={index} x1={tick.x1} y1={tick.y1} x2={tick.x2} y2={tick.y2} />
          ))}
        </g>
        <circle cx="80" cy="80" r={RING_RADIUS} className="stroke-slate-800" strokeWidth="10" fill="none" />
        <circle
          cx="80"
          cy="80"
          r={RING_RADIUS}
          stroke={strokeColor}
          strokeWidth="10"
          fill="none"
          strokeLinecap="round"
          strokeDasharray={RING_CIRCUMFERENCE}
          strokeDashoffset={ringOffset}
          className={`transition-[stroke-dashoffset] duration-1000 ease-linear ${isWarning ? "animate-glow-pulse" : ""}`}
        />
      </svg>
      <div className="absolute inset-0 flex flex-col items-center justify-center">
        <span
          className={`font-mono text-4xl font-bold tabular-nums transition-colors ${
            isExpired ? "text-rose-400" : isWarning ? "text-amber-300" : "text-white"
          }`}
          data-testid="countdown"
        >
          {formatCountdown(remainingSeconds)}
        </span>
        <span className="mt-1 text-[10px] uppercase tracking-[0.2em] text-slate-400">
          {isExpired ? "kedaluwarsa" : isWarning ? "segera habis" : "sisa waktu"}
        </span>
      </div>
    </div>
  );
});
