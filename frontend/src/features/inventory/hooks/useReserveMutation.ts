import { useMutation, useQueryClient } from "@tanstack/react-query";
import { reserveStock } from "../api";
import { inventoryQueryKeys } from "../queryKeys";

/** Reserves stock and refreshes the stock query on success. */
export function useReserveMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: reserveStock,
    onSuccess: (_response, variables) => {
      void queryClient.invalidateQueries({ queryKey: inventoryQueryKeys.stock(variables.item_id) });
    },
  });
}
