import { useMutation, useQueryClient } from "@tanstack/react-query";
import { confirmReservation } from "../api";
import { inventoryQueryKeys } from "../queryKeys";
import type { ActiveReservationModel } from "../types";

/** Confirms a reservation and refreshes the stock query on success. */
export function useConfirmMutation(reservation: ActiveReservationModel | null) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: confirmReservation,
    onSuccess: () => {
      if (reservation) {
        void queryClient.invalidateQueries({ queryKey: inventoryQueryKeys.stock(reservation.itemId) });
      }
    },
  });
}
