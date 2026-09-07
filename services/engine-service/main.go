package main

import (
	"fmt"
	"log"
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
}

type Engine struct {
	buying_orders  []Order
	selling_orders []Order
	mu             sync.Mutex
}

func (e *Engine) Match() {
	wasWaiting := false

	for {
		e.mu.Lock()
		if len(e.buying_orders) == 0 || len(e.selling_orders) == 0 {
			e.mu.Unlock()
			time.Sleep(10 * time.Millisecond)
			continue
		}
		biggest_buying := e.buying_orders[0]
		smallest_selling := e.selling_orders[0]

		if biggest_buying.Price >= smallest_selling.Price {
			fmt.Printf("Matched successfully. Matched orders are: %v and %v", e.buying_orders[0], e.selling_orders[0])

			if err := UpdateOrderStatus(biggest_buying.Id, "filled"); err != nil {
				fmt.Println("[DB ERROR]: Failed to update buy order status", err)
			}

			if err := UpdateOrderStatus(smallest_selling.Id, "filled"); err != nil {
				fmt.Println("[DB ERROR]: Failed to update sell order status", err)
			}

			e.buying_orders = e.buying_orders[1:]
			e.selling_orders = e.selling_orders[1:]
			wasWaiting = false
		} else {
			if !wasWaiting {
				fmt.Println("no match yet, waiting for price convergence...")
				wasWaiting = true
			}

		}
		e.mu.Unlock()
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

func main() {

	if err := InitDB(); err != nil {
		log.Fatal(err)
	}

	m := &Engine{
		buying_orders:  make([]Order, 0),
		selling_orders: make([]Order, 0),
	}

	go m.Match()

	StartServer(m)
}
