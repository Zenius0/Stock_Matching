package main

import "testing"

func TestAtttemptMatch_Success(t *testing.T) {
	var updateIDs []int

	e := &Engine{
		buying_orders:  []Order{{Id: 1, Stock_name: "AAPL", Price: 150, Order_type: "buy"}},
		selling_orders: []Order{{Id: 2, Stock_name: "APPL", Price: 150, Order_type: "sell"}},
		updateStatus: func(id int, status string) error {
			updateIDs = append(updateIDs, id)
			return nil
		},
	}

	matched := e.attemptMatch()

	if !matched {
		t.Fatal("expected attemptMatch to return true")
	}
	if len(e.buying_orders) != 0 {
		t.Errorf("expected buying_orders to be empty, got %d", len(e.buying_orders))
	}
	if len(e.selling_orders) != 0 {
		t.Errorf("expected selling_orders to be empty, got %d", len(e.selling_orders))
	}
	if len(updateIDs) != 2 {
		t.Fatalf("expected 2 status updates, got %d", len(updateIDs))
	}
}

func TestAttemptMatch_NoMatch(t *testing.T) {
	e := &Engine{
		buying_orders:  []Order{{Id: 1, Stock_name: "TSLA", Price: 100, Order_type: "buy"}},
		selling_orders: []Order{{Id: 2, Stock_name: "TSLA", Price: 200, Order_type: "sell"}},
		updateStatus: func(id int, status string) error {
			t.Error("updateStatus should not be called when there is no match")
			return nil
		},
	}

	matched := e.attemptMatch()

	if matched {
		t.Fatal("expected attemptMatch to return false")
	}
	if len(e.buying_orders) != 1 || len(e.selling_orders) != 1 {
		t.Error("expected orders to remain in queues when no match occurs")
	}
}

func TestAttemptMatch_EmptyQueues(t *testing.T) {
	e := &Engine{
		buying_orders:  make([]Order, 0),
		selling_orders: make([]Order, 0),
	}

	matched := e.attemptMatch()

	if matched {
		t.Fatal("expected attemptMatch to return false when queues are empty")
	}
}
