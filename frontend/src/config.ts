const rawBaseUrl: unknown = import.meta.env.VITE_API_BASE_URL;

/**
 * Empty by default: requests go to a relative /api path, proxied by Vite in
 * development and by nginx in the Docker image.
 */
export const API_BASE_URL: string = typeof rawBaseUrl === "string" ? rawBaseUrl : "";

/** Stock polling interval (ms). */
export const STOCK_POLL_INTERVAL_MS = 5_000;

/** Item shown in the mini dashboard. */
export const DEFAULT_ITEM_ID = "item_4021";

/** localStorage key used to survive page refreshes with an active reservation. */
export const ACTIVE_RESERVATION_STORAGE_KEY = "indico.activeReservation";
