package repository

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Restaurant struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	WebhookURL string    `json:"webhook_url"`
	CreatedAt  time.Time `json:"created_at"`
}

type MenuItem struct {
	ID           uuid.UUID `json:"id"`
	RestaurantID uuid.UUID `json:"restaurant_id"`
	Name         string    `json:"name"`
	Description  *string   `json:"description"`
	Price        int       `json:"price"`
	IsAvailable  bool      `json:"is_available"`
	CreatedAt    time.Time `json:"created_at"`
}

type OrderStatus string

const (
	OrderStatusCreated    OrderStatus = "CREATED"
	OrderStatusAccepted   OrderStatus = "ACCEPTED"
	OrderStatusPreparing  OrderStatus = "PREPARING"
	OrderStatusDelivering OrderStatus = "DELIVERING"
	OrderStatusDelivered  OrderStatus = "DELIVERED"
	OrderStatusCancelled  OrderStatus = "CANCELLED"
)

type Order struct {
	ID           uuid.UUID   `json:"id"`
	UserID       uuid.UUID   `json:"user_id"`
	RestaurantID uuid.UUID   `json:"restaurant_id"`
	Status       OrderStatus `json:"status"`
	TotalAmount  int         `json:"total_amount"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}

type OrderItem struct {
	ID           uuid.UUID `json:"id"`
	OrderID      uuid.UUID `json:"order_id"`
	MenuItemID   uuid.UUID `json:"menu_item_id"`
	Quantity     int       `json:"quantity"`
	PriceAtOrder int       `json:"price_at_order"`
}
