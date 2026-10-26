package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"example.com/go-commerce-platform/inventory-service/models"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/lib/pq"
)

type InventoryRepository struct {
	DB *sql.DB
}

func NewInventoryRepository(db *sql.DB) *InventoryRepository {
	return &InventoryRepository{DB: db}
}

func (r *InventoryRepository) Create(
	ctx context.Context,
	name, sku string,
	quantity int,
) (*models.Product, error) {
	product := &models.Product{
		Name:          name,
		Sku:           sku,
		StockQuantity: quantity,
	}

	if err := product.Insert(
		ctx,
		r.DB,
		boil.Infer(),
	); err != nil {
		return nil, fmt.Errorf("create product: %w", err)
	}

	return product, nil
}

func (r *InventoryRepository) GetByID(
	ctx context.Context,
	id int64,
) (*models.Product, error) {
	product, err := models.Products(
		qm.Where("id = ?", id),
	).One(ctx, r.DB)

	if err != nil {
		return nil, fmt.Errorf("get product: %w", err)
	}

	return product, nil
}

func (r *InventoryRepository) List(
	ctx context.Context,
) ([]*models.Product, error) {
	products, err := models.Products(
		qm.OrderBy("id ASC"),
	).All(ctx, r.DB)

	if err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}

	return products, nil
}

func (r *InventoryRepository) Update(
	ctx context.Context,
	id int64,
	name string,
) (*models.Product, error) {
	product, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	product.Name = name

	_, err = product.Update(
		ctx,
		r.DB,
		boil.Whitelist("name", "updated_at"),
	)
	if err != nil {
		return nil, fmt.Errorf("update product: %w", err)
	}

	return product, nil
}

func (r *InventoryRepository) ChangeStock(
	ctx context.Context,
	id int64,
	delta int,
) (*models.Product, error) {
	query := `
		UPDATE products
		SET stock_quantity = stock_quantity + $1,
		    updated_at = NOW()
		WHERE id = $2
		  AND stock_quantity + $1 >= 0
		RETURNING id, name, sku, stock_quantity, created_at, updated_at
	`

	product := &models.Product{}

	err := r.DB.QueryRowContext(ctx, query, delta, id).Scan(
		&product.ID,
		&product.Name,
		&product.Sku,
		&product.StockQuantity,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("change stock: %w", err)
	}

	return product, nil
}

func IsDuplicateSKUError(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}

func IsNotFoundError(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}
