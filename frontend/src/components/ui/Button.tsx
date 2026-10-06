import { memo } from "react";
import type { ButtonHTMLAttributes } from "react";

type ButtonVariant = "primary" | "secondary" | "danger";

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
}

const variantClasses: Record<ButtonVariant, string> = {
  primary: "bg-indigo-600 text-white hover:bg-indigo-500 disabled:bg-indigo-950 disabled:text-indigo-400/60",
  secondary: "bg-slate-700 text-slate-100 hover:bg-slate-600 disabled:bg-slate-800 disabled:text-slate-500",
  danger: "bg-emerald-600 text-white hover:bg-emerald-500 disabled:bg-emerald-950 disabled:text-emerald-400/60",
};

export const Button = memo(function Button({ variant = "primary", className = "", ...rest }: ButtonProps) {
  return (
    <button
      className={`inline-flex items-center justify-center rounded-lg px-4 py-2 text-sm font-semibold transition-colors disabled:cursor-not-allowed ${variantClasses[variant]} ${className}`}
      {...rest}
    />
  );
});
