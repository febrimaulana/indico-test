import { memo, useMemo } from "react";
import { Badge } from "../../../components/ui/Badge";
import type { BadgeTone } from "../../../components/ui/Badge";
import { Button } from "../../../components/ui/Button";
import { Spinner } from "../../../components/ui/Spinner";
import { useAnimatedNumber } from "../../../hooks/useAnimatedNumber";
import { formatClock, stockPercentage } from "../../../utils/format";
import { useStockHistory } from "../hooks/useStockHistory";
import type { StockResponse } from "../types";
import { ErrorBanner } from "./ErrorBanner";
import { Sparkline } from "./Sparkline";

const RING_RADIUS = 54;
const RING_CIRCUMFERENCE = 2 * Math.PI * RING_RADIUS;

interface StockPanelProps {
  itemId: string;
  stock: StockResponse | undefined;
  isFetching: boolean;
  isPending: boolean;
  error: unknown;
  lastUpdatedAt: number | null;
  onRefresh: () => void;
}

interface StatTileProps {
  label: string;
  value: number;
  percentage: number;
  barClass: string;
  hint: string;
  animatedValue: number;
}

const StatTile = memo(function StatTile({ label, value, percentage, barClass, hint, animatedValue }: StatTileProps) {
  return (
    <div className="rounded-xl border border-slate-800/80 bg-slate-950/60 p-3 transition-colors hover:border-slate-700">
      <div className="flex items-center justify-between text-[11px]">
        <span className="uppercase tracking-wider text-slate-400">{label}</span>
        <span className="font-mono text-slate-500">{percentage}%</span>
      </div>
      <p className="mt-1 text-2xl font-bold tabular-nums text-white" aria-label={`${label}: ${value}`}>
        {animatedValue}
      </p>
      <div className="mt-2 h-1.5 overflow-hidden rounded-full bg-slate-800">
        <div
          className={`h-full rounded-full transition-all duration-700 ease-out ${barClass}`}
          style={{ width: `${percentage}%` }}
        />
      </div>
      <p className="mt-1.5 text-[11px] text-slate-500">{hint}</p>
    </div>
  );
});

function SkeletonBlock({ className }: { className: string }) {
  return <div className={`skeleton rounded-xl ${className}`} />;
}

export const StockPanel = memo(function StockPanel({
  itemId,
  stock,
  isFetching,
  isPending,
  error,
  lastUpdatedAt,
  onRefresh,
}: StockPanelProps) {
  const history = useStockHistory(stock, lastUpdatedAt ?? 0);
  const animatedAvailable = useAnimatedNumber(stock?.available_stock ?? 0);
  const animatedTotal = useAnimatedNumber(stock?.total_stock ?? 0);
  const animatedReserved = useAnimatedNumber(stock?.reserved_stock ?? 0);

  const availablePercentage = useMemo(
    () => (stock ? stockPercentage(stock.available_stock, stock.total_stock) : 0),
    [stock],
  );
  const reservedPercentage = useMemo(
    () => (stock ? stockPercentage(stock.reserved_stock, stock.total_stock) : 0),
    [stock],
  );
  const ringOffset = useMemo(() => RING_CIRCUMFERENCE * (1 - availablePercentage / 100), [availablePercentage]);
  const status = useMemo<{ tone: BadgeTone; label: string }>(() => {
    if (!stock) {
      return { tone: "neutral", label: "memuat data" };
    }
    if (stock.available_stock <= 0) {
      return { tone: "danger", label: "stok habis" };
    }
    if (availablePercentage <= 20) {
      return { tone: "warning", label: "stok terbatas" };
    }
    return { tone: "success", label: "stok tersedia" };
  }, [stock, availablePercentage]);

  const trendDelta = useMemo(() => {
    if (history.length < 2) {
      return 0;
    }
    return history[history.length - 1]!.available - history[0]!.available;
  }, [history]);

  return (
    <section className="rounded-2xl bg-linear-to-b from-white/10 via-white/5 to-transparent p-px shadow-2xl shadow-slate-950/60">
      <div className="rounded-[15px] bg-slate-900/70 p-5 backdrop-blur">
        <header className="flex flex-wrap items-start justify-between gap-3">
          <div>
            <div className="flex items-center gap-2">
              <h2 className="text-lg font-semibold text-white">Live Inventory</h2>
              <Badge tone={status.tone}>{status.label}</Badge>
            </div>
            <p className="mt-1 text-xs text-slate-400">
              item <span className="rounded bg-slate-800/80 px-1.5 py-0.5 font-mono text-slate-300">{itemId}</span> ·
              diperbarui {formatClock(lastUpdatedAt)}
            </p>
          </div>
          <Button variant="secondary" onClick={onRefresh} disabled={isFetching} className="btn-shine" aria-label="Refresh stok">
            {isFetching ? (
              <Spinner label="Memuat" />
            ) : (
              <>
                <svg viewBox="0 0 20 20" className="h-4 w-4" fill="none" aria-hidden="true">
                  <path
                    d="M16 10a6 6 0 1 1-1.8-4.3M16 3v3h-3"
                    stroke="currentColor"
                    strokeWidth="1.8"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                  />
                </svg>
                Refresh
              </>
            )}
          </Button>
        </header>

        {isPending && !stock ? (
          <div className="mt-5 flex items-center gap-6">
            <SkeletonBlock className="h-40 w-40 rounded-full!" />
            <div className="grid flex-1 gap-3">
              <SkeletonBlock className="h-24" />
              <SkeletonBlock className="h-24" />
            </div>
          </div>
        ) : null}

        {stock ? (
          <>
            <div className="mt-5 flex flex-col items-center gap-6 sm:flex-row sm:items-stretch">
              <div className="relative h-40 w-40 shrink-0 self-center">
                <svg viewBox="0 0 140 140" className="h-full w-full -rotate-90">
                  <circle cx="70" cy="70" r={RING_RADIUS} className="stroke-slate-800" strokeWidth="12" fill="none" />
                  <circle
                    cx="70"
                    cy="70"
                    r={RING_RADIUS}
                    stroke="url(#stockRingGradient)"
                    strokeWidth="12"
                    fill="none"
                    strokeLinecap="round"
                    strokeDasharray={RING_CIRCUMFERENCE}
                    strokeDashoffset={ringOffset}
                    className="drop-shadow-[0_0_10px_rgba(99,102,241,0.45)] transition-[stroke-dashoffset] duration-700 ease-out"
                  />
                  <defs>
                    <linearGradient id="stockRingGradient" x1="0" y1="0" x2="1" y2="1">
                      <stop offset="0%" stopColor="#34d399" />
                      <stop offset="100%" stopColor="#6366f1" />
                    </linearGradient>
                  </defs>
                </svg>
                <div className="absolute inset-0 flex flex-col items-center justify-center">
                  <span className="text-4xl font-bold tabular-nums text-white" aria-label={`Tersedia: ${stock.available_stock}`}>
                    {animatedAvailable}
                  </span>
                  <span className="text-[10px] uppercase tracking-[0.2em] text-slate-400">tersedia</span>
                </div>
              </div>

              <div className="grid w-full flex-1 gap-3">
                <StatTile
                  label="Total stok"
                  value={stock.total_stock}
                  animatedValue={animatedTotal}
                  percentage={100}
                  barClass="bg-slate-500"
                  hint="kapasitas awal flash sale"
                />
                <StatTile
                  label="Direservasi"
                  value={stock.reserved_stock}
                  animatedValue={animatedReserved}
                  percentage={reservedPercentage}
                  barClass="bg-linear-to-r from-indigo-500 to-violet-500"
                  hint={`${availablePercentage}% masih tersedia untuk pembeli lain`}
                />
              </div>
            </div>

            <div className="mt-5 rounded-xl border border-slate-800/80 bg-slate-950/60 p-3">
              <div className="flex items-center justify-between text-[11px]">
                <span className="uppercase tracking-wider text-slate-400">Tren stok tersedia</span>
                <span
                  className={`font-mono ${
                    trendDelta < 0 ? "text-rose-300" : trendDelta > 0 ? "text-emerald-300" : "text-slate-500"
                  }`}
                >
                  {trendDelta > 0 ? `▲ +${trendDelta}` : trendDelta < 0 ? `▼ ${trendDelta}` : "—"} · {history.length} sampel
                </span>
              </div>
              <div className="mt-2">
                <Sparkline samples={history} />
              </div>
            </div>
          </>
        ) : null}

        {error ? (
          <div className="mt-4">
            <ErrorBanner error={error} fallbackMessage="gagal memuat stok" />
          </div>
        ) : null}
      </div>
    </section>
  );
});
