import { memo, useCallback, useEffect } from "react";
import { Button } from "../../../components/ui/Button";
import { Spinner } from "../../../components/ui/Spinner";
import { useCountdown } from "../hooks/useCountdown";
import type { ActiveReservationModel } from "../types";
import { CountdownTimer } from "./CountdownTimer";
import { ErrorBanner } from "./ErrorBanner";

interface ActiveReservationProps {
  reservation: ActiveReservationModel;
  isConfirming: boolean;
  error: unknown;
  onConfirm: () => void;
  onExpired: () => void;
}

/** Countdown + Confirm Purchase workflow for the current reservation. */
export const ActiveReservation = memo(function ActiveReservation({
  reservation,
  isConfirming,
  error,
  onConfirm,
  onExpired,
}: ActiveReservationProps) {
  const { remainingSeconds, isExpired } = useCountdown(reservation.expiresAt);

  // Notify the dashboard exactly once when this reservation expires; the
  // effect re-runs when a new reservation (new expiresAt) replaces it.
  useEffect(() => {
    if (isExpired) {
      onExpired();
    }
  }, [isExpired, onExpired]);

  const handleConfirm = useCallback(() => {
    if (!isExpired) {
      onConfirm();
    }
  }, [isExpired, onConfirm]);

  return (
    <section className="rounded-2xl border border-indigo-500/30 bg-indigo-950/30 p-5 shadow-xl">
      <header className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 className="text-lg font-semibold text-white">Active Reservation</h2>
          <p className="font-mono text-xs text-slate-400">{reservation.reservationId}</p>
        </div>
        <CountdownTimer remainingSeconds={remainingSeconds} isExpired={isExpired} />
      </header>

      <dl className="mt-4 grid grid-cols-2 gap-3 text-sm">
        <div className="rounded-lg bg-slate-900/70 p-3">
          <dt className="text-xs uppercase tracking-wide text-slate-400">Item</dt>
          <dd className="mt-1 font-mono text-slate-100">{reservation.itemId}</dd>
        </div>
        <div className="rounded-lg bg-slate-900/70 p-3">
          <dt className="text-xs uppercase tracking-wide text-slate-400">Quantity</dt>
          <dd className="mt-1 font-mono text-slate-100">{reservation.quantity}</dd>
        </div>
      </dl>

      {isExpired ? (
        <p className="mt-4 rounded-lg border border-amber-500/40 bg-amber-950/40 p-3 text-sm text-amber-200">
          Reservasi ini sudah kedaluwarsa. Stoknya otomatis dikembalikan — silakan buat reservasi baru.
        </p>
      ) : null}

      {error ? (
        <div className="mt-4">
          <ErrorBanner error={error} fallbackMessage="konfirmasi gagal" />
        </div>
      ) : null}

      <Button
        variant="danger"
        className="mt-4 w-full"
        onClick={handleConfirm}
        disabled={isExpired || isConfirming}
      >
        {isConfirming ? <Spinner label="Mengkonfirmasi..." /> : "Confirm Purchase"}
      </Button>
    </section>
  );
});
