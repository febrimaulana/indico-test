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

/** Renders any backend negative state (400/404/409/410/500) as a readable banner. */
export const ErrorBanner = memo(function ErrorBanner({ error, fallbackMessage = "terjadi kesalahan" }: ErrorBannerProps) {
  if (!error) {
    return null;
  }
  const { message, code, details } = describeError(error, fallbackMessage);

  return (
    <div
      role="alert"
      className="animate-slide-in flex gap-3 rounded-xl border border-rose-500/40 bg-linear-to-r from-rose-950/60 to-slate-950/40 p-3.5 text-sm"
    >
      <span className="mt-0.5 flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-rose-500/20">
        <svg viewBox="0 0 20 20" className="h-3.5 w-3.5" fill="none" aria-hidden="true">
          <path d="M10 6v5M10 14h.01" stroke="#fda4af" strokeWidth="2" strokeLinecap="round" />
        </svg>
      </span>
      <div className="min-w-0">
        <p className="flex flex-wrap items-center gap-2 font-semibold text-rose-100">
          <span>{message}</span>
          {code ? (
            <span className="rounded bg-rose-900/80 px-1.5 py-0.5 font-mono text-[10px] tracking-wide text-rose-300">
              {code}
            </span>
          ) : null}
        </p>
        {details.length > 0 ? (
          <ul className="mt-1 list-inside list-disc space-y-0.5 text-xs text-rose-300">
            {details.map((detail) => (
              <li key={detail}>{detail}</li>
            ))}
          </ul>
        ) : null}
      </div>
    </div>
  );
});
