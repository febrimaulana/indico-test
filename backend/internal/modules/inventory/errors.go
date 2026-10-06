package inventory

import (
	"errors"
	"net/http"

	"indico/backend/internal/platform/httpx"
)

// Domain sentinel errors. Handlers map these to the frozen HTTP envelope.
var (
	ErrItemNotFound                = httpx.New(http.StatusNotFound, httpx.CodeItemNotFound, "item tidak ditemukan")
	ErrInsufficientStock           = httpx.New(http.StatusConflict, httpx.CodeInsufficientStock, "stok tidak mencukupi")
	ErrReservationNotFound         = httpx.New(http.StatusNotFound, httpx.CodeReservationNotFound, "reservasi tidak ditemukan")
	ErrReservationExpired          = httpx.New(http.StatusGone, httpx.CodeReservationExpired, "reservasi sudah kedaluwarsa")
	ErrReservationAlreadyConfirmed = httpx.New(http.StatusConflict, httpx.CodeReservationAlreadyConfirmed, "reservasi sudah dikonfirmasi sebelumnya")
)

// ErrReservationIDConflict is internal: the service retries with a new id.
var ErrReservationIDConflict = errors.New("konflik id reservasi")
