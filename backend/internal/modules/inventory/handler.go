package inventory

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"indico/backend/internal/platform/httpx"
	"indico/backend/internal/platform/validator"
)

// Handler holds the HTTP layer of the inventory module.
type Handler struct {
	service Service
}

// NewHandler builds the inventory HTTP handler.
func NewHandler(service Service) *Handler { return &Handler{service: service} }

// Reserve godoc
//
//	@Summary		Reservasi stok item
//	@Description	Mengunci stok secara atomik untuk seorang user. Reservasi otomatis kedaluwarsa dalam 5 menit jika tidak dikonfirmasi.
//	@Tags			inventory
//	@Accept			json
//	@Produce		json
//	@Param			request	body		ReserveRequest	true	"Payload reservasi"
//	@Success		201		{object}	ReserveResponse
//	@Failure		400		{object}	httpx.ErrorResponse	"Input tidak valid"
//	@Failure		404		{object}	httpx.ErrorResponse	"Item tidak ditemukan"
//	@Failure		409		{object}	httpx.ErrorResponse	"Stok tidak mencukupi"
//	@Failure		500		{object}	httpx.ErrorResponse	"Kesalahan internal"
//	@Router			/api/v1/inventory/reserve [post]
func (h *Handler) Reserve(c *gin.Context) {
	var request ReserveRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpx.Error(c, httpx.ErrInvalidRequest.WithDetails(validator.Translate(err)...))
		return
	}

	response, err := h.service.Reserve(c.Request.Context(), request)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.Success(c, http.StatusCreated, response)
}

// Confirm godoc
//
//	@Summary		Konfirmasi reservasi
//	@Description	Mengubah reservasi aktif menjadi permanen dan mengurangi stok fisik. Reservasi yang sudah kedaluwarsa atau sudah dikonfirmasi akan ditolak.
//	@Tags			inventory
//	@Accept			json
//	@Produce		json
//	@Param			request	body		ConfirmRequest	true	"Payload konfirmasi"
//	@Success		200		{object}	ConfirmResponse
//	@Failure		400		{object}	httpx.ErrorResponse	"Input tidak valid"
//	@Failure		404		{object}	httpx.ErrorResponse	"Reservasi tidak ditemukan"
//	@Failure		409		{object}	httpx.ErrorResponse	"Reservasi sudah dikonfirmasi"
//	@Failure		410		{object}	httpx.ErrorResponse	"Reservasi sudah kedaluwarsa"
//	@Failure		500		{object}	httpx.ErrorResponse	"Kesalahan internal"
//	@Router			/api/v1/inventory/confirm [post]
func (h *Handler) Confirm(c *gin.Context) {
	var request ConfirmRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httpx.Error(c, httpx.ErrInvalidRequest.WithDetails(validator.Translate(err)...))
		return
	}

	response, err := h.service.Confirm(c.Request.Context(), request)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.Success(c, http.StatusOK, response)
}

// GetStock godoc
//
//	@Summary		Status stok item
//	@Description	Menampilkan rincian stok real-time: total, yang sedang direservasi, dan yang tersedia.
//	@Tags			inventory
//	@Produce		json
//	@Param			item_id	query		string	true	"ID item"	example(item_4021)
//	@Success		200		{object}	StockResponse
//	@Failure		400		{object}	httpx.ErrorResponse	"Parameter tidak valid"
//	@Failure		404		{object}	httpx.ErrorResponse	"Item tidak ditemukan"
//	@Failure		500		{object}	httpx.ErrorResponse	"Kesalahan internal"
//	@Router			/api/v1/inventory/stock [get]
func (h *Handler) GetStock(c *gin.Context) {
	var query StockQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		httpx.Error(c, httpx.ErrInvalidRequest.WithDetails(validator.Translate(err)...))
		return
	}

	response, err := h.service.GetStock(c.Request.Context(), query.ItemID)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	httpx.Success(c, http.StatusOK, response)
}
