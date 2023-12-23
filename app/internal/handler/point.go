package handler

import (
	"context"
	"encoding/json"
	"log"
	"point-service/app/internal/model"
	"point-service/app/internal/service"

	"github.com/IBM/sarama"
)

type PointHandler interface {
	SuccessOrderProcess(ctx context.Context, message *sarama.ConsumerMessage) error
}

type pointHandler struct {
	pointService service.PointService
}

func NewPointHandler(pointService service.PointService) PointHandler {
	return &pointHandler{
		pointService: pointService,
	}
}

func (handler *pointHandler) SuccessOrderProcess(ctx context.Context, message *sarama.ConsumerMessage) error {
	var successOrder model.SuccessOrder
	err := json.Unmarshal(message.Value, &successOrder)
	if err != nil {
		log.Printf("unmarshal message value error: %s", err.Error())
		return nil
	}

	err = handler.pointService.DecreasePoint(ctx, successOrder)
	if err != nil {
		log.Printf("decrease point error: %s", err.Error())
		return nil
	}

	return nil
}
