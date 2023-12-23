package service

import (
	"context"
	"encoding/json"
	"point-service/app/internal/constant"
	"point-service/app/internal/model"
	"point-service/app/internal/repository"
	"point-service/app/pkg/kafka"

	"github.com/pkg/errors"
)

type PointService interface {
	DecreasePoint(ctx context.Context, successOrder model.SuccessOrder) error
}

type pointService struct {
	pointRepository           repository.PointRepository
	productRepository         repository.ProductRepository
	producer                  kafka.Producer
	decreasePointSuccessTopic string
}

func NewPointService(pointRepository repository.PointRepository, productRepository repository.ProductRepository, producer kafka.Producer, decreasePointSuccessTopic string) PointService {
	return &pointService{
		pointRepository:           pointRepository,
		productRepository:         productRepository,
		producer:                  producer,
		decreasePointSuccessTopic: decreasePointSuccessTopic,
	}
}

func (service *pointService) DecreasePoint(ctx context.Context, successOrder model.SuccessOrder) error {
	decreasePointSuccess := model.DecreasePointSuccess{
		OrderId: successOrder.OrderId,
	}

	// find price of product
	product, err := service.productRepository.GetProductById(ctx, successOrder.ProductId)
	if err != nil {
		return errors.Wrap(err, "get product by id error")
	}

	// decrease point by success order transaction
	switch {
	case product.Price >= 1001:
		decreasePointSuccess.PointLevel = constant.GOLD

		err := service.pointRepository.DecreaseGoldPoint(ctx)
		if err != nil {
			return errors.Wrap(err, "decrease gold point error")
		}

	case product.Price >= 101 && product.Price <= 1000:
		decreasePointSuccess.PointLevel = constant.SILVER

		err := service.pointRepository.DecreaseSilverPoint(ctx)
		if err != nil {
			return errors.Wrap(err, "decrease silver point error")
		}

	case product.Price <= 100:
		decreasePointSuccess.PointLevel = constant.BRONZE

		err := service.pointRepository.DecreaseBronzePoint(ctx)
		if err != nil {
			return errors.Wrap(err, "decrease bronze point error")
		}

	default:
		return errors.New("unexpected price category")
	}

	// produce decrease point result for increase user point
	decreasePointSuccessJson, err := json.Marshal(decreasePointSuccess)
	if err != nil {
		return errors.Wrap(err, "marshal increase point error")
	}

	decreasePointSuccessHeader := map[string]string{}

	err = service.producer.SendMessage(
		service.decreasePointSuccessTopic,
		string(decreasePointSuccessJson),
		decreasePointSuccessHeader,
	)
	if err != nil {
		return errors.Wrap(err, "produce message decrease point result error")
	}

	return nil
}
