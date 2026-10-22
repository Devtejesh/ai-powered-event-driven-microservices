package service

import (
	"context"
	"fmt"

	"example.com/go-commerce-platform/order-service/models"
	"example.com/go-commerce-platform/order-service/repository"
	"github.com/ericlagergren/decimal"
)

type OrderService struct {
	Repo *repository.OrderRepository
}

func NewOrderService(repo *repository.OrderRepository) *OrderService {
	return &OrderService{
		Repo: repo,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, order *models.Order) error {
	if order.UserID <= 0 {
		return fmt.Errorf("invalid user id")
	}
	if order.ProductName == "" {
		return fmt.Errorf("product name is required")

	}
	if order.Quantity <= 0 {
		return fmt.Errorf("quantity must be greater than 0")
	}
	if order.UnitPrice.Big == nil || order.UnitPrice.Cmp(decimal.New(0, 0)) <= 0 {
		return fmt.Errorf("price must be greater than 0")
	}
	return s.Repo.CreateOrder(ctx, order)
}
func (s *OrderService) GetOrderByID(ctx context.Context, id int64) (*models.Order, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid order id")
	}

	return s.Repo.GetOrderByID(ctx, id)
}

func (s *OrderService) ListOrders(ctx context.Context) (models.OrderSlice, error) {
	return s.Repo.ListOrders(ctx)
}

func (s *OrderService) UpdateOrder(ctx context.Context, id int64, order *models.Order) error {
	if id <= 0 {
		return fmt.Errorf("invalid order id")
	}
	if order.ProductName == "" {
		return fmt.Errorf("product name is required")
	}
	if order.Quantity <= 0 {
		return fmt.Errorf("quantity must be greater than 0")
	}
	if order.UnitPrice.Big == nil || order.UnitPrice.Cmp(new(decimal.Big)) <= 0 {
		return fmt.Errorf("price must be greater than 0")
	}

	switch order.Status {
	case "pending", "confirmed", "shipped", "delivered", "cancelled":
		// Valid status.
	default:
		return fmt.Errorf("invalid order status")
	}

	// Path nunchi vachina ID matrame use chestunnam.
	order.ID = id

	return s.Repo.UpdateOrder(ctx, order)
}

func (s *OrderService) DeleteOrder(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid order id")
	}

	return s.Repo.DeleteOrder(ctx, id)
}
