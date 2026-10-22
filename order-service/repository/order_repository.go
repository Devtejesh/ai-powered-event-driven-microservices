package repository

import (
	"context"
	"database/sql"

	"example.com/go-commerce-platform/order-service/models"
	"github.com/aarondl/sqlboiler/v4/boil"
)

type OrderRepository struct {
	DB *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{
		DB: db,
	}
}
func (r *OrderRepository) CreateOrder(ctx context.Context, order *models.Order) error {
	err := order.Insert(ctx, r.DB, boil.Infer())
	if err != nil {
		return err
	}
	return nil
}
func (r *OrderRepository) GetOrderByID(ctx context.Context, id int64) (*models.Order, error) {
	return models.FindOrder(ctx, r.DB, id)
}

func (r *OrderRepository) ListOrders(ctx context.Context) (models.OrderSlice, error) {
	return models.Orders().All(ctx, r.DB)
}

func (r *OrderRepository) UpdateOrder(ctx context.Context, order *models.Order) error {
	_, err := order.Update(
		ctx,
		r.DB,
		boil.Whitelist("product_name", "quantity", "unit_price", "status"),
	)
	return err
}

func (r *OrderRepository) DeleteOrder(ctx context.Context, id int64) error {
	order, err := models.FindOrder(ctx, r.DB, id)
	if err != nil {
		return err
	}

	_, err = order.Delete(ctx, r.DB)
	return err
}
