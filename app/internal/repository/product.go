package repository

import (
	"context"
	"point-service/app/internal/model"

	"gorm.io/gorm"
)

type ProductRepository interface {
	GetProductById(ctx context.Context, productId uint) (model.Product, error)
}

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{
		db: db,
	}
}

func (repository *productRepository) GetProductById(ctx context.Context, productId uint) (model.Product, error) {
	var product model.Product

	err := repository.db.WithContext(ctx).Model(&model.Product{}).Where("id = ?", productId).First(&product).Error
	if err != nil {
		return product, err
	}

	return product, nil
}
