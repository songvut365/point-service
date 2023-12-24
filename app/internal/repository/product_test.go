package repository_test

import (
	"context"
	"errors"
	"point-service/app/internal/repository"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type ProductRepositoryTestSuite struct {
	suite.Suite
}

func (suite *ProductRepositoryTestSuite) SetupTest() {}

func (suite *ProductRepositoryTestSuite) setupDbMockCustomTrx(process func(sqlmock.Sqlmock)) *gorm.DB {
	// new mock instance
	mockDb, sqlMock, err := sqlmock.New()
	if err != nil {
		panic(err)
	}

	// new postgres dialector for gorm
	dialector := postgres.New(postgres.Config{
		Conn:       mockDb,
		DriverName: "postgres",
	})

	process(sqlMock)

	// initialize gorm database
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		panic(err)
	}

	return db
}

func (suite *ProductRepositoryTestSuite) TestProductRepository_HappyCase() {
	db := suite.setupDbMockCustomTrx(func(sqlMock sqlmock.Sqlmock) {
		rows := sqlmock.NewRows([]string{"id", "name", "price"}).AddRow(1, "mobile suite", 1500)
		sqlMock.ExpectQuery(regexp.QuoteMeta(`
			SELECT * FROM "products" 
			WHERE id = $1
			AND "products"."deleted_at" IS NULL 
			ORDER BY "products"."id" 
			LIMIT 1
		`)).WithArgs(1).WillReturnRows(rows)
	})
	repository := repository.NewProductRepository(db)

	product, err := repository.GetProductById(context.Background(), 1)
	suite.Nil(err)
	suite.NotNil(product)
	suite.Equal(product.Name, "mobile suite")
	suite.Equal(product.Price, float64(1500))

}

func (suite *ProductRepositoryTestSuite) TestProductRepository_GetProductError() {
	db := suite.setupDbMockCustomTrx(func(sqlMock sqlmock.Sqlmock) {
		sqlMock.ExpectQuery(regexp.QuoteMeta(`
			SELECT * FROM "products" 
			WHERE id = $1
			AND "products"."deleted_at" IS NULL 
			ORDER BY "products"."id" 
			LIMIT 1
		`)).WithArgs(1).WillReturnError(errors.New("select product error"))
	})
	repository := repository.NewProductRepository(db)

	_, err := repository.GetProductById(context.Background(), 1)
	suite.NotNil(err)
}

func TestProductRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(ProductRepositoryTestSuite))
}
