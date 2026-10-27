package service

import (
	"context"
	"fmt"

	"example.com/go-commerce-platform/order-service/kafka"
	"example.com/go-commerce-platform/order-service/models"
	"example.com/go-commerce-platform/order-service/repository"
	"github.com/ericlagergren/decimal"
)

type OrderService struct {
	Repo     *repository.OrderRepository
	Producer *kafka.Producer
}

func NewOrderService(
	repo *repository.OrderRepository,
	producer *kafka.Producer,
) *OrderService {
	return &OrderService{
		Repo:     repo,
		Producer: producer,
	}
}

func (s *OrderService) CreateOrder(
	ctx context.Context,
	order *models.Order,
) error {
	if order.UserID <= 0 {
		return fmt.Errorf("invalid user id")
	}

	if order.ProductID.Valid == false ||
		order.ProductID.Int64 <= 0 {
		return fmt.Errorf("valid product id is required")
	}

	if order.ProductName == "" {
		return fmt.Errorf("product name is required")
	}

	if order.Quantity <= 0 {
		return fmt.Errorf("quantity must be greater than 0")
	}

	if order.UnitPrice.Big == nil ||
		order.UnitPrice.Cmp(decimal.New(0, 0)) <= 0 {
		return fmt.Errorf("price must be greater than 0")
	}

	// Step 1: Save the order in the database.
	if err := s.Repo.CreateOrder(ctx, order); err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}

	// Step 2: Calculate the total order amount.
	total := new(decimal.Big).Mul(
		order.UnitPrice.Big,
		decimal.New(int64(order.Quantity), 0),
	)

	amount, ok := total.Float64()
	if !ok {
		return fmt.Errorf("failed to convert order amount to float64")
	}

	// Step 3: Create the Kafka event with product details.
	event := kafka.OrderCreatedEvent{
		Event:   "OrderCreated",
		OrderID: int(order.ID),
		UserID:  int(order.UserID),
		Amount:  amount,
		Items: []kafka.OrderItem{
			{
				ProductID: order.ProductID.Int64,
				Quantity:  order.Quantity,
			},
		},
	}

	// Step 4: Publish the event to Kafka.
	if err := s.Producer.PublishOrderCreated(ctx, event); err != nil {
		return fmt.Errorf(
			"order saved but event publishing failed: %w",
			err,
		)
	}

	return nil
}

func (s *OrderService) GetOrderByID(
	ctx context.Context,
	id int64,
) (*models.Order, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid order id")
	}

	return s.Repo.GetOrderByID(ctx, id)
}

func (s *OrderService) ListOrders(
	ctx context.Context,
) (models.OrderSlice, error) {
	return s.Repo.ListOrders(ctx)
}

func (s *OrderService) UpdateOrder(
	ctx context.Context,
	id int64,
	order *models.Order,
) error {
	if id <= 0 {
		return fmt.Errorf("invalid order id")
	}

	if order.ProductName == "" {
		return fmt.Errorf("product name is required")
	}

	if order.Quantity <= 0 {
		return fmt.Errorf("quantity must be greater than 0")
	}

	if order.UnitPrice.Big == nil ||
		order.UnitPrice.Cmp(decimal.New(0, 0)) <= 0 {
		return fmt.Errorf("price must be greater than 0")
	}

	switch order.Status {
	case "pending", "confirmed", "shipped", "delivered", "cancelled":
		// Valid order status.
	default:
		return fmt.Errorf("invalid order status")
	}

	// Use the ID received from the URL.
	order.ID = id

	return s.Repo.UpdateOrder(ctx, order)
}

func (s *OrderService) DeleteOrder(
	ctx context.Context,
	id int64,
) error {
	if id <= 0 {
		return fmt.Errorf("invalid order id")
	}

	return s.Repo.DeleteOrder(ctx, id)
}
