import { memo } from "react";
import { ApiError } from "../../../lib/axios";

interface ErrorBannerProps {
  error: unknown;
  fallbackMessage?: string;
}

interface ErrorDescription {
  message: string;
  code: string | null;
  details: string[];
}

function describeError(error: unknown, fallbackMessage: string): ErrorDescription {
  if (error instanceof ApiError) {
    return {
      message: error.message,
      code: error.code,
      details: error.details.map((detail) => detail.message),
    };
  }
  if (error instanceof Error) {
    return { message: error.message, code: null, details: [] };
  }
  return { message: fallbackMessage, code: null, details: [] };
}

/**
 * Renders any backend negative state as a centered, user-friendly banner.
 * The user only sees the Bahasa Indonesia message; the English error code is
 * developer-facing (it stays in the API response, Swagger and the code
 * branching) and is shown in the UI only during development.
 */
export const ErrorBanner = memo(function ErrorBanner({ error, fallbackMessage = "terjadi kesalahan" }: ErrorBannerProps) {
  if (!error) {
    return null;
  }
  const { message, code, details } = describeError(error, fallbackMessage);

  return (
    <div
      role="alert"
      className="animate-slide-in rounded-xl border border-rose-500/40 bg-linear-to-r from-rose-950/60 to-slate-950/40 px-4 py-3.5 text-sm"
    >
      <p className="flex items-center justify-center gap-2 text-center font-semibold text-rose-100">
        <svg viewBox="0 0 20 20" className="h-4 w-4 shrink-0" fill="none" aria-hidden="true">
          <path d="M10 6v5M10 14h.01" stroke="#fda4af" strokeWidth="2" strokeLinecap="round" />
        </svg>
        <span>{message}</span>
      </p>

      {details.length > 0 ? (
        <ul className="mt-1.5 space-y-0.5 text-center text-xs text-rose-300">
          {details.map((detail) => (
            <li key={detail}>{detail}</li>
          ))}
        </ul>
      ) : null}

      {import.meta.env.DEV && code ? (
        <p className="mt-2 text-center font-mono text-[10px] tracking-wide text-rose-400/70">dev · {code}</p>
      ) : null}
    </div>
  );
});
