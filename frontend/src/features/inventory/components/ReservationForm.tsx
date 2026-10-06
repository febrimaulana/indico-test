import { memo, useCallback, useState } from "react";
import type { ChangeEvent, FormEvent } from "react";
import { Button } from "../../../components/ui/Button";
import { Spinner } from "../../../components/ui/Spinner";
import type { ReservationFormInput } from "../types";
import { ErrorBanner } from "./ErrorBanner";

interface ReservationFormProps {
  itemId: string;
  isPending: boolean;
  error: unknown;
  onReserve: (input: ReservationFormInput) => void;
}

export const ReservationForm = memo(function ReservationForm({ itemId, isPending, error, onReserve }: ReservationFormProps) {
  const [userId, setUserId] = useState("usr_9981");
  const [quantity, setQuantity] = useState(1);

  const handleUserIdChange = useCallback((event: ChangeEvent<HTMLInputElement>) => setUserId(event.target.value), []);
  const handleQuantityChange = useCallback((event: ChangeEvent<HTMLInputElement>) => setQuantity(event.target.valueAsNumber), []);

  const handleSubmit = useCallback(
    (event: FormEvent<HTMLFormElement>) => {
      event.preventDefault();
      onReserve({ userId: userId.trim(), quantity });
    },
    [onReserve, quantity, userId],
  );

  const isValidUserId = userId.trim().length >= 3;
  const isValidQuantity = Number.isFinite(quantity) && quantity > 0 && quantity <= 1000;

  return (
    <section className="rounded-2xl border border-slate-800 bg-slate-900/60 p-5 shadow-xl">
      <h2 className="text-lg font-semibold text-white">Reservation Form</h2>
      <p className="text-xs text-slate-400">
        item: <span className="font-mono text-slate-300">{itemId}</span> · reservasi kedaluwarsa otomatis dalam 5 menit
      </p>

      <form className="mt-4 space-y-4" onSubmit={handleSubmit}>
        <div>
          <label className="mb-1 block text-xs font-medium uppercase tracking-wide text-slate-400" htmlFor="user-id">
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
            className="w-full rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 font-mono text-sm text-slate-100 outline-none focus:border-indigo-500"
          />
        </div>

        <div>
          <label className="mb-1 block text-xs font-medium uppercase tracking-wide text-slate-400" htmlFor="quantity">
            Quantity
          </label>
          <input
            id="quantity"
            name="quantity"
            type="number"
            min={1}
            max={1000}
            value={Number.isFinite(quantity) ? quantity : ""}
            onChange={handleQuantityChange}
            className="w-full rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 font-mono text-sm text-slate-100 outline-none focus:border-indigo-500"
          />
        </div>

        {error ? <ErrorBanner error={error} fallbackMessage="reservasi gagal" /> : null}

        <Button type="submit" className="w-full" disabled={!isValidUserId || !isValidQuantity || isPending}>
          {isPending ? <Spinner label="Memproses reservasi..." /> : "Reserve Stock"}
        </Button>
      </form>
    </section>
  );
});
