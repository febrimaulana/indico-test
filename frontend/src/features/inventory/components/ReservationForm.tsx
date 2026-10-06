import { memo, useCallback, useState } from "react";
import type { ChangeEvent, FormEvent } from "react";
import { Badge } from "../../../components/ui/Badge";
import { Button } from "../../../components/ui/Button";
import { Spinner } from "../../../components/ui/Spinner";
import type { ReservationFormInput } from "../types";
import { ErrorBanner } from "./ErrorBanner";

const QUANTITY_PRESETS = [1, 2, 5, 10];
const MAX_QUANTITY = 1000;

interface ReservationFormProps {
  itemId: string;
  isPending: boolean;
  error: unknown;
  onReserve: (input: ReservationFormInput) => void;
}

export const ReservationForm = memo(function ReservationForm({
  itemId,
  isPending,
  error,
  onReserve,
}: ReservationFormProps) {
  const [userId, setUserId] = useState("usr_9981");
  const [quantity, setQuantity] = useState(1);

  const handleUserIdChange = useCallback((event: ChangeEvent<HTMLInputElement>) => setUserId(event.target.value), []);
  const handleQuantityInput = useCallback(
    (event: ChangeEvent<HTMLInputElement>) => setQuantity(event.target.valueAsNumber),
    [],
  );
  const decreaseQuantity = useCallback(
    () => setQuantity((previous) => Math.max(1, (Number.isFinite(previous) ? previous : 1) - 1)),
    [],
  );
  const increaseQuantity = useCallback(
    () => setQuantity((previous) => Math.min(MAX_QUANTITY, (Number.isFinite(previous) ? previous : 0) + 1)),
    [],
  );

  const handleSubmit = useCallback(
    (event: FormEvent<HTMLFormElement>) => {
      event.preventDefault();
      onReserve({ userId: userId.trim(), quantity });
    },
    [onReserve, quantity, userId],
  );

  const trimmedUserId = userId.trim();
  const isValidUserId = trimmedUserId.length >= 3 && trimmedUserId.length <= 64;
  const isValidQuantity = Number.isFinite(quantity) && quantity >= 1 && quantity <= MAX_QUANTITY;
  const canSubmit = isValidUserId && isValidQuantity && !isPending;

  return (
    <section className="h-full rounded-2xl bg-linear-to-b from-white/10 via-white/5 to-transparent p-px shadow-2xl shadow-slate-950/60">
      <div className="flex h-full flex-col rounded-[15px] bg-slate-900/70 p-5 backdrop-blur">
        <header className="flex flex-wrap items-start justify-between gap-3">
          <div>
            <h2 className="text-lg font-semibold text-white">Buat Reservasi</h2>
            <p className="mt-1 text-xs text-slate-400">
              Stok dikunci untuk Anda selama 5 menit. Bisa reservasi berkali-kali — setiap reservasi punya countdown sendiri.
            </p>
          </div>
          <Badge tone="info">
            <span className="font-mono">{itemId}</span>
          </Badge>
        </header>

        <form className="mt-5 flex flex-1 flex-col gap-5" onSubmit={handleSubmit}>
          <div>
            <label className="mb-1.5 block text-xs font-medium uppercase tracking-wider text-slate-400" htmlFor="user-id">
              User ID
            </label>
            <input
              id="user-id"
              name="user_id"
              type="text"
              value={userId}
              onChange={handleUserIdChange}
              placeholder="usr_9981"
              autoComplete="off"
              spellCheck={false}
              disabled={isPending}
              className={`w-full rounded-xl border bg-slate-950/80 px-3.5 py-2.5 font-mono text-sm text-slate-100 outline-none transition-all placeholder:text-slate-600 focus:ring-2 focus:ring-indigo-500/30 disabled:cursor-not-allowed disabled:opacity-50 ${
                isValidUserId || trimmedUserId.length === 0
                  ? "border-slate-700 focus:border-indigo-500"
                  : "border-rose-500/70 focus:border-rose-400"
              }`}
            />
            <p className="mt-1.5 text-[11px] text-slate-500">
              {isValidUserId ? "contoh: usr_9981 · 3–64 karakter (huruf, angka, _ , -)" : "minimal 3 karakter"}
            </p>
          </div>

          <div>
            <span className="mb-1.5 block text-xs font-medium uppercase tracking-wider text-slate-400">Jumlah</span>
            <div className="flex items-center gap-2">
              <Button
                type="button"
                variant="secondary"
                className="h-11 w-11 shrink-0 px-0 text-lg"
                onClick={decreaseQuantity}
                disabled={isPending}
                aria-label="Kurangi jumlah"
              >
                −
              </Button>
              <input
                id="quantity"
                name="quantity"
                type="number"
                min={1}
                max={MAX_QUANTITY}
                value={Number.isFinite(quantity) ? quantity : ""}
                onChange={handleQuantityInput}
                disabled={isPending}
                className="h-11 w-full rounded-xl border border-slate-700 bg-slate-950/80 px-3 text-center font-mono text-base text-slate-100 outline-none transition-all focus:border-indigo-500 focus:ring-2 focus:ring-indigo-500/30 disabled:cursor-not-allowed disabled:opacity-50"
              />
              <Button
                type="button"
                variant="secondary"
                className="h-11 w-11 shrink-0 px-0 text-lg"
                onClick={increaseQuantity}
                disabled={isPending}
                aria-label="Tambah jumlah"
              >
                +
              </Button>
            </div>

            <div className="mt-2.5 flex flex-wrap gap-2">
              {QUANTITY_PRESETS.map((preset) => (
                <button
                  key={preset}
                  type="button"
                  onClick={() => setQuantity(preset)}
                  disabled={isPending}
                  className={`rounded-lg border px-3 py-1 font-mono text-xs transition-all duration-150 active:scale-95 disabled:cursor-not-allowed disabled:opacity-50 ${
                    quantity === preset
                      ? "border-indigo-400/70 bg-linear-to-r from-indigo-500/30 to-violet-500/30 text-indigo-100 shadow-sm shadow-indigo-950/40"
                      : "border-slate-700 text-slate-400 hover:border-slate-500 hover:text-slate-200"
                  }`}
                >
                  {preset}
                </button>
              ))}
            </div>
          </div>

          {error ? <ErrorBanner error={error} fallbackMessage="reservasi gagal" /> : null}

          <Button type="submit" className="mt-auto w-full" disabled={!canSubmit}>
            {isPending ? (
              <Spinner label="Memproses reservasi..." />
            ) : (
              <>
                Reserve Stock
                <svg viewBox="0 0 20 20" className="h-4 w-4" fill="none" aria-hidden="true">
                  <path
                    d="M4 10h12M11 5l5 5-5 5"
                    stroke="currentColor"
                    strokeWidth="1.8"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                  />
                </svg>
              </>
            )}
          </Button>
        </form>
      </div>
    </section>
  );
});
