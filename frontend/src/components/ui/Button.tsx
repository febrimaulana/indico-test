import { memo } from "react";
import type { ButtonHTMLAttributes } from "react";

type ButtonVariant = "primary" | "secondary" | "success" | "ghost";

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
}

const variantClasses: Record<ButtonVariant, string> = {
  primary:
    "bg-linear-to-r from-indigo-500 to-violet-500 text-white shadow-lg shadow-indigo-950/50 hover:from-indigo-400 hover:to-violet-400 disabled:from-slate-800 disabled:to-slate-800 disabled:text-slate-500 disabled:shadow-none",
  secondary:
    "border border-slate-700/80 bg-slate-900/70 text-slate-200 hover:border-slate-500 hover:bg-slate-800/80 disabled:border-slate-800 disabled:text-slate-600",
  success:
    "bg-linear-to-r from-emerald-500 to-teal-500 text-white shadow-lg shadow-emerald-950/50 hover:from-emerald-400 hover:to-teal-400 disabled:from-slate-800 disabled:to-slate-800 disabled:text-slate-500 disabled:shadow-none",
  ghost: "text-slate-300 hover:bg-slate-800/70 hover:text-white disabled:text-slate-600",
};

export const Button = memo(function Button({ variant = "primary", className = "", ...rest }: ButtonProps) {
  const shine = variant === "primary" || variant === "success" ? "btn-shine" : "";
  return (
    <button
      className={`inline-flex items-center justify-center gap-2 rounded-xl px-4 py-2.5 text-sm font-semibold transition-all duration-150 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-400 active:scale-[0.98] disabled:cursor-not-allowed disabled:active:scale-100 ${variantClasses[variant]} ${shine} ${className}`}
      {...rest}
    />
  );
});
