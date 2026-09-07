package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var db *sql.DB

func InitDB() error {
	if err := godotenv.Load(); err != nil {
		fmt.Println("[WARNING]: .env file not found, relying on system environment variables")
	}

	connStr := fmt.Sprintf(
		"user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
	)

	var err error

	db, err = sql.Open("postgres", connStr)

	if err != nil {
		return fmt.Errorf("[DB ERROR]: Failed to open database connection: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	err = db.Ping()

	if err != nil {
		return fmt.Errorf("[ERROR]: cannot reach database! Is postgres running? %w", err)
	}

	fmt.Println("Successfully connected to PostgreSQL!")

	createTableSql := `
		CREATE TABLE IF NOT EXISTS orders (
			id SERIAL PRIMARY KEY,
			userid INT NOT NULL,
			stock_name VARCHAR(10) NOT NULL,
			amount NUMERIC NOT NULL,
			price NUMERIC NOT NULL,
			order_type VARCHAR(5) NOT NULL,
			status VARCHAR(10) NOT NULL DEFAULT 'pending', -- 3) pending, filled, cancelled, partial,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`

	if _, err = db.Exec(createTableSql); err != nil {
		return fmt.Errorf("[ERROR]: failed to create index on 'stock_name': %w", err)
	}

	fmt.Println("'orders' table is ready!")

	createIndexSql := `CREATE INDEX IF NOT EXISTS idx_orders_stock_name ON orders(stock_name);`

	if _, err = db.Exec(createIndexSql); err != nil {
		return fmt.Errorf("[ERROR]: Failed to create index on stock_name: %w", err)
	}

	fmt.Println("Index on stock_name is ready!")

	return nil

}

func SaveOrderToDB(o *Order) error {
	query := `INSERT INTO orders (userid, stock_name, amount, price, order_type) 
	          VALUES ($1, $2, $3, $4, $5) RETURNING id`

	err := db.QueryRow(query, o.UserId, o.Stock_name, o.Amount, o.Price, o.Order_type).Scan(&o.Id)
	return err
}

func UpdateOrderStatus(id int, status string) error {
	query := `UPDATE orders SET status = $1 WHERE id = $2`

	_, err := db.Exec(query, status, id)
	return err
}
