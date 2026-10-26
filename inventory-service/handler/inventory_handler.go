package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"example.com/go-commerce-platform/inventory-service/service"
)

type InventoryHandler struct {
	Service *service.InventoryService
}

func NewInventoryHandler(s *service.InventoryService) *InventoryHandler {
	return &InventoryHandler{Service: s}
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func (h *InventoryHandler) Products(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var req struct {
			Name          string `json:"name"`
			SKU           string `json:"sku"`
			StockQuantity int    `json:"stock_quantity"`
		}

		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
		decoder.DisallowUnknownFields()

		if err := decoder.Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON request")
			return
		}

		product, err := h.Service.Create(
			r.Context(), req.Name, req.SKU, req.StockQuantity,
		)
		if err != nil {
			h.handleError(w, err)
			return
		}

		writeJSON(w, http.StatusCreated, product)

	case http.MethodGet:
		products, err := h.Service.List(r.Context())
		if err != nil {
			h.handleError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, products)

	default:
		w.Header().Set("Allow", "GET, POST")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *InventoryHandler) ProductByID(w http.ResponseWriter, r *http.Request) {
	idText := strings.TrimPrefix(r.URL.Path, "/products/")
	id, err := strconv.ParseInt(idText, 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid product ID")
		return
	}

	switch r.Method {
	case http.MethodGet:
		product, err := h.Service.GetByID(r.Context(), id)
		if err != nil {
			h.handleError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, product)

	case http.MethodPatch:
		var req struct {
			Name string `json:"name"`
		}

		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
		decoder.DisallowUnknownFields()

		if err := decoder.Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON request")
			return
		}

		product, err := h.Service.Update(r.Context(), id, req.Name)
		if err != nil {
			h.handleError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, product)

	default:
		w.Header().Set("Allow", "GET, PATCH")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *InventoryHandler) Stock(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	idText := strings.TrimPrefix(r.URL.Path, "/products/")
	idText = strings.TrimSuffix(idText, "/stock")
	id, err := strconv.ParseInt(idText, 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid product ID")
		return
	}

	var req struct {
		Delta int `json:"delta"`
	}

	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON request")
		return
	}

	product, err := h.Service.ChangeStock(r.Context(), id, req.Delta)
	if err != nil {
		h.handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, product)
}

func (h *InventoryHandler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrProductNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrInsufficientStock):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, service.ErrDuplicateSKU):
		writeError(w, http.StatusConflict, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
