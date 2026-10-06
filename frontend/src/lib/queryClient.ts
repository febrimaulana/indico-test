import { QueryClient } from "@tanstack/react-query";
import { ApiError } from "./axios";

const MAX_RETRIES = 2;

/**
 * Server state lives here. 4xx errors (insufficient stock, expired
 * reservation, validation) are never retried; network/5xx errors get two
 * extra attempts.
 */
export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 2_000,
      refetchOnWindowFocus: true,
      retry: (failureCount, error) => {
        if (error instanceof ApiError && error.status >= 400 && error.status < 500) {
          return false;
        }
        return failureCount < MAX_RETRIES;
      },
    },
    mutations: {
      retry: false,
    },
  },
});
