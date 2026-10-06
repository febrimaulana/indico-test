import { memo } from "react";

interface SpinnerProps {
  label?: string;
}

export const Spinner = memo(function Spinner({ label }: SpinnerProps) {
  return (
    <span className="inline-flex items-center gap-2 text-sm text-slate-300" role="status" aria-live="polite">
      <span
        className="h-4 w-4 animate-spin rounded-full border-2 border-slate-500 border-t-transparent"
        aria-hidden="true"
      />
      {label ? <span>{label}</span> : null}
    </span>
  );
});
