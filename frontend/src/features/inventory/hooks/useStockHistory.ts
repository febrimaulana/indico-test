import { useEffect, useState } from "react";

export interface StockSample {
  timestamp: number;
  available: number;
  reserved: number;
}

const MAX_SAMPLES = 40;

/**
 * Keeps a small client-side history of poll results so the dashboard can draw
 * a trend sparkline. Samples are de-duplicated by the query's updatedAt
 * timestamp, and the buffer is capped to avoid unbounded growth.
 */
export function useStockHistory(
  stock: { available_stock: number; reserved_stock: number } | undefined,
  updatedAt: number,
): StockSample[] {
  const [history, setHistory] = useState<StockSample[]>([]);

  useEffect(() => {
    if (!stock || updatedAt <= 0) {
      return;
    }
    setHistory((previous) => {
      const last = previous[previous.length - 1];
      if (last && last.timestamp === updatedAt) {
        return previous;
      }
      const next = [
        ...previous,
        { timestamp: updatedAt, available: stock.available_stock, reserved: stock.reserved_stock },
      ];
      return next.length > MAX_SAMPLES ? next.slice(next.length - MAX_SAMPLES) : next;
    });
  }, [stock, updatedAt]);

  return history;
}
