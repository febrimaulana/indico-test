import { ACTIVE_RESERVATION_STORAGE_KEY } from "../../config";
import type { ActiveReservationModel } from "./types";

/** Restore the active reservation after a page refresh (best effort). */
export function loadStoredReservation(): ActiveReservationModel | null {
  try {
    const raw = window.localStorage.getItem(ACTIVE_RESERVATION_STORAGE_KEY);
    if (!raw) {
      return null;
    }
    const parsed: unknown = JSON.parse(raw);
    if (
      typeof parsed === "object" &&
      parsed !== null &&
      "reservationId" in parsed &&
      "itemId" in parsed &&
      "quantity" in parsed &&
      "expiresAt" in parsed
    ) {
      return parsed as ActiveReservationModel;
    }
    return null;
  } catch {
    return null;
  }
}

/** Persist the active reservation across refreshes. */
export function storeReservation(reservation: ActiveReservationModel): void {
  try {
    window.localStorage.setItem(ACTIVE_RESERVATION_STORAGE_KEY, JSON.stringify(reservation));
  } catch {
    // Storage can be unavailable (private mode): ignoring is safe.
  }
}

/** Remove the stored reservation. */
export function clearStoredReservation(): void {
  try {
    window.localStorage.removeItem(ACTIVE_RESERVATION_STORAGE_KEY);
  } catch {
    // ignore
  }
}
