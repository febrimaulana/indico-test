import axios from "axios";
import type { AxiosError, AxiosInstance } from "axios";
import { API_BASE_URL } from "../config";

/** One invalid field reported by the backend. */
export interface ApiErrorDetail {
  field?: string;
  rule?: string;
  message: string;
}

interface ApiErrorBody {
  status: "error";
  error: {
    code: string;
    message: string;
    details?: ApiErrorDetail[];
  };
}

/**
 * Normalized API error. `code` stays English for branching, `message` is the
 * Bahasa Indonesia message rendered to the user.
 */
export class ApiError extends Error {
  readonly code: string;
  readonly status: number;
  readonly details: ApiErrorDetail[];

  constructor(code: string, message: string, status: number, details: ApiErrorDetail[] = []) {
    super(message);
    this.name = "ApiError";
    this.code = code;
    this.status = status;
    this.details = details;
  }
}

/** The single axios instance used by the whole application. */
export const api: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10_000,
  headers: { "Content-Type": "application/json" },
});

api.interceptors.response.use(
  (response) => response,
  (error: AxiosError<ApiErrorBody>) => {
    const response = error.response;
    if (response) {
      const payload = response.data?.error;
      return Promise.reject(
        new ApiError(
          payload?.code ?? "INTERNAL_ERROR",
          payload?.message ?? "terjadi kesalahan pada server",
          response.status,
          payload?.details ?? [],
        ),
      );
    }
    if (error.code === "ECONNABORTED") {
      return Promise.reject(new ApiError("TIMEOUT", "koneksi ke server timeout", 0));
    }
    if (error.code === "ERR_CANCELED") {
      // Request aborted by TanStack Query on unmount: not a user-facing error.
      return Promise.reject(error);
    }
    return Promise.reject(new ApiError("NETWORK_ERROR", "tidak dapat terhubung ke server", 0));
  },
);
