import { useQuery } from "@tanstack/react-query";
import { STOCK_POLL_INTERVAL_MS } from "../../../config";
import { getStock } from "../api";
import { inventoryQueryKeys } from "../queryKeys";

/**
 * Live inventory tracker: polls the stock endpoint, pauses while the tab is
 * hidden (default TanStack behavior) and cancels in-flight requests on
 * unmount through the query signal.
 */
export function useStockQuery(itemId: string) {
  return useQuery({
    queryKey: inventoryQueryKeys.stock(itemId),
    queryFn: ({ signal }) => getStock(itemId, signal),
    refetchInterval: STOCK_POLL_INTERVAL_MS,
    enabled: itemId.length > 0,
  });
}
