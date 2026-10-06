import { memo } from "react";

const STEPS = [
  { title: "Isi user & jumlah", description: "Tentukan siapa pemesan dan berapa unit yang dikunci." },
  { title: "Reserve stok", description: "Server mengunci stok secara atomik — tidak bisa oversell." },
  { title: "Konfirmasi", description: "Selesaikan pembelian sebelum countdown 5 menit habis." },
];

export const EmptyState = memo(function EmptyState() {
  return (
    <section className="rounded-2xl border border-dashed border-slate-700/80 bg-slate-900/30 p-6">
      <div className="flex flex-col items-center text-center">
        <span className="animate-float-slow flex h-16 w-16 items-center justify-center rounded-2xl bg-linear-to-br from-indigo-500/20 to-slate-900 ring-1 ring-indigo-500/30 shadow-lg shadow-indigo-950/40">
          <svg viewBox="0 0 24 24" className="h-8 w-8" fill="none" aria-hidden="true">
            <path
              d="M4 7h16l-1.5 10.5a2 2 0 0 1-2 1.5H7.5a2 2 0 0 1-2-1.5L4 7zM4 7l1.5-3h13L20 7M9 11v4M15 11v4"
              stroke="#a5b4fc"
              strokeWidth="1.6"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
        </span>
        <h2 className="mt-4 text-lg font-semibold text-white">Belum ada reservasi aktif</h2>
        <p className="mt-1 max-w-md text-sm text-slate-400">
          Reservasi akan muncul di sini lengkap dengan countdown 5 menit dan tombol konfirmasi.
        </p>
      </div>

      <ol className="relative mt-6 grid gap-3 sm:grid-cols-3">
        <span
          aria-hidden="true"
          className="absolute left-8 right-8 top-8 hidden border-t border-dashed border-slate-700/70 sm:block"
        />
        {STEPS.map((step, index) => (
          <li
            key={step.title}
            className="group relative rounded-xl border border-slate-800/80 bg-slate-950/60 p-4 transition-all duration-200 hover:-translate-y-0.5 hover:border-indigo-500/40"
          >
            <span className="flex h-6 w-6 items-center justify-center rounded-full bg-indigo-500/20 text-xs font-bold text-indigo-300 ring-1 ring-indigo-500/30 transition-colors group-hover:bg-indigo-500/30">
              {index + 1}
            </span>
            <p className="mt-2.5 text-sm font-semibold text-slate-100">{step.title}</p>
            <p className="mt-1 text-xs leading-relaxed text-slate-400">{step.description}</p>
          </li>
        ))}
      </ol>
    </section>
  );
});
