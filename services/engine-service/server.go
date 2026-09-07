package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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

		if NewOrder.Stock_name == "" {
			http.Error(w, "stock name cannot be empty", http.StatusBadRequest)
			return
		}

		if NewOrder.Amount <= 0 || NewOrder.Price <= 0 {
			http.Error(w, "amount and price must be greater than 0", http.StatusBadRequest)
			return
		}

		if NewOrder.Order_type != "buy" && NewOrder.Order_type != "sell" {
			http.Error(w, "order_type must be 'buy' or 'sell'", http.StatusBadRequest)
			return
		}

		if err := SaveOrderToDB(&NewOrder); err != nil {
			http.Error(w, "failed to save order:"+err.Error(), http.StatusInternalServerError)
			return
		}

		engine.AddOrder(NewOrder)

		fmt.Printf("[API]: New %s order received via REST for %s at %.2f\n", NewOrder.Order_type, NewOrder.Stock_name, NewOrder.Price)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Order successfully sent to stock matching engine!"))
	}
}

func StartServer(engine *Engine) {

	mux := http.NewServeMux()
	mux.HandleFunc("/add-order", HandleAddOrder(engine))

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("REST API Gateway is running on port %s...\n", port)

	err := http.ListenAndServe(":"+port, mux)

	if err != nil {
		fmt.Println("[SERVER ERROR]: Failed to start API Gateway:", err)
	}
}
