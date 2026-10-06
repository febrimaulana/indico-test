import { lazy, Suspense } from "react";
import { InventoryDashboard } from "./features/inventory/InventoryDashboard";

// Loaded only in development: the dead branch is stripped from production builds.
const ReactQueryDevtools = lazy(() =>
  import("@tanstack/react-query-devtools").then((module) => ({ default: module.ReactQueryDevtools })),
);

export default function App() {
  return (
    <div className="min-h-screen bg-slate-950 text-slate-100">
      <InventoryDashboard />
      {import.meta.env.DEV ? (
        <Suspense fallback={null}>
          <ReactQueryDevtools initialIsOpen={false} />
        </Suspense>
      ) : null}
    </div>
  );
}
