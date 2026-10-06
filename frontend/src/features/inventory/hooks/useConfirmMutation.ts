import { useMutation, useQueryClient } from "@tanstack/react-query";
import { confirmReservation } from "../api";
import { inventoryQueryKeys } from "../queryKeys";

/**
 * Confirms a reservation. The dashboard displays a single item, so the stock
 * query for that item is invalidated on success.
 */
export function useConfirmMutation(itemId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: confirmReservation,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: inventoryQueryKeys.stock(itemId) });
    },
  });
}
