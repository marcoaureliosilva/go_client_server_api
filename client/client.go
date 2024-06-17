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

const (
	serverUrl      = "http://server:8080/cotacao"
	clientTimeout  = 300 * time.Millisecond
	outputFilePath = "/app/cotacao.txt"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), clientTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", serverUrl, nil)
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to get response: %v", err)
	}
	defer resp.Body.Close()

	var data map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		log.Fatalf("Failed to decode response: %v", err)
	}

	bid, ok := data["bid"].(string)
	if !ok {
		log.Fatalf("Failed to get bid value from response")
	}

	content := fmt.Sprintf("Dólar: %s", bid)
	err = ioutil.WriteFile(outputFilePath, []byte(content), 0644)
	if err != nil {
		log.Fatalf("Failed to write to file: %v", err)
	}

	fmt.Println("Cotação armazenada com sucesso:", content)
}
