package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type ExchangeRate struct {
	Bid string `json:"bid"`
}

func getDollarRate(ctx context.Context) (ExchangeRate, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return ExchangeRate{}, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return ExchangeRate{}, err
	}
	defer resp.Body.Close()

	var result map[string]ExchangeRate
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ExchangeRate{}, err
	}

	return result["USDBRL"], nil
}

func saveRateToDB(ctx context.Context, db *sql.DB, rate ExchangeRate) error {
	query := "INSERT INTO rates (bid, created_at) VALUES (?, ?)"
	_, err := db.ExecContext(ctx, query, rate.Bid, time.Now())
	return err
}

func handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Timeout para pegar a cotação
	ctx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()

	rate, err := getDollarRate(ctx)
	if err != nil {
		http.Error(w, "Error getting dollar rate", http.StatusInternalServerError)
		log.Println("Error getting dollar rate:", err)
		return
	}

	db, err := sql.Open("sqlite3", "./data.db")
	if err != nil {
		http.Error(w, "Error opening database", http.StatusInternalServerError)
		log.Println("Error opening database:", err)
		return
	}
	defer db.Close()

	// Timeout para salvar cotação no banco
	dbCtx, dbCancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer dbCancel()

	if err := saveRateToDB(dbCtx, db, rate); err != nil {
		http.Error(w, "Error saving to database", http.StatusInternalServerError)
		log.Println("Error saving to database:", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rate)
}

func main() {
	db, err := sql.Open("sqlite3", "./data.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	query := `CREATE TABLE IF NOT EXISTS rates (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		bid TEXT,
		created_at DATETIME
	)`
	if _, err := db.Exec(query); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/cotacao", handler)
	fmt.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
