import { memo, useCallback, useEffect } from "react";
import { Badge } from "../../../components/ui/Badge";
import { Button } from "../../../components/ui/Button";
import { Spinner } from "../../../components/ui/Spinner";
import { formatIsoClock } from "../../../utils/format";
import { useCountdown } from "../hooks/useCountdown";
import type { ActiveReservationModel } from "../types";
import { CountdownRing } from "./CountdownRing";
import { ErrorBanner } from "./ErrorBanner";

const RESERVATION_LIFETIME_SECONDS = 5 * 60;

interface ActiveReservationProps {
  reservation: ActiveReservationModel;
  isConfirming: boolean;
  error: unknown;
  onConfirm: () => void;
  onExpired: () => void;
  onClearExpired: () => void;
}

interface DetailItemProps {
  label: string;
  value: string;
  mono?: boolean;
}

const DetailItem = memo(function DetailItem({ label, value, mono = false }: DetailItemProps) {
  return (
    <div className="rounded-xl border border-slate-800/80 bg-slate-950/60 p-3 transition-colors hover:border-slate-700">
      <dt className="text-[10px] uppercase tracking-widest text-slate-500">{label}</dt>
      <dd className={`mt-1 truncate text-sm text-slate-100 ${mono ? "font-mono" : ""}`}>{value}</dd>
    </div>
  );
});

/** Countdown + Confirm Purchase workflow for the current reservation. */
export const ActiveReservation = memo(function ActiveReservation({
  reservation,
  isConfirming,
  error,
  onConfirm,
  onExpired,
  onClearExpired,
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

  const borderGradient = isExpired
    ? "from-amber-400/40 via-amber-500/10 to-transparent"
    : "from-indigo-400/40 via-violet-500/20 to-transparent";

  return (
    <section className={`animate-pop-in rounded-2xl bg-linear-to-b p-px shadow-2xl shadow-slate-950/60 ${borderGradient}`}>
      <div className={`rounded-[15px] p-5 backdrop-blur ${isExpired ? "bg-amber-950/25" : "bg-indigo-950/30"}`}>
        <header className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <div className="flex items-center gap-2">
              <h2 className="text-lg font-semibold text-white">Reservasi Aktif</h2>
              <Badge tone={isExpired ? "warning" : "success"} pulse={!isExpired}>
                {isExpired ? "kedaluwarsa" : "menunggu konfirmasi"}
              </Badge>
            </div>
            <p className="mt-1 text-xs text-slate-400">
              reservasi <span className="font-mono text-slate-300">{reservation.reservationId}</span>
            </p>
          </div>
        </header>

        <div className="mt-5 grid items-center gap-6 md:grid-cols-[auto_1fr]">
          <div className="justify-self-center">
            <CountdownRing
              remainingSeconds={remainingSeconds}
              totalSeconds={RESERVATION_LIFETIME_SECONDS}
              isExpired={isExpired}
            />
          </div>

          <div className="min-w-0">
            <dl className="grid gap-3 sm:grid-cols-2">
              <DetailItem label="Item" value={reservation.itemId} mono />
              <DetailItem label="Jumlah" value={String(reservation.quantity)} mono />
              <DetailItem label="User" value={reservation.userId ?? "-"} mono />
              <DetailItem label="Kedaluwarsa pukul" value={formatIsoClock(reservation.expiresAt)} />
            </dl>

            {isExpired ? (
              <div className="animate-slide-in mt-4 flex flex-wrap items-center gap-3 rounded-xl border border-amber-500/40 bg-amber-950/40 px-3.5 py-3 text-xs text-amber-200">
                <span>Reservasi ini sudah kedaluwarsa. Stoknya otomatis dikembalikan oleh server.</span>
                <Button variant="secondary" className="ml-auto" onClick={onClearExpired}>
                  Reservasi baru
                </Button>
              </div>
            ) : null}

            {error ? (
              <div className="mt-4">
                <ErrorBanner error={error} fallbackMessage="konfirmasi gagal" />
              </div>
            ) : null}

            <Button
              variant="success"
              className="mt-4 w-full"
              onClick={handleConfirm}
              disabled={isExpired || isConfirming}
            >
              {isConfirming ? (
                <Spinner label="Mengkonfirmasi..." />
              ) : (
                <>
                  Confirm Purchase
                  <svg viewBox="0 0 20 20" className="h-4 w-4" fill="none" aria-hidden="true">
                    <path
                      d="M4 10.5l4 4L16 6"
                      stroke="currentColor"
                      strokeWidth="1.8"
                      strokeLinecap="round"
                      strokeLinejoin="round"
                    />
                  </svg>
                </>
              )}
            </Button>
          </div>
        </div>
      </div>
    </section>
  );
});
