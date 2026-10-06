import { memo } from "react";
import { Badge } from "../../../components/ui/Badge";
import { STOCK_POLL_INTERVAL_MS } from "../../../config";

const POLL_SECONDS = Math.round(STOCK_POLL_INTERVAL_MS / 1000);

export const TopBar = memo(function TopBar() {
  return (
    <header className="sticky top-0 z-20 border-b border-slate-800/70 bg-slate-950/75 backdrop-blur">
      <div className="mx-auto flex max-w-6xl items-center justify-between gap-4 px-4 py-3">
        <div className="flex items-center gap-3">
          <span className="flex h-9 w-9 items-center justify-center rounded-xl bg-linear-to-br from-indigo-500 to-violet-600 shadow-lg shadow-indigo-950/50 transition-transform duration-300 hover:rotate-6">
            <svg viewBox="0 0 24 24" className="h-5 w-5" fill="none" aria-hidden="true">
              <path d="M4 8l8-4 8 4v8l-8 4-8-4V8z" stroke="white" strokeWidth="1.8" strokeLinejoin="round" />
              <path d="M4 8l8 4 8-4M12 12v8" stroke="white" strokeWidth="1.8" strokeLinejoin="round" />
            </svg>
          </span>
          <div className="leading-tight">
            <p className="text-sm font-bold tracking-tight text-white">INDICO</p>
            <p className="text-[11px] text-slate-400">Flash Sale Inventory</p>
          </div>
        </div>

        <div className="flex items-center gap-3">
          <Badge tone="success" pulse>
            LIVE · poll {POLL_SECONDS}s
          </Badge>
          <a
            href="/swagger/index.html"
            target="_blank"
            rel="noreferrer"
            className="hidden rounded-lg border border-slate-700/80 px-3 py-1.5 text-xs font-medium text-slate-300 transition-all hover:border-slate-500 hover:text-white hover:shadow-md hover:shadow-indigo-950/40 sm:inline-flex"
          >
            API Docs ↗
          </a>
        </div>
      </div>
      <div className="h-px bg-linear-to-r from-transparent via-indigo-500/50 to-transparent" aria-hidden="true" />
    </header>
  );
});
