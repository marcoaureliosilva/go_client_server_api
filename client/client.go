package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type ExchangeRate struct {
	Bid string `json:"bid"`
}

func main() {
	// Criar um contexto com timeout de 300ms para a requisição HTTP
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to perform request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Server returned non-200 status: %v", resp.Status)
	}

	var rate ExchangeRate
	if err := json.NewDecoder(resp.Body).Decode(&rate); err != nil {
		log.Fatalf("Failed to decode response: %v", err)
	}

	content := fmt.Sprintf("Dólar: %s", rate.Bid)
	if err := ioutil.WriteFile("cotacao.txt", []byte(content), 0644); err != nil {
		log.Fatalf("Failed to write to file: %v", err)
	}

	fmt.Println("Cotação salva em 'cotacao.txt'")
}
