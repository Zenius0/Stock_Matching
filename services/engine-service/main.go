package main

import (
	"log/slog"
	"os"
	"sort"
	"sync"
	"time"
)

type Order struct {
	Id         int     `json:"id,omitempty"`
	UserId     int     `json:"user_id"`
	Stock_name string  `json:"stock_name"`
	Amount     float64 `json:"amount"`
	Price      float64 `json:"price"`
	Order_type string  `json:"order_type"`
	Status     string  `json:"status,omitempty"`
}

type Engine struct {
	buying_orders  []Order
	selling_orders []Order
	mu             sync.Mutex
	updateStatus   func(id int, status string) error
	noMatchLogged  bool
}

func InitLogger() {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

func (e *Engine) attemptMatch() bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	if len(e.buying_orders) == 0 || len(e.selling_orders) == 0 {
		return false
	}

	biggest_buying := e.buying_orders[0]
	smallest_selling := e.selling_orders[0]

	if biggest_buying.Price < smallest_selling.Price {
		if !e.noMatchLogged {
			slog.Debug("no match yet, waiting for price convergence",
				"best_buy_price", biggest_buying.Price,
				"best_sell_price", smallest_selling.Price,
			)
			e.noMatchLogged = true
		}
		return false
	}

	slog.Info("orders matched",
		"buy_order_id", biggest_buying.Id,
		"sell_order_id", smallest_selling.Id,
		"stock_name", biggest_buying.Stock_name,
		"price", smallest_selling.Price,
	)

	if e.updateStatus != nil {
		if err := e.updateStatus(biggest_buying.Id, "filled"); err != nil {
			slog.Error("failed to update buy order status", "error", err, "order_id", biggest_buying.Id)
		}
		if err := e.updateStatus(smallest_selling.Id, "filled"); err != nil {
			slog.Error("failed to update sell order status", "error", err, "order_id", smallest_selling.Id)
		}
	}

	e.buying_orders = e.buying_orders[1:]
	e.selling_orders = e.selling_orders[1:]
	e.noMatchLogged = false

	return true
}

func (e *Engine) Match() {
	for {
		e.attemptMatch()
		time.Sleep(10 * time.Millisecond)
	}
}

func (e *Engine) AddOrder(o Order) {
	e.mu.Lock()
	if o.Order_type == "buy" {
		e.buying_orders = append(e.buying_orders, o)
		sort.Slice(e.buying_orders, func(i int, j int) bool {
			return e.buying_orders[i].Price > e.buying_orders[j].Price
		})
	} else {
		e.selling_orders = append(e.selling_orders, o)
		sort.Slice(e.selling_orders, func(i, j int) bool {
			return e.selling_orders[i].Price < e.selling_orders[j].Price
		})
	}
	e.mu.Unlock()
}

func (e *Engine) CancelOrder(id int) bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	for i, o := range e.buying_orders {
		if o.Id == id {
			e.buying_orders = append(e.buying_orders[:i], e.buying_orders[i+1:]...)
			return true
		}

	}

	for i, o := range e.selling_orders {
		if o.Id == id {
			e.selling_orders = append(e.selling_orders[:i], e.selling_orders[i+1:]...)
			return true
		}

	}
	return false
}

func (e *Engine) GetOrderBook(stockName string) (buys []Order, sells []Order) {
	e.mu.Lock()
	defer e.mu.Unlock()

	for _, o := range e.buying_orders {
		if o.Stock_name == stockName {
			buys = append(buys, o)
		}
	}
	for _, o := range e.selling_orders {
		if o.Stock_name == stockName {
			sells = append(sells, o)
		}
	}
	return buys, sells
}

func (e *Engine) LoadPendingErrors() error {
	orders, err := GetAllOrders()
	if err != nil {
		return err
	}

	for _, o := range orders {
		if o.Status == "pending" {
			e.AddOrder(o)
		}
	}

	return nil
}

func main() {
	InitLogger()

	if err := InitDB(); err != nil {
		slog.Error("failed to initialize database", "error", err)
	}

	m := &Engine{
		buying_orders:  make([]Order, 0),
		selling_orders: make([]Order, 0),
		updateStatus:   UpdateOrderStatus,
	}

	if err := m.LoadPendingErrors(); err != nil {
		slog.Error("failed to load pending errors from database", "error", err)
		os.Exit(1)
	}
	slog.Info("pending orders loaded into engine!")

	go m.Match()

	StartServer(m)
}
