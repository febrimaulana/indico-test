import { memo, useMemo } from "react";
import type { StockSample } from "../hooks/useStockHistory";

const WIDTH = 220;
const HEIGHT = 48;
const PADDING = 4;

interface SparklineProps {
  samples: StockSample[];
}

interface Geometry {
  line: string;
  area: string;
  lastX: number;
  lastY: number;
}

function buildGeometry(samples: StockSample[]): Geometry {
  const values = samples.map((sample) => sample.available);
  let min = Math.min(...values);
  let max = Math.max(...values);
  if (min === max) {
    min -= 1;
    max += 1;
  }

  const stepX = (WIDTH - PADDING * 2) / Math.max(1, samples.length - 1);
  const points = samples.map((sample, index) => {
    const x = PADDING + index * stepX;
    const y = HEIGHT - PADDING - ((sample.available - min) / (max - min)) * (HEIGHT - PADDING * 2);
    return { x, y };
  });

  const line = points.map((point, index) => `${index === 0 ? "M" : "L"} ${point.x.toFixed(1)} ${point.y.toFixed(1)}`).join(" ");
  const first = points[0]!;
  const last = points[points.length - 1]!;
  const area = `${line} L ${last.x.toFixed(1)} ${HEIGHT - PADDING} L ${first.x.toFixed(1)} ${HEIGHT - PADDING} Z`;

  return { line, area, lastX: last.x, lastY: last.y };
}

/** Client-side trend of the available stock across polling samples. */
export const Sparkline = memo(function Sparkline({ samples }: SparklineProps) {
  const geometry = useMemo(() => (samples.length >= 2 ? buildGeometry(samples) : null), [samples]);

  if (!geometry) {
    return (
      <div className="flex h-12 items-center justify-center rounded-lg border border-dashed border-slate-800 text-[11px] text-slate-500">
        mengumpulkan sampel polling…
      </div>
    );
  }

  return (
    <svg viewBox={`0 0 ${WIDTH} ${HEIGHT}`} className="h-12 w-full" preserveAspectRatio="none" aria-hidden="true">
      <defs>
        <linearGradient id="sparkFill" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor="#6366f1" stopOpacity="0.35" />
          <stop offset="100%" stopColor="#6366f1" stopOpacity="0" />
        </linearGradient>
      </defs>
      <path d={geometry.area} fill="url(#sparkFill)" />
      <path
        d={geometry.line}
        fill="none"
        stroke="#818cf8"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <circle cx={geometry.lastX} cy={geometry.lastY} r="6" fill="#818cf8" opacity="0.25">
        <animate attributeName="r" values="4;9;4" dur="2.4s" repeatCount="indefinite" />
        <animate attributeName="opacity" values="0.35;0.05;0.35" dur="2.4s" repeatCount="indefinite" />
      </circle>
      <circle cx={geometry.lastX} cy={geometry.lastY} r="2.6" fill="#c7d2fe" />
    </svg>
  );
});
