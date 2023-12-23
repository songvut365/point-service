package service_test

import (
	"context"
	"errors"
	"point-service/app/internal/model"
	mockRepository "point-service/app/internal/repository/mocks"
	"point-service/app/internal/service"
	mockKafka "point-service/app/pkg/kafka/mocks"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type PointServiceTestSuite struct {
	suite.Suite
	pointService service.PointService
}

func (suite *PointServiceTestSuite) SetupTest() {

	pointRepository := new(mockRepository.PointRepository)
	pointRepository.On("DecreaseBronzePoint", mock.Anything).Return(nil)
	pointRepository.On("DecreaseSilverPoint", mock.Anything).Return(nil)
	pointRepository.On("DecreaseGoldPoint", mock.Anything).Return(nil)

	productRepository := new(mockRepository.ProductRepository)
	productRepository.On("GetProductById", mock.Anything, uint(1)).Return(model.Product{Name: "mobile suite", Price: 1500}, nil)
	productRepository.On("GetProductById", mock.Anything, uint(2)).Return(model.Product{Name: "jaeger", Price: 800}, nil)
	productRepository.On("GetProductById", mock.Anything, uint(3)).Return(model.Product{Name: "car", Price: 77}, nil)
	productRepository.On("GetProductById", mock.Anything, uint(4)).Return(model.Product{}, errors.New("get product error"))

	producer := new(mockKafka.Producer)
	// producer.On("SendMessage", "decrease.point.success", mock.Anything, mock.Anything).Return(nil)
	producer.On("SendMessage", "decrease.point.success", `{"order_id":1,"point_level":"gold"}`, mock.Anything).Return(nil)
	producer.On("SendMessage", "decrease.point.success", `{"order_id":2,"point_level":"silver"}`, mock.Anything).Return(nil)
	producer.On("SendMessage", "decrease.point.success", `{"order_id":3,"point_level":"bronze"}`, mock.Anything).Return(nil)
	producer.On("SendMessage", "decrease.point.success", `{"order_id":5,"point_level":"gold"}`, mock.Anything).Return(errors.New("produce message error"))

	suite.pointService = service.NewPointService(pointRepository, productRepository, producer, "decrease.point.success")

}

func (suite *PointServiceTestSuite) TestPointService_HappyCase_DecreaseGold() {
	ctx := context.Background()
	successOrder := model.SuccessOrder{
		OrderId:   1,
		ProductId: 1,
	}

	err := suite.pointService.DecreasePoint(ctx, successOrder)
	suite.Empty(err)
}

func (suite *PointServiceTestSuite) TestPointService_HappyCase_DecreaseSilver() {
	ctx := context.Background()
	successOrder := model.SuccessOrder{
		OrderId:   2,
		ProductId: 2,
	}

	err := suite.pointService.DecreasePoint(ctx, successOrder)
	suite.Empty(err)
}

func (suite *PointServiceTestSuite) TestPointService_HappyCase_DecreaseBronze() {
	ctx := context.Background()
	successOrder := model.SuccessOrder{
		OrderId:   3,
		ProductId: 3,
	}

	err := suite.pointService.DecreasePoint(ctx, successOrder)
	suite.Empty(err)
}

func (suite *PointServiceTestSuite) TestPointService_GetProductError() {
	ctx := context.Background()
	successOrder := model.SuccessOrder{
		OrderId:   4,
		ProductId: 4,
	}

	err := suite.pointService.DecreasePoint(ctx, successOrder)
	suite.NotNil(err)
}

func (suite *PointServiceTestSuite) TestPointService_ProduceError() {
	ctx := context.Background()
	successOrder := model.SuccessOrder{
		OrderId:   5,
		ProductId: 1,
	}

	err := suite.pointService.DecreasePoint(ctx, successOrder)
	suite.NotNil(err)
}

func TestPointServiceTestSuite(t *testing.T) {
	suite.Run(t, new(PointServiceTestSuite))
}
