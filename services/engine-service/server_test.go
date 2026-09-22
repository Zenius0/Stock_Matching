package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestValidateOrder(t *testing.T) {
	tests := []struct {
		name      string
		order     Order
		expectErr bool
	}{
		{
			name:      "valid buy order",
			order:     Order{UserId: 1, Stock_name: "AAPL", Amount: 10, Price: 150, Order_type: "buy"},
			expectErr: false,
		},
		{
			name:      "negative user id",
			order:     Order{UserId: -1, Stock_name: "AAPL", Amount: 10, Price: 150, Order_type: "buy"},
			expectErr: true,
		},
		{
			name:      "zero user id",
			order:     Order{UserId: 0, Stock_name: "AAPL", Amount: 10, Price: 150, Order_type: "buy"},
			expectErr: true,
		},
		{
			name:      "empty stock name",
			order:     Order{UserId: 1, Stock_name: "", Amount: 10, Price: 150, Order_type: "buy"},
			expectErr: true,
		},
		{
			name:      "invalid stock name format",
			order:     Order{UserId: 1, Stock_name: "AAPL123", Amount: 10, Price: 150, Order_type: "buy"},
			expectErr: true,
		},
		{
			name:      "lowercase stock name gets normalized, should pass",
			order:     Order{UserId: 1, Stock_name: "aapl", Amount: 10, Price: 150, Order_type: "buy"},
			expectErr: false,
		},
		{
			name:      "zero amount",
			order:     Order{UserId: 1, Stock_name: "AAPL", Amount: 0, Price: 150, Order_type: "buy"},
			expectErr: true,
		},
		{
			name:      "negative price",
			order:     Order{UserId: 1, Stock_name: "AAPL", Amount: 10, Price: -50, Order_type: "buy"},
			expectErr: true,
		},
		{
			name:      "amount exceeds max",
			order:     Order{UserId: 1, Stock_name: "AAPL", Amount: 99_999_999, Price: 150, Order_type: "buy"},
			expectErr: true,
		},
		{
			name:      "invalid order type",
			order:     Order{UserId: 1, Stock_name: "AAPL", Amount: 10, Price: 150, Order_type: "hold"},
			expectErr: true,
		},
		{
			name:      "uppercase BUY gets normalized, should pass",
			order:     Order{UserId: 1, Stock_name: "AAPL", Amount: 10, Price: 150, Order_type: "BUY"},
			expectErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateOrder(&tc.order)

			if tc.expectErr && err == nil {
				t.Errorf("expected an error but got nil")
			}
			if !tc.expectErr && err != nil {
				t.Errorf("expected no error but got: %v", err)
			}
		})
	}
}

func TestHandleAddOrder_WrongMethod(t *testing.T) {
	engine := &Engine{
		buying_orders:  make([]Order, 0),
		selling_orders: make([]Order, 0),
	}

	req := httptest.NewRequest(http.MethodGet, "/add-order", nil)
	rec := httptest.NewRecorder()

	handler := HandleAddOrder(engine)
	handler(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
	}
}

func TestHandleAddOrder_InvalidJSON(t *testing.T) {
	engine := &Engine{
		buying_orders:  make([]Order, 0),
		selling_orders: make([]Order, 0),
	}

	body := bytes.NewBufferString(`{"invalid json`)
	req := httptest.NewRequest(http.MethodPost, "/add-order", body)
	rec := httptest.NewRecorder()

	handler := HandleAddOrder(engine)
	handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %v", http.StatusBadRequest, rec.Code)
	}
}

func TestHandleAddOrder_ValidationFailure(t *testing.T) {
	engine := &Engine{
		buying_orders:  make([]Order, 0),
		selling_orders: make([]Order, 0),
	}

	invalidOrder := Order{UserId: -1, Stock_name: "AAPL", Amount: 10, Price: 150, Order_type: "buy"}
	jsonBody, _ := json.Marshal(invalidOrder)

	req := httptest.NewRequest(http.MethodPost, "/add-order", bytes.NewBuffer(jsonBody))
	rec := httptest.NewRecorder()

	handler := HandleAddOrder(engine)
	handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	if len(engine.buying_orders) != 0 {
		t.Errorf("expecting no orders added to engine, got %d", len(engine.buying_orders))
	}
}
