package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"avito-kitchen/internal/repository"
	"github.com/google/uuid"
)

type KitchenService struct {
	repo       *repository.Repository
	httpClient *http.Client
}

func NewKitchenService(repo *repository.Repository) *KitchenService {
	return &KitchenService{
		repo:       repo,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func (s *KitchenService) GetRestaurants(ctx context.Context) ([]repository.Restaurant, error) {
	return s.repo.GetRestaurants(ctx)
}

func (s *KitchenService) GetMenu(ctx context.Context, restaurantID uuid.UUID) ([]repository.MenuItem, error) {
	return s.repo.GetMenuItemsByRestaurant(ctx, restaurantID)
}

func (s *KitchenService) GetOrder(ctx context.Context, orderID uuid.UUID) (repository.Order, error) {
	return s.repo.GetOrder(ctx, orderID)
}

func (s *KitchenService) UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status repository.OrderStatus) (repository.Order, error) {
	return s.repo.UpdateOrderStatus(ctx, orderID, status)
}


type OrderItemRequest struct {
	MenuItemID uuid.UUID `json:"menu_item_id"`
	Quantity   int       `json:"quantity"`
}

type OrderRequest struct {
	UserID       uuid.UUID          `json:"user_id"`
	RestaurantID uuid.UUID          `json:"restaurant_id"`
	Items        []OrderItemRequest `json:"items"`
}


type WebhookPayload struct {
	OrderID      uuid.UUID `json:"order_id"`
	RestaurantID uuid.UUID `json:"restaurant_id"`
	TotalAmount  int       `json:"total_amount"`
}

func (s *KitchenService) GetOrderItems(ctx context.Context, orderID uuid.UUID) ([]repository.OrderItem, error) {
	return s.repo.GetOrderItems(ctx, orderID)
}

func (s *KitchenService) CreateOrder(ctx context.Context, req OrderRequest) (repository.Order, error) {

	rest, err := s.repo.GetRestaurantByID(ctx, req.RestaurantID)
	if err != nil {
		return repository.Order{}, fmt.Errorf("restaurant not found: %w", err)
	}


	menuItems, err := s.repo.GetMenuItemsByRestaurant(ctx, req.RestaurantID)
	if err != nil {
		return repository.Order{}, fmt.Errorf("failed to get menu: %w", err)
	}

	menuMap := make(map[uuid.UUID]repository.MenuItem)
	for _, mi := range menuItems {
		menuMap[mi.ID] = mi
	}

	totalAmount := 0
	var orderItems []repository.OrderItem
	for _, item := range req.Items {
		mi, ok := menuMap[item.MenuItemID]
		if !ok {
			return repository.Order{}, fmt.Errorf("menu item %s not found or unavailable", item.MenuItemID)
		}
		totalAmount += mi.Price * item.Quantity
		orderItems = append(orderItems, repository.OrderItem{
			MenuItemID:   item.MenuItemID,
			Quantity:     item.Quantity,
			PriceAtOrder: mi.Price,
		})
	}

	if totalAmount == 0 {
		return repository.Order{}, fmt.Errorf("order cannot be empty")
	}


	order, err := s.repo.CreateOrderWithItems(ctx, req.UserID, req.RestaurantID, totalAmount, orderItems)
	if err != nil {
		return repository.Order{}, fmt.Errorf("failed to create order: %w", err)
	}


	go s.sendWebhook(rest.WebhookURL, order)

	return order, nil
}

func (s *KitchenService) sendWebhook(url string, order repository.Order) {
	payload := WebhookPayload{
		OrderID:      order.ID,
		RestaurantID: order.RestaurantID,
		TotalAmount:  order.TotalAmount,
	}
	
	body, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("[Webhook Error] failed to create request to %s: %v", url, err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		log.Printf("[Webhook Error] failed to send webhook to %s: %v", url, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		log.Printf("[Webhook Warning] restaurant returned status %d", resp.StatusCode)
	} else {
		log.Printf("[Webhook Success] order %s sent to %s", order.ID, url)
	}
}
