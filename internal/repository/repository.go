package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetRestaurants(ctx context.Context) ([]Restaurant, error) {
	query := `SELECT id, name, webhook_url, created_at FROM restaurants`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Restaurant
	for rows.Next() {
		var i Restaurant
		if err := rows.Scan(&i.ID, &i.Name, &i.WebhookURL, &i.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

func (r *Repository) GetRestaurantByID(ctx context.Context, id uuid.UUID) (Restaurant, error) {
	query := `SELECT id, name, webhook_url, created_at FROM restaurants WHERE id = $1`
	var i Restaurant
	err := r.db.QueryRow(ctx, query, id).Scan(&i.ID, &i.Name, &i.WebhookURL, &i.CreatedAt)
	return i, err
}

func (r *Repository) CreateRestaurant(ctx context.Context, name, webhookURL string) (Restaurant, error) {
	query := `INSERT INTO restaurants (name, webhook_url) VALUES ($1, $2) RETURNING id, name, webhook_url, created_at`
	var i Restaurant
	err := r.db.QueryRow(ctx, query, name, webhookURL).Scan(&i.ID, &i.Name, &i.WebhookURL, &i.CreatedAt)
	return i, err
}

func (r *Repository) GetMenuItemsByRestaurant(ctx context.Context, restaurantID uuid.UUID) ([]MenuItem, error) {
	query := `SELECT id, restaurant_id, name, description, price, is_available, created_at FROM menu_items WHERE restaurant_id = $1 AND is_available = true`
	rows, err := r.db.Query(ctx, query, restaurantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []MenuItem
	for rows.Next() {
		var i MenuItem
		if err := rows.Scan(&i.ID, &i.RestaurantID, &i.Name, &i.Description, &i.Price, &i.IsAvailable, &i.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

func (r *Repository) CreateOrderWithItems(ctx context.Context, userID, restaurantID uuid.UUID, totalAmount int, items []OrderItem) (Order, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Order{}, err
	}
	defer tx.Rollback(ctx)

	query := `INSERT INTO orders (user_id, restaurant_id, status, total_amount) VALUES ($1, $2, $3, $4) RETURNING id, user_id, restaurant_id, status, total_amount, created_at, updated_at`
	var o Order
	err = tx.QueryRow(ctx, query, userID, restaurantID, OrderStatusCreated, totalAmount).Scan(
		&o.ID, &o.UserID, &o.RestaurantID, &o.Status, &o.TotalAmount, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		return Order{}, err
	}

	for _, item := range items {
		_, err := tx.Exec(ctx, `INSERT INTO order_items (order_id, menu_item_id, quantity, price_at_order) VALUES ($1, $2, $3, $4)`,
			o.ID, item.MenuItemID, item.Quantity, item.PriceAtOrder)
		if err != nil {
			return Order{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return Order{}, err
	}

	return o, nil
}

func (r *Repository) GetOrder(ctx context.Context, id uuid.UUID) (Order, error) {
	query := `SELECT id, user_id, restaurant_id, status, total_amount, created_at, updated_at FROM orders WHERE id = $1`
	var i Order
	err := r.db.QueryRow(ctx, query, id).Scan(
		&i.ID, &i.UserID, &i.RestaurantID, &i.Status, &i.TotalAmount, &i.CreatedAt, &i.UpdatedAt,
	)
	return i, err
}

func (r *Repository) GetOrderItems(ctx context.Context, orderID uuid.UUID) ([]OrderItem, error) {
	query := `SELECT id, order_id, menu_item_id, quantity, price_at_order FROM order_items WHERE order_id = $1`
	rows, err := r.db.Query(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []OrderItem
	for rows.Next() {
		var i OrderItem
		if err := rows.Scan(&i.ID, &i.OrderID, &i.MenuItemID, &i.Quantity, &i.PriceAtOrder); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

func (r *Repository) UpdateOrderStatus(ctx context.Context, id uuid.UUID, status OrderStatus) (Order, error) {
	query := `UPDATE orders SET status = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1 RETURNING id, user_id, restaurant_id, status, total_amount, created_at, updated_at`
	var i Order
	err := r.db.QueryRow(ctx, query, id, status).Scan(
		&i.ID, &i.UserID, &i.RestaurantID, &i.Status, &i.TotalAmount, &i.CreatedAt, &i.UpdatedAt,
	)
	return i, err
}
