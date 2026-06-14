package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/giovaniif/sharding-test/models"
)

type OrderHandler struct {
	db *sql.DB
}

func NewOrderHandler(db *sql.DB) *OrderHandler {
	return &OrderHandler{db: db}
}

func (h *OrderHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.QueryContext(r.Context(),
		`SELECT id, product, quantity, price, created_at FROM orders ORDER BY id`)
	if err != nil {
		http.Error(w, "failed to fetch orders", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	orders := []models.Order{}
	for rows.Next() {
		var o models.Order
		if err := rows.Scan(&o.ID, &o.Product, &o.Quantity, &o.Price, &o.CreatedAt); err != nil {
			http.Error(w, "failed to scan order", http.StatusInternalServerError)
			return
		}
		orders = append(orders, o)
	}

	writeJSON(w, http.StatusOK, orders)
}

func (h *OrderHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var o models.Order
	err = h.db.QueryRowContext(r.Context(),
		`SELECT id, product, quantity, price, created_at FROM orders WHERE id = $1`, id).
		Scan(&o.ID, &o.Product, &o.Quantity, &o.Price, &o.CreatedAt)
	if err == sql.ErrNoRows {
		http.Error(w, "order not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "failed to fetch order", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, o)
}

func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Product == "" || req.Quantity <= 0 || req.Price <= 0 {
		http.Error(w, "product, quantity, and price are required and must be positive", http.StatusBadRequest)
		return
	}

	var o models.Order
	err := h.db.QueryRowContext(r.Context(),
		`INSERT INTO orders (product, quantity, price) VALUES ($1, $2, $3)
		 RETURNING id, product, quantity, price, created_at`,
		req.Product, req.Quantity, req.Price).
		Scan(&o.ID, &o.Product, &o.Quantity, &o.Price, &o.CreatedAt)
	if err != nil {
		http.Error(w, "failed to create order", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, o)
}

func (h *OrderHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	result, err := h.db.ExecContext(r.Context(),
		`DELETE FROM orders WHERE id = $1`, id)
	if err != nil {
		http.Error(w, "failed to delete order", http.StatusInternalServerError)
		return
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		http.Error(w, "order not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
