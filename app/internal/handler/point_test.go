package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"point-service/app/internal/handler"
	"point-service/app/internal/model"
	mockService "point-service/app/internal/service/mocks"
	"testing"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type PointHandlerTestSuite struct {
	suite.Suite
	handler handler.PointHandler
}

func (suite *PointHandlerTestSuite) SetupTest() {
	pointService := new(mockService.PointService)
	pointService.On("DecreasePoint", mock.Anything, model.SuccessOrder{OrderId: 1, ProductId: 1}).Return(nil)
	pointService.On("DecreasePoint", mock.Anything, model.SuccessOrder{}).Return(errors.New("decrease point error"))

	suite.handler = handler.NewPointHandler(pointService)
}

func (suite *PointHandlerTestSuite) TestPointHandler_HappyCase() {
	successOrder := model.SuccessOrder{OrderId: 1, ProductId: 1}
	b, _ := json.Marshal(successOrder)
	message := sarama.ConsumerMessage{
		Value: b,
	}

	err := suite.handler.SuccessOrderProcess(context.Background(), &message)
	suite.Nil(err)
}

func (suite *PointHandlerTestSuite) TestPointHandler_UnmarshalError() {
	successOrder := "invalid body"
	b, _ := json.Marshal(successOrder)
	message := sarama.ConsumerMessage{
		Value: b,
	}

	err := suite.handler.SuccessOrderProcess(context.Background(), &message)
	suite.Nil(err)
}

func (suite *PointHandlerTestSuite) TestPointHandler_DecreasePointError() {
	successOrder := model.SuccessOrder{}
	b, _ := json.Marshal(successOrder)
	message := sarama.ConsumerMessage{
		Value: b,
	}

	err := suite.handler.SuccessOrderProcess(context.Background(), &message)
	suite.Nil(err)
}

func TestPointHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(PointHandlerTestSuite))
}
