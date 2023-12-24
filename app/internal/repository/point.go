package repository

import (
	"context"
	"errors"
	"point-service/app/internal/constant"
	"point-service/app/internal/model"
	"time"

	"gorm.io/gorm"
)

type PointRepository interface {
	DecreaseBronzePoint(ctx context.Context) error
	DecreaseSilverPoint(ctx context.Context) error
	DecreaseGoldPoint(ctx context.Context) error
}

type pointRepository struct {
	db         *gorm.DB
	waitTime   time.Duration
	maxAttempt uint
}

func NewPointRepository(db *gorm.DB, waitTime time.Duration, maxAttempt uint) PointRepository {
	return &pointRepository{
		db:         db,
		waitTime:   waitTime,
		maxAttempt: maxAttempt,
	}
}

func (repository *pointRepository) DecreaseBronzePoint(ctx context.Context) error {
	return repository.decreasePoint(ctx, constant.BRONZE)
}

func (repository *pointRepository) DecreaseSilverPoint(ctx context.Context) error {
	return repository.decreasePoint(ctx, constant.SILVER)
}

func (repository *pointRepository) DecreaseGoldPoint(ctx context.Context) error {
	return repository.decreasePoint(ctx, constant.GOLD)
}

func (repository *pointRepository) decreasePoint(ctx context.Context, level string) error {
	attempt := 1

	tx := repository.db.Begin().WithContext(ctx)
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// optimistic locking
	for {
		var point model.Point

		// find remaining point
		err := tx.Model(&model.Point{}).Where("level = ?", level).First(&point).Error
		if err != nil {
			tx.Rollback()
			return err
		}

		// decrease point
		if point.Remaining <= 0 {
			tx.Rollback()
			return errors.New("not enough points")
		}

		remaining := point.Remaining - 1

		// update point after decrease
		result := repository.db.Model(&model.Point{}).
			Where("updated_at = ? AND level = ?", point.UpdatedAt, level).
			Update("remaining", remaining)

		if result.Error != nil {
			tx.Rollback()
			return result.Error
		}

		// update success
		if result.RowsAffected == 1 {
			tx.Commit()
			return nil
		}

		if attempt == int(repository.maxAttempt) {
			return errors.New("maximum attempts reached")
		}

		time.Sleep(repository.waitTime)
		attempt++
	}
}
