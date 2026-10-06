import { api } from "../../lib/axios";
import type { ConfirmRequest, ConfirmResponse, ReserveRequest, ReserveResponse, StockResponse } from "./types";

/** GET /api/v1/inventory/stock */
export async function getStock(itemId: string, signal?: AbortSignal): Promise<StockResponse> {
  const response = await api.get<StockResponse>("/api/v1/inventory/stock", {
    params: { item_id: itemId },
    signal,
  });
  return response.data;
}

/** POST /api/v1/inventory/reserve */
export async function reserveStock(payload: ReserveRequest): Promise<ReserveResponse> {
  const response = await api.post<ReserveResponse>("/api/v1/inventory/reserve", payload);
  return response.data;
}

/** POST /api/v1/inventory/confirm */
export async function confirmReservation(payload: ConfirmRequest): Promise<ConfirmResponse> {
  const response = await api.post<ConfirmResponse>("/api/v1/inventory/confirm", payload);
  return response.data;
}
