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
    <div role="alert" className="rounded-lg border border-rose-500/40 bg-rose-950/60 p-3 text-sm text-rose-200">
      <p className="flex flex-wrap items-center gap-2 font-semibold">
        <span aria-hidden="true">⚠️</span>
        <span>{message}</span>
        {code ? (
          <span className="rounded bg-rose-900/80 px-1.5 py-0.5 font-mono text-[10px] tracking-wide text-rose-300">
            {code}
          </span>
        ) : null}
      </p>
      {details.length > 0 ? (
        <ul className="mt-1 list-inside list-disc space-y-0.5 text-rose-300">
          {details.map((detail) => (
            <li key={detail}>{detail}</li>
          ))}
        </ul>
      ) : null}
    </div>
  );
});
