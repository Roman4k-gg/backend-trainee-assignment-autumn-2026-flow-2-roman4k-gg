package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type WebhookPayload struct {
	OrderID      string `json:"order_id"`
	RestaurantID string `json:"restaurant_id"`
	TotalAmount  int    `json:"total_amount"`
}

var kitchenAPIURL string

func main() {
	kitchenAPIURL = os.Getenv("KITCHEN_API_URL")
	if kitchenAPIURL == "" {
		kitchenAPIURL = "http://localhost:8080"
	}

	http.HandleFunc("/webhook", handleWebhook)

	log.Println("Starting Mock Restaurant Service on :8081")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatalf("Mock Restaurant failed: %v", err)
	}
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload WebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Printf("Failed to decode webhook payload: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	log.Printf("[RESTAURANT] Received new order: %s, Total: %d", payload.OrderID, payload.TotalAmount)


	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))


	go processOrder(payload.OrderID)
}

func processOrder(orderID string) {

	time.Sleep(2 * time.Second)
	updateOrderStatus(orderID, "ACCEPTED")


	time.Sleep(5 * time.Second)
	updateOrderStatus(orderID, "PREPARING")


	time.Sleep(3 * time.Second)
	updateOrderStatus(orderID, "DELIVERING")


	time.Sleep(2 * time.Second)
	updateOrderStatus(orderID, "DELIVERED")
}

func updateOrderStatus(orderID, status string) {
	url := fmt.Sprintf("%s/api/v1/orders/%s/status", kitchenAPIURL, orderID)
	
	reqBody, _ := json.Marshal(map[string]string{
		"status": status,
	})

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(reqBody))
	if err != nil {
		log.Printf("[RESTAURANT ERROR] Failed to create request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[RESTAURANT ERROR] Failed to send status %s for order %s: %v", status, orderID, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		log.Printf("[RESTAURANT ERROR] API returned status %d when setting status %s", resp.StatusCode, status)
	} else {
		log.Printf("[RESTAURANT] Order %s status updated to: %s", orderID, status)
	}
}
