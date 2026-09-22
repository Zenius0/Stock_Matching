package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var stockNameRegex = regexp.MustCompile(`^[A-Z]{1,5}$`)

const (
	maxOrderAmount = 1_000_000.0
	maxOrderPrice  = 1_000_000.0
)

func HandleAddOrder(engine *Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}

		var NewOrder Order

		err := json.NewDecoder(r.Body).Decode(&NewOrder)

		if err != nil {
			http.Error(w, "Invalid JSON format: "+err.Error(), http.StatusBadRequest)
			return
		}

		if err := validateOrder(&NewOrder); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := SaveOrderToDB(&NewOrder); err != nil {
			slog.Error("failed to save order to database", "error", err)
			http.Error(w, "failed to save order:"+err.Error(), http.StatusInternalServerError)
			return
		}

		engine.AddOrder(NewOrder)

		slog.Info("[API]: New order received via REST",
			"order_id", NewOrder.Id,
			"order_type", NewOrder.Order_type,
			"stock_name", NewOrder.Stock_name,
			"price", NewOrder.Price,
		)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(NewOrder)
	}
}

func validateOrder(o *Order) error {
	if o.UserId <= 0 {
		return fmt.Errorf("user_id must be a positive integer")
	}

	o.Stock_name = strings.ToUpper(strings.TrimSpace(o.Stock_name))
	if !stockNameRegex.MatchString(o.Stock_name) {
		return fmt.Errorf("stock_name must be 1-5 uppercase letters (e.g. AAPL, TSLA)")
	}

	if o.Amount <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}

	if o.Amount > maxOrderAmount {
		return fmt.Errorf("amount exceeds maximum allowed value (%v)", maxOrderAmount)
	}

	if o.Price <= 0 {
		return fmt.Errorf("price must be greater than 0")
	}

	if o.Price > maxOrderPrice {
		return fmt.Errorf("price exceeds maximum allowed value (%v)", maxOrderPrice)
	}

	o.Order_type = strings.ToLower(strings.TrimSpace(o.Order_type))
	if o.Order_type != "buy" && o.Order_type != "sell" {
		return fmt.Errorf("order_type must be 'buy' or 'sell'")
	}
	return nil
}

func HandleGetOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := GetAllOrders()
	if err != nil {
		slog.Error("failed to fetch orders", "error", err)
		http.Error(w, "Failed to fetch orders", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func HandleCancelOrder(engine *Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid order id", http.StatusBadRequest)
			return
		}
		if err := CancelOrderInDB(id); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "Order not found", http.StatusNotFound)
				return
			}
			slog.Error("failed to cancel order in db", "error", err, "order_id", id)
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}

		removed := engine.CancelOrder(id)

		slog.Info("order cancelled", "order_id", id, "removed from engine", removed)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Order cancelled successfully",
		})
	}
}

func HandleGetOrderBook(engine *Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stockName := r.PathValue("stock")

		buys, sells := engine.GetOrderBook(stockName)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"stock_name":  stockName,
			"buy_orders":  buys,
			"sell_orders": sells,
		})
	}
}

func StartServer(engine *Engine) {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /add-order", HandleAddOrder(engine))
	mux.HandleFunc("GET /orders", HandleGetOrders)
	mux.HandleFunc("DELETE /orders/{id}", HandleCancelOrder(engine))
	mux.HandleFunc("GET /orderbook/{stock}", HandleGetOrderBook(engine))

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	slog.Info("REST API Gateway is running on port", "port", port)

	err := http.ListenAndServe(":"+port, mux)

	if err != nil {
		slog.Error("Failed to start API Gateway", "error", err)
	}
}
