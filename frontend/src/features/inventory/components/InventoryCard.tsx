import { memo, useMemo } from "react";
import { Button } from "../../../components/ui/Button";
import { Spinner } from "../../../components/ui/Spinner";
import { formatClock, stockPercentage } from "../../../utils/format";
import type { StockResponse } from "../types";
import { ErrorBanner } from "./ErrorBanner";

interface InventoryCardProps {
  itemId: string;
  stock: StockResponse | undefined;
  isFetching: boolean;
  isPending: boolean;
  error: unknown;
  lastUpdatedAt: number | null;
  onRefresh: () => void;
}

interface StatProps {
  label: string;
  value: number;
  accentClass: string;
}

const Stat = memo(function Stat({ label, value, accentClass }: StatProps) {
  return (
    <div className="rounded-lg bg-slate-900/70 p-3">
      <p className="text-xs uppercase tracking-wide text-slate-400">{label}</p>
      <p className={`mt-1 text-2xl font-bold tabular-nums ${accentClass}`}>{value}</p>
    </div>
  );
});

export const InventoryCard = memo(function InventoryCard({
  itemId,
  stock,
  isFetching,
  isPending,
  error,
  lastUpdatedAt,
  onRefresh,
}: InventoryCardProps) {
  const reservedPercentage = useMemo(
    () => (stock ? stockPercentage(stock.reserved_stock, stock.total_stock) : 0),
    [stock],
  );
  const availablePercentage = useMemo(
    () => (stock ? stockPercentage(stock.available_stock, stock.total_stock) : 0),
    [stock],
  );

  return (
    <section className="rounded-2xl border border-slate-800 bg-slate-900/60 p-5 shadow-xl">
      <header className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 className="text-lg font-semibold text-white">Live Inventory Tracker</h2>
          <p className="text-xs text-slate-400">
            item: <span className="font-mono text-slate-300">{itemId}</span> · diperbarui {formatClock(lastUpdatedAt)}
          </p>
        </div>
        <Button variant="secondary" onClick={onRefresh} disabled={isFetching} aria-label="Refresh stok">
          {isFetching ? <Spinner label="Memuat..." /> : "↻ Refresh"}
        </Button>
      </header>

      {isPending ? (
        <p className="mt-4 text-sm text-slate-400">
          <Spinner label="Mengambil data stok..." />
        </p>
      ) : null}

      {stock ? (
        <>
          <div className="mt-4 grid grid-cols-3 gap-3">
            <Stat label="Total" value={stock.total_stock} accentClass="text-slate-100" />
            <Stat label="Reserved" value={stock.reserved_stock} accentClass="text-indigo-300" />
            <Stat label="Available" value={stock.available_stock} accentClass="text-emerald-300" />
          </div>

          <div className="mt-4 space-y-2">
            <div>
              <div className="flex justify-between text-[11px] text-slate-400">
                <span>reserved</span>
                <span>{reservedPercentage}%</span>
              </div>
              <div className="h-2 overflow-hidden rounded-full bg-slate-800">
                <div className="h-full rounded-full bg-indigo-500 transition-all" style={{ width: `${reservedPercentage}%` }} />
              </div>
            </div>
            <div>
              <div className="flex justify-between text-[11px] text-slate-400">
                <span>available</span>
                <span>{availablePercentage}%</span>
              </div>
              <div className="h-2 overflow-hidden rounded-full bg-slate-800">
                <div className="h-full rounded-full bg-emerald-500 transition-all" style={{ width: `${availablePercentage}%` }} />
              </div>
            </div>
          </div>
        </>
      ) : null}

      {error ? (
        <div className="mt-4">
          <ErrorBanner error={error} fallbackMessage="gagal memuat stok" />
        </div>
      ) : null}
    </section>
  );
});
