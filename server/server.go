package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	apiUrl        = "https://economia.awesomeapi.com.br/json/last/USD-BRL"
	dbTimeout     = 10 * time.Second
	apiTimeout    = 200 * time.Second
	serverAddress = ":8080"
)

type Cotacao struct {
	Code string `json:"code"`
	Bid  string `json:"bid"`
}

type ApiResponse struct {
	USD_BRL Cotacao `json:"USDBRL"`
}

func fetchCotacao(ctx context.Context) (Cotacao, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", apiUrl, nil)
	if err != nil {
		return Cotacao{}, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return Cotacao{}, err
	}
	defer resp.Body.Close()

	var apiResp ApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return Cotacao{}, err
	}

	return apiResp.USD_BRL, nil
}

func saveToDB(ctx context.Context, db *sql.DB, cotacao Cotacao) error {
	query := "INSERT INTO cotacoes (code, bid) VALUES (?, ?)"
	stmt, err := db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, cotacao.Code, cotacao.Bid)
	return err
}

func cotacaoHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), apiTimeout)
	defer cancel()

	cotacao, err := fetchCotacao(ctx)
	if err != nil {
		http.Error(w, "Failed to fetch cotacao", http.StatusInternalServerError)
		log.Printf("Failed to fetch cotacao: %v", err)
		return
	}

	db, err := sql.Open("sqlite3", "./cotacoes.db")
	if err != nil {
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		log.Printf("Failed to connect to database: %v", err)
		return
	}
	defer db.Close()

	dbCtx, dbCancel := context.WithTimeout(r.Context(), dbTimeout)
	defer dbCancel()

	if err := saveToDB(dbCtx, db, cotacao); err != nil {
		http.Error(w, "Failed to save cotacao", http.StatusInternalServerError)
		log.Printf("Failed to save cotacao: %v", err)
		return
	}

	json.NewEncoder(w).Encode(cotacao)
}

func main() {
	db, err := sql.Open("sqlite3", "./cotacoes.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	createTableQuery := `CREATE TABLE IF NOT EXISTS cotacoes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		code TEXT,
		bid TEXT
	)`
	_, err = db.Exec(createTableQuery)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	http.HandleFunc("/cotacao", cotacaoHandler)
	log.Fatal(http.ListenAndServe(serverAddress, nil))
}
