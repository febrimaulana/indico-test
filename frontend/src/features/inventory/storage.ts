import { ACTIVE_RESERVATION_STORAGE_KEY } from "../../config";
import type { ActiveReservationModel } from "./types";

const MAX_STORED_RESERVATIONS = 50;
// Drop entries that expired more than an hour ago (keeps storage tidy).
const STALE_AFTER_MS = 60 * 60 * 1000;

function isReservation(value: unknown): value is ActiveReservationModel {
  return (
    typeof value === "object" &&
    value !== null &&
    "reservationId" in value &&
    "itemId" in value &&
    "quantity" in value &&
    "expiresAt" in value
  );
}

/**
 * Restore active reservations after a page refresh (best effort). Older builds
 * stored a single reservation object, so both shapes are accepted.
 */
export function loadStoredReservations(): ActiveReservationModel[] {
  try {
    const raw = window.localStorage.getItem(ACTIVE_RESERVATION_STORAGE_KEY);
    if (!raw) {
      return [];
    }
    const parsed: unknown = JSON.parse(raw);
    const candidates = Array.isArray(parsed) ? parsed : [parsed];
    const now = Date.now();

    return candidates
      .filter(isReservation)
      .filter((reservation) => {
        const expiry = new Date(reservation.expiresAt).getTime();
        return Number.isFinite(expiry) ? expiry + STALE_AFTER_MS > now : true;
      })
      .slice(-MAX_STORED_RESERVATIONS);
  } catch {
    return [];
  }
}

/** Persist the active reservations across refreshes. */
export function storeReservations(reservations: ActiveReservationModel[]): void {
  try {
    window.localStorage.setItem(
      ACTIVE_RESERVATION_STORAGE_KEY,
      JSON.stringify(reservations.slice(-MAX_STORED_RESERVATIONS)),
    );
  } catch {
    // Storage can be unavailable (private mode): ignoring is safe.
  }
}

/** Remove every stored reservation. */
export function clearStoredReservations(): void {
  try {
    window.localStorage.removeItem(ACTIVE_RESERVATION_STORAGE_KEY);
  } catch {
    // ignore
  }
}
