# Frontend Rules (React + Vite + TypeScript + Tailwind)

Feature-first. One feature = one folder in `src/features/<domain>/`.

## Required conventions

| Path | Responsibility |
|---|---|
| `src/lib/axios.ts` | THE single axios instance. Base URL, timeout, error-envelope normalization. |
| `src/lib/queryClient.ts` | TanStack Query client config (retry policy, staleTime). |
| `src/features/<domain>/api.ts` | Typed API functions using the axios instance. |
| `src/features/<domain>/types.ts` | All response/error types for the feature. |
| `src/features/<domain>/queryKeys.ts` | Centralized query keys — no string literals scattered around. |
| `src/features/<domain>/hooks/` | `useQuery` / `useMutation` wrappers + local hooks. |
| `src/features/<domain>/components/` | Feature components. Leaf components wrapped in `React.memo`. |
| `src/components/ui/` | Generic, domain-agnostic components only. |

## Hard Rules

1. **All HTTP via `src/lib/axios.ts`.** No `fetch` elsewhere, no second axios instance.
2. **All server state via TanStack Query.** Never store API data in `useState`; never fetch in raw `useEffect`.
3. **Query keys centralized** in `queryKeys.ts`.
4. **Every `useMutation` must invalidate related queries in `onSuccess`.**
5. **Retry only network/5xx errors.** Never retry 4xx (409 insufficient stock, 410 expired).
6. Every `useEffect` MUST clean up: `clearInterval`/`clearTimeout`, `AbortController.abort()`, `removeEventListener`.
7. Countdown timers derive from the server `expires_at` (UTC), not local countdown state.
8. `React.memo` on leaf components; `useCallback` for handlers passed to memoized children; `useMemo` for derived values.
9. All UI error rendering uses the backend's Bahasa Indonesia `message`; error codes stay English for branching.
10. Tailwind utilities only. No new CSS files besides `index.css`.

## Quality Gate

```bash
npx tsc --noEmit   # zero errors
npm run build      # production build succeeds
```
