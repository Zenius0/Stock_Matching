package main

import "testing"

func TestAddOrder_BuyOrderSorting(t *testing.T) {
	e := &Engine{
		buying_orders:  make([]Order, 0),
		selling_orders: make([]Order, 0),
	}

	e.AddOrder(Order{Id: 1, Stock_name: "AAPL", Price: 100, Order_type: "buy"})
	e.AddOrder(Order{Id: 2, Stock_name: "AAPL", Price: 150, Order_type: "buy"})
	e.AddOrder(Order{Id: 3, Stock_name: "AAPL", Price: 120, Order_type: "buy"})

	if len(e.buying_orders) != 3 {
		t.Fatalf("expected 3 buy orders, got %d", len(e.buying_orders))
	}

	if e.buying_orders[0].Price != 150 {
		t.Errorf("expected highest price (150) first, got %v", e.buying_orders[0].Price)
	}

	if e.buying_orders[1].Price != 120 {
		t.Errorf("expected second price (120), got %v", e.buying_orders[1].Price)
	}

	if e.buying_orders[2].Price != 100 {
		t.Errorf("expected lowest price (100) last, got %v", e.buying_orders[2].Price)
	}
}

func TestAddOrder_SellOrderSorting(t *testing.T) {
	e := Engine{
		buying_orders:  make([]Order, 0),
		selling_orders: make([]Order, 0),
	}

	e.AddOrder(Order{Id: 1, Stock_name: "AAPL", Price: 100, Order_type: "sell"})
	e.AddOrder(Order{Id: 2, Stock_name: "AAPL", Price: 80, Order_type: "sell"})
	e.AddOrder(Order{Id: 3, Stock_name: "AAPL", Price: 90, Order_type: "sell"})

	if len(e.selling_orders) != 3 {
		t.Fatalf("expected 3 selling orders, got %d", len(e.selling_orders))
	}

	if e.selling_orders[0].Price != 80 {
		t.Errorf("expected lowest price (80) first, got %v", e.selling_orders[0].Price)
	}
}

func TestCancelOrder(t *testing.T) {
	e := &Engine{
		buying_orders:  make([]Order, 0),
		selling_orders: make([]Order, 0),
	}

	e.AddOrder(Order{Id: 1, Stock_name: "AAPL", Price: 100, Order_type: "buy"})
	e.AddOrder(Order{Id: 2, Stock_name: "AAPL", Price: 150, Order_type: "buy"})

	removed := e.CancelOrder(1)

	if !removed {
		t.Fatal("expected CancelOrder to return true for existing order")
	}

	if len(e.buying_orders) != 1 {
		t.Errorf("expected 1 remaining buy order, got %d", len(e.buying_orders))
	}

	if e.buying_orders[0].Id != 2 {
		t.Errorf("expected remaining order to have Id 2, got %d", e.buying_orders[0].Id)
	}

	removed = e.CancelOrder(999)

	if removed {
		t.Error("expected CancelOrder to return false for non-existing order")
	}
}
