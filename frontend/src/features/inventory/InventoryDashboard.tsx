import { useCallback, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { DEFAULT_ITEM_ID } from "../../config";
import { ApiError } from "../../lib/axios";
import { ActiveReservation } from "./components/ActiveReservation";
import { EmptyState } from "./components/EmptyState";
import { ReservationForm } from "./components/ReservationForm";
import { StockPanel } from "./components/StockPanel";
import { SuccessBanner } from "./components/SuccessBanner";
import { TopBar } from "./components/TopBar";
import { useConfirmMutation } from "./hooks/useConfirmMutation";
import { useReserveMutation } from "./hooks/useReserveMutation";
import { useStockQuery } from "./hooks/useStockQuery";
import { inventoryQueryKeys } from "./queryKeys";
import { clearStoredReservation, loadStoredReservation, storeReservation } from "./storage";
import type { ActiveReservationModel, ReservationFormInput } from "./types";

/** Errors after which the local reservation is no longer usable. */
const TERMINAL_ERROR_CODES = new Set(["RESERVATION_EXPIRED", "RESERVATION_NOT_FOUND", "RESERVATION_ALREADY_CONFIRMED"]);

interface ConfirmationReceipt {
  reservationId: string;
  confirmedAt: string;
}

export function InventoryDashboard() {
  const itemId = DEFAULT_ITEM_ID;
  const queryClient = useQueryClient();

  const stockQuery = useStockQuery(itemId);
  const reserveMutation = useReserveMutation();
  const [activeReservation, setActiveReservation] = useState<ActiveReservationModel | null>(() => loadStoredReservation());
  const [receipt, setReceipt] = useState<ConfirmationReceipt | null>(null);
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
              userId: input.userId,
            };
            setActiveReservation(reservation);
            setReceipt(null);
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
        onSuccess: (response) => {
          setReceipt({ reservationId: response.reservation_id, confirmedAt: response.confirmed_at });
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

  const handleClearExpired = useCallback(() => {
    setActiveReservation(null);
    clearStoredReservation();
  }, []);

  const handleDismissReceipt = useCallback(() => setReceipt(null), []);

  const lastUpdatedAt = stockQuery.dataUpdatedAt > 0 ? stockQuery.dataUpdatedAt : null;

  return (
    <div className="relative min-h-screen">
      <div aria-hidden="true" className="pointer-events-none absolute inset-0 overflow-hidden">
        <div className="bg-grid absolute inset-x-0 top-0 h-80 opacity-[0.10] [mask-image:radial-gradient(ellipse_at_top,black_35%,transparent_75%)]" />
        <div className="animate-float-slow absolute -left-32 -top-24 h-80 w-80 rounded-full bg-indigo-600/25 blur-3xl" />
        <div
          className="animate-float-slow absolute -top-10 right-0 h-72 w-72 rounded-full bg-violet-600/20 blur-3xl"
          style={{ animationDelay: "-4.5s" }}
        />
        <div
          className="animate-float-slow absolute left-1/3 top-40 h-64 w-64 rounded-full bg-fuchsia-600/10 blur-3xl"
          style={{ animationDelay: "-2s" }}
        />
      </div>

      <div className="relative">
        <TopBar />

        <main className="mx-auto max-w-6xl px-4 py-10">
          <section className="animate-fade-up mx-auto max-w-3xl text-center">
            <p className="text-xs font-semibold uppercase tracking-[0.35em] text-indigo-400">INDICO Flash Sale</p>
            <h1 className="mt-3 text-3xl font-bold tracking-tight text-white sm:text-5xl">
              Kunci stok sebelum{" "}
              <span className="text-gradient-animated bg-linear-to-r from-indigo-300 via-violet-300 to-indigo-300 bg-clip-text text-transparent">
                kehabisan
              </span>
              .
            </h1>
            <p className="mx-auto mt-4 max-w-2xl text-sm leading-relaxed text-slate-400 sm:text-base">
              Reservasi atomik tanpa oversell, countdown 5 menit, dan stok kembali otomatis kalau tidak dikonfirmasi.
            </p>
            <div className="mt-5 flex flex-wrap items-center justify-center gap-2">
              <span className="rounded-full border border-slate-700/80 bg-slate-900/60 px-3 py-1 text-[11px] font-medium text-slate-300">
                ⚡ Anti-oversell
              </span>
              <span className="rounded-full border border-slate-700/80 bg-slate-900/60 px-3 py-1 text-[11px] font-medium text-slate-300">
                ⏱️ Auto-expiry 5 menit
              </span>
              <span className="rounded-full border border-slate-700/80 bg-slate-900/60 px-3 py-1 text-[11px] font-medium text-slate-300">
                🧪 100% test coverage
              </span>
            </div>
          </section>

          {receipt ? (
            <div className="mt-8">
              <SuccessBanner
                reservationId={receipt.reservationId}
                confirmedAt={receipt.confirmedAt}
                onDismiss={handleDismissReceipt}
              />
            </div>
          ) : null}

          <div className="mt-8 grid gap-6 lg:grid-cols-5">
            <div className="animate-fade-up lg:col-span-3">
              <StockPanel
                itemId={itemId}
                stock={stockQuery.data}
                isFetching={stockQuery.isFetching}
                isPending={stockQuery.isPending}
                error={stockQuery.error}
                lastUpdatedAt={lastUpdatedAt}
                onRefresh={handleRefresh}
              />
            </div>
            <div className="animate-fade-up lg:col-span-2" style={{ animationDelay: "0.08s" }}>
              <ReservationForm
                itemId={itemId}
                isPending={reserveMutation.isPending}
                error={reserveMutation.error}
                hasActiveReservation={activeReservation !== null}
                onReserve={handleReserve}
              />
            </div>
          </div>

          <div className="animate-fade-up mt-6" style={{ animationDelay: "0.16s" }}>
            {activeReservation ? (
              <ActiveReservation
                reservation={activeReservation}
                isConfirming={confirmMutation.isPending}
                error={confirmMutation.error}
                onConfirm={handleConfirm}
                onExpired={handleExpired}
                onClearExpired={handleClearExpired}
              />
            ) : (
              <EmptyState />
            )}
          </div>

          <footer className="mt-10 flex flex-wrap items-center justify-center gap-x-4 gap-y-2 text-center text-xs text-slate-500">
            <span>Auto-refresh tiap 5 detik</span>
            <span aria-hidden="true">·</span>
            <span className="font-mono">/api/v1</span>
            <span aria-hidden="true">·</span>
            <a
              href="/swagger/index.html"
              target="_blank"
              rel="noreferrer"
              className="text-slate-400 underline-offset-2 transition-colors hover:text-slate-200 hover:underline"
            >
              Swagger UI
            </a>
          </footer>
        </main>
      </div>
    </div>
  );
}
