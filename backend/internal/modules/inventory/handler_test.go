package inventory

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"indico/backend/internal/platform/httpx"
	"indico/backend/internal/platform/validator"
)

type fakeService struct {
	reserveFn func(ctx context.Context, req ReserveRequest) (*ReserveResponse, error)
	confirmFn func(ctx context.Context, req ConfirmRequest) (*ConfirmResponse, error)
	stockFn   func(ctx context.Context, itemID string) (*StockResponse, error)
}

func (f *fakeService) Reserve(ctx context.Context, req ReserveRequest) (*ReserveResponse, error) {
	return f.reserveFn(ctx, req)
}

func (f *fakeService) Confirm(ctx context.Context, req ConfirmRequest) (*ConfirmResponse, error) {
	return f.confirmFn(ctx, req)
}

func (f *fakeService) GetStock(ctx context.Context, itemID string) (*StockResponse, error) {
	return f.stockFn(ctx, itemID)
}

func newHandlerEngine(t *testing.T, service Service) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	require.NoError(t, validator.Register())

	handler := NewHandler(service)
	engine := gin.New()
	engine.POST("/reserve", handler.Reserve)
	engine.POST("/confirm", handler.Confirm)
	engine.GET("/stock", handler.GetStock)
	return engine
}

func doJSON(engine *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)
	return recorder
}

func decodeError(t *testing.T, recorder *httptest.ResponseRecorder) httpx.ErrorResponse {
	t.Helper()
	var envelope httpx.ErrorResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &envelope))
	return envelope
}

func TestHandlerReserveSuccess(t *testing.T) {
	service := &fakeService{reserveFn: func(_ context.Context, req ReserveRequest) (*ReserveResponse, error) {
		require.Equal(t, "usr_9981", req.UserID)
		require.Equal(t, "item_4021", req.ItemID)
		require.Equal(t, 2, req.Quantity)
		return &ReserveResponse{
			Status:        "success",
			ReservationID: "res_883291",
			ItemID:        "item_4021",
			Quantity:      2,
			ExpiresAt:     "2026-07-20T16:35:00Z",
		}, nil
	}}
	engine := newHandlerEngine(t, service)

	recorder := doJSON(engine, http.MethodPost, "/reserve", `{"user_id":"usr_9981","item_id":"item_4021","quantity":2}`)

	require.Equal(t, http.StatusCreated, recorder.Code)
	var response ReserveResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.Equal(t, "success", response.Status)
	require.Equal(t, "res_883291", response.ReservationID)
	require.Equal(t, "2026-07-20T16:35:00Z", response.ExpiresAt)
}

func TestHandlerReserveMalformedJSON(t *testing.T) {
	engine := newHandlerEngine(t, &fakeService{})

	recorder := doJSON(engine, http.MethodPost, "/reserve", `{bukan json`)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	envelope := decodeError(t, recorder)
	require.Equal(t, httpx.CodeInvalidRequest, envelope.Error.Code)
	require.Len(t, envelope.Error.Details, 1)
	require.Equal(t, "format request tidak valid", envelope.Error.Details[0].Message)
}

func TestHandlerReserveValidationErrors(t *testing.T) {
	engine := newHandlerEngine(t, &fakeService{})

	recorder := doJSON(engine, http.MethodPost, "/reserve", `{"user_id":"x","item_id":"item_4021","quantity":0}`)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	envelope := decodeError(t, recorder)
	require.Equal(t, httpx.CodeInvalidRequest, envelope.Error.Code)
	require.GreaterOrEqual(t, len(envelope.Error.Details), 2)

	messages := map[string]string{}
	for _, detail := range envelope.Error.Details {
		messages[detail.Field+"/"+detail.Rule] = detail.Message
	}
	require.Equal(t, "user_id minimal 3 karakter", messages["user_id/min"])
	require.Equal(t, "quantity harus lebih dari 0", messages["quantity/gt"])
}

func TestHandlerReserveDomainErrors(t *testing.T) {
	cases := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{"insufficient stock", ErrInsufficientStock, http.StatusConflict, httpx.CodeInsufficientStock},
		{"item not found", ErrItemNotFound, http.StatusNotFound, httpx.CodeItemNotFound},
		{"unknown error", errors.New("boom"), http.StatusInternalServerError, httpx.CodeInternalError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := &fakeService{reserveFn: func(context.Context, ReserveRequest) (*ReserveResponse, error) {
				return nil, tc.err
			}}
			engine := newHandlerEngine(t, service)

			recorder := doJSON(engine, http.MethodPost, "/reserve", `{"user_id":"usr_9981","item_id":"item_4021","quantity":1}`)

			require.Equal(t, tc.wantStatus, recorder.Code)
			require.Equal(t, tc.wantCode, decodeError(t, recorder).Error.Code)
		})
	}
}

func TestHandlerConfirmSuccess(t *testing.T) {
	service := &fakeService{confirmFn: func(_ context.Context, req ConfirmRequest) (*ConfirmResponse, error) {
		require.Equal(t, "res_883291", req.ReservationID)
		return &ConfirmResponse{Status: "success", ReservationID: "res_883291", ConfirmedAt: "2026-07-20T16:32:00Z"}, nil
	}}
	engine := newHandlerEngine(t, service)

	recorder := doJSON(engine, http.MethodPost, "/confirm", `{"reservation_id":"res_883291"}`)

	require.Equal(t, http.StatusOK, recorder.Code)
	var response ConfirmResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.Equal(t, "2026-07-20T16:32:00Z", response.ConfirmedAt)
}

func TestHandlerConfirmMalformedJSON(t *testing.T) {
	engine := newHandlerEngine(t, &fakeService{})

	recorder := doJSON(engine, http.MethodPost, "/confirm", `[]`)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.Equal(t, httpx.CodeInvalidRequest, decodeError(t, recorder).Error.Code)
}

func TestHandlerConfirmDomainErrors(t *testing.T) {
	cases := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{"expired", ErrReservationExpired, http.StatusGone, httpx.CodeReservationExpired},
		{"already confirmed", ErrReservationAlreadyConfirmed, http.StatusConflict, httpx.CodeReservationAlreadyConfirmed},
		{"not found", ErrReservationNotFound, http.StatusNotFound, httpx.CodeReservationNotFound},
		{"unknown error", errors.New("boom"), http.StatusInternalServerError, httpx.CodeInternalError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := &fakeService{confirmFn: func(context.Context, ConfirmRequest) (*ConfirmResponse, error) {
				return nil, tc.err
			}}
			engine := newHandlerEngine(t, service)

			recorder := doJSON(engine, http.MethodPost, "/confirm", `{"reservation_id":"res_883291"}`)

			require.Equal(t, tc.wantStatus, recorder.Code)
			require.Equal(t, tc.wantCode, decodeError(t, recorder).Error.Code)
		})
	}
}

func TestHandlerGetStockSuccess(t *testing.T) {
	service := &fakeService{stockFn: func(_ context.Context, itemID string) (*StockResponse, error) {
		require.Equal(t, "item_4021", itemID)
		return &StockResponse{ItemID: "item_4021", TotalStock: 100, ReservedStock: 15, AvailableStock: 85}, nil
	}}
	engine := newHandlerEngine(t, service)

	recorder := doJSON(engine, http.MethodGet, "/stock?item_id=item_4021", "")

	require.Equal(t, http.StatusOK, recorder.Code)
	var response StockResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.Equal(t, 100, response.TotalStock)
	require.Equal(t, 15, response.ReservedStock)
	require.Equal(t, 85, response.AvailableStock)
}

func TestHandlerGetStockMissingParam(t *testing.T) {
	engine := newHandlerEngine(t, &fakeService{})

	recorder := doJSON(engine, http.MethodGet, "/stock", "")

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	envelope := decodeError(t, recorder)
	require.Equal(t, httpx.CodeInvalidRequest, envelope.Error.Code)
	require.Equal(t, "item_id wajib diisi", envelope.Error.Details[0].Message)
}

func TestHandlerGetStockNotFound(t *testing.T) {
	service := &fakeService{stockFn: func(context.Context, string) (*StockResponse, error) {
		return nil, ErrItemNotFound
	}}
	engine := newHandlerEngine(t, service)

	recorder := doJSON(engine, http.MethodGet, "/stock?item_id=item_9999", "")

	require.Equal(t, http.StatusNotFound, recorder.Code)
	require.Equal(t, httpx.CodeItemNotFound, decodeError(t, recorder).Error.Code)
}
