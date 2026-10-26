package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"example.com/go-commerce-platform/inventory-service/models"
	"example.com/go-commerce-platform/inventory-service/repository"
)

var (
	ErrInvalidInput      = errors.New("invalid input")
	ErrProductNotFound   = errors.New("product not found")
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrDuplicateSKU      = errors.New("SKU already exists")
)

type InventoryService struct {
	Repo *repository.InventoryRepository
}

func NewInventoryService(repo *repository.InventoryRepository) *InventoryService {
	return &InventoryService{Repo: repo}
}

func (s *InventoryService) Create(
	ctx context.Context,
	name, sku string,
	quantity int,
) (*models.Product, error) {
	name = strings.TrimSpace(name)
	sku = strings.TrimSpace(sku)

	if name == "" || sku == "" || quantity < 0 {
		return nil, ErrInvalidInput
	}

	product, err := s.Repo.Create(ctx, name, sku, quantity)
	if err != nil {
		if repository.IsDuplicateSKUError(err) {
			return nil, ErrDuplicateSKU
		}
		return nil, err
	}

	return product, nil
}

func (s *InventoryService) GetByID(
	ctx context.Context,
	id int64,
) (*models.Product, error) {
	if id <= 0 {
		return nil, ErrInvalidInput
	}

	product, err := s.Repo.GetByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrProductNotFound
	}
	if err != nil {
		return nil, err
	}

	return product, nil
}

func (s *InventoryService) List(
	ctx context.Context,
) ([]*models.Product, error) {
	return s.Repo.List(ctx)
}

func (s *InventoryService) Update(
	ctx context.Context,
	id int64,
	name string,
) (*models.Product, error) {
	name = strings.TrimSpace(name)

	if id <= 0 || name == "" {
		return nil, ErrInvalidInput
	}

	product, err := s.Repo.Update(ctx, id, name)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrProductNotFound
	}
	if err != nil {
		return nil, err
	}

	return product, nil
}

func (s *InventoryService) ChangeStock(
	ctx context.Context,
	id int64,
	delta int,
) (*models.Product, error) {
	if id <= 0 || delta == 0 {
		return nil, ErrInvalidInput
	}

	product, err := s.Repo.ChangeStock(ctx, id, delta)
	if errors.Is(err, sql.ErrNoRows) {
		existing, lookupErr := s.Repo.GetByID(ctx, id)
		if errors.Is(lookupErr, sql.ErrNoRows) {
			return nil, ErrProductNotFound
		}
		if lookupErr != nil {
			return nil, lookupErr
		}
		if existing.StockQuantity+delta < 0 {
			return nil, ErrInsufficientStock
		}
		return nil, fmt.Errorf("change stock failed")
	}
	if err != nil {
		return nil, err
	}

	return product, nil
}
