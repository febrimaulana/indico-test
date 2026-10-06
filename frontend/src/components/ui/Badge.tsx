import { memo } from "react";
import type { ReactNode } from "react";

export type BadgeTone = "neutral" | "info" | "success" | "warning" | "danger";

interface BadgeProps {
  tone?: BadgeTone;
  pulse?: boolean;
  children: ReactNode;
}

const toneClasses: Record<BadgeTone, string> = {
  neutral: "border-slate-600/60 bg-slate-800/70 text-slate-300",
  info: "border-indigo-500/40 bg-indigo-500/15 text-indigo-300",
  success: "border-emerald-500/40 bg-emerald-500/15 text-emerald-300",
  warning: "border-amber-500/40 bg-amber-500/15 text-amber-300",
  danger: "border-rose-500/40 bg-rose-500/15 text-rose-300",
};

const dotClasses: Record<BadgeTone, string> = {
  neutral: "bg-slate-400",
  info: "bg-indigo-400",
  success: "bg-emerald-400",
  warning: "bg-amber-400",
  danger: "bg-rose-400",
};

export const Badge = memo(function Badge({ tone = "neutral", pulse = false, children }: BadgeProps) {
  return (
    <span
      className={`inline-flex items-center gap-1.5 rounded-full border px-2.5 py-1 text-[11px] font-semibold tracking-wide ${toneClasses[tone]}`}
    >
      <span
        className={`h-1.5 w-1.5 rounded-full ${dotClasses[tone]} ${pulse ? "animate-pulse" : ""}`}
        aria-hidden="true"
      />
      {children}
    </span>
  );
});
