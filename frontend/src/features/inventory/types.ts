/** POST /api/v1/inventory/reserve request. */
export interface ReserveRequest {
  user_id: string;
  item_id: string;
  quantity: number;
}

/** POST /api/v1/inventory/reserve success response. */
export interface ReserveResponse {
  status: "success";
  reservation_id: string;
  item_id: string;
  quantity: number;
  expires_at: string;
}

/** POST /api/v1/inventory/confirm request. */
export interface ConfirmRequest {
  reservation_id: string;
}

/** POST /api/v1/inventory/confirm success response. */
export interface ConfirmResponse {
  status: "success";
  reservation_id: string;
  confirmed_at: string;
}

/** GET /api/v1/inventory/stock success response. */
export interface StockResponse {
  item_id: string;
  total_stock: number;
  reserved_stock: number;
  available_stock: number;
}

/** Reservation currently held by the user in the dashboard. */
export interface ActiveReservationModel {
  reservationId: string;
  itemId: string;
  quantity: number;
  expiresAt: string;
}

/** Input collected by the reservation form. */
export interface ReservationFormInput {
  userId: string;
  quantity: number;
}
