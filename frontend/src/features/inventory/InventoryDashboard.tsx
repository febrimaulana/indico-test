import { useCallback, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { DEFAULT_ITEM_ID } from "../../config";
import { ApiError } from "../../lib/axios";
import { ActiveReservation } from "./components/ActiveReservation";
import { InventoryCard } from "./components/InventoryCard";
import { ReservationForm } from "./components/ReservationForm";
import { useConfirmMutation } from "./hooks/useConfirmMutation";
import { useReserveMutation } from "./hooks/useReserveMutation";
import { useStockQuery } from "./hooks/useStockQuery";
import { inventoryQueryKeys } from "./queryKeys";
import { clearStoredReservation, loadStoredReservation, storeReservation } from "./storage";
import type { ActiveReservationModel, ReservationFormInput } from "./types";

/** Errors after which the local reservation is no longer usable. */
const TERMINAL_ERROR_CODES = new Set(["RESERVATION_EXPIRED", "RESERVATION_NOT_FOUND", "RESERVATION_ALREADY_CONFIRMED"]);

export function InventoryDashboard() {
  const itemId = DEFAULT_ITEM_ID;
  const queryClient = useQueryClient();

  const stockQuery = useStockQuery(itemId);
  const reserveMutation = useReserveMutation();
  const [activeReservation, setActiveReservation] = useState<ActiveReservationModel | null>(() => loadStoredReservation());
  const confirmMutation = useConfirmMutation(activeReservation);

  const { refetch } = stockQuery;
  const handleRefresh = useCallback(() => {
    void refetch();
  }, [refetch]);

  const handleReserve = useCallback(
    (input: ReservationFormInput) => {
      reserveMutation.mutate(
        { user_id: input.userId, item_id: itemId, quantity: input.quantity },
        {
          onSuccess: (response) => {
            const reservation: ActiveReservationModel = {
              reservationId: response.reservation_id,
              itemId: response.item_id,
              quantity: response.quantity,
              expiresAt: response.expires_at,
            };
            setActiveReservation(reservation);
            storeReservation(reservation);
          },
        },
      );
    },
    [itemId, reserveMutation],
  );

  const handleConfirm = useCallback(() => {
    if (!activeReservation) {
      return;
    }
    confirmMutation.mutate(
      { reservation_id: activeReservation.reservationId },
      {
        onSuccess: () => {
          setActiveReservation(null);
          clearStoredReservation();
        },
        onError: (error) => {
          if (error instanceof ApiError && TERMINAL_ERROR_CODES.has(error.code)) {
            setActiveReservation(null);
            clearStoredReservation();
          }
        },
      },
    );
  }, [activeReservation, confirmMutation]);

  const handleExpired = useCallback(() => {
    // The sweeper returns the stock; refresh the tracker so the UI matches.
    void queryClient.invalidateQueries({ queryKey: inventoryQueryKeys.stock(itemId) });
  }, [itemId, queryClient]);

  const lastUpdatedAt = stockQuery.dataUpdatedAt > 0 ? stockQuery.dataUpdatedAt : null;

  return (
    <main className="mx-auto max-w-5xl px-4 py-10">
      <header className="text-center">
        <p className="text-xs font-semibold uppercase tracking-[0.3em] text-indigo-400">INDICO Flash Sale</p>
        <h1 className="mt-2 text-3xl font-bold text-white sm:text-4xl">Inventory Reservation Dashboard</h1>
        <p className="mx-auto mt-3 max-w-2xl text-sm text-slate-400">
          Reservasi stok secara atomik, pantau countdown 5 menit, lalu konfirmasi pembelian sebelum kedaluwarsa.
        </p>
      </header>

      <div className="mt-8 grid gap-6 lg:grid-cols-2">
        <InventoryCard
          itemId={itemId}
          stock={stockQuery.data}
          isFetching={stockQuery.isFetching}
          isPending={stockQuery.isPending}
          error={stockQuery.error}
          lastUpdatedAt={lastUpdatedAt}
          onRefresh={handleRefresh}
        />
        <ReservationForm
          itemId={itemId}
          isPending={reserveMutation.isPending}
          error={reserveMutation.error}
          onReserve={handleReserve}
        />
      </div>

      {activeReservation ? (
        <div className="mt-6">
          <ActiveReservation
            reservation={activeReservation}
            isConfirming={confirmMutation.isPending}
            error={confirmMutation.error}
            onConfirm={handleConfirm}
            onExpired={handleExpired}
          />
        </div>
      ) : null}

      <footer className="mt-10 text-center text-xs text-slate-500">
        Auto-refresh tiap 5 detik · API &amp; Swagger: <span className="font-mono">/api/v1</span> dan{" "}
        <span className="font-mono">/swagger/index.html</span>
      </footer>
    </main>
  );
}
