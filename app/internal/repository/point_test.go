package repository_test

import (
	"context"
	"errors"
	"point-service/app/internal/repository"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PointRepositoryTestSuite struct {
	suite.Suite
}

func (suite *PointRepositoryTestSuite) SetupTest() {}

func (suite *PointRepositoryTestSuite) setupDbMockTrxSuccess(level string) *gorm.DB {
	var err error

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

	// step of success case in a transaction
	// step1: expect beginning of transaction
	sqlMock.ExpectBegin()

	// step2: expect query point by level
	rows := sqlmock.NewRows([]string{"id", "level", "remaining"}).AddRow(1, level, 1000)
	sqlMock.ExpectQuery(regexp.QuoteMeta(`
		SELECT * FROM "points" 
		WHERE level = $1
		AND "points"."deleted_at" IS NULL 
		ORDER BY "points"."id" 
		LIMIT 1
	`)).WithArgs(level).WillReturnRows(rows)

	// step3: expect beginning of transaction again
	sqlMock.ExpectBegin()

	// step4: expect update remaining point by level
	sqlMock.ExpectExec(regexp.QuoteMeta(`
		UPDATE "points" 
		SET "remaining"=$1,"updated_at"=$2 
		WHERE (updated_at = $3 AND level = $4) 
		AND "points"."deleted_at" IS NULL
	`)).WithArgs(999, sqlmock.AnyArg(), sqlmock.AnyArg(), level).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// step5: expect commit of transaction
	sqlMock.ExpectCommit()

	// initialize gorm database
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		panic(err)
	}

	return db
}

func (suite *PointRepositoryTestSuite) setupDbMockCustomTrx(level string, process func(sqlmock.Sqlmock)) *gorm.DB {
	var err error

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

func (suite *PointRepositoryTestSuite) TestPointRepository_HappyCase_DecreaseBronze() {
	db := suite.setupDbMockTrxSuccess("bronze")
	repository := repository.NewPointRepository(db, time.Second, 3)

	err := repository.DecreaseBronzePoint(context.Background())
	suite.Nil(err)
}

func (suite *PointRepositoryTestSuite) TestPointRepository_HappyCase_DecreaseSilver() {
	db := suite.setupDbMockTrxSuccess("silver")
	repository := repository.NewPointRepository(db, time.Second, 3)

	err := repository.DecreaseSilverPoint(context.Background())
	suite.Nil(err)
}

func (suite *PointRepositoryTestSuite) TestPointRepository_HappyCase_DecreaseGold() {
	db := suite.setupDbMockTrxSuccess("gold")
	repository := repository.NewPointRepository(db, time.Second, 3)

	err := repository.DecreaseGoldPoint(context.Background())
	suite.Nil(err)
}

func (suite *PointRepositoryTestSuite) TestPointRepository_BeginTrxError() {
	db := suite.setupDbMockCustomTrx("bronze", func(sqlMock sqlmock.Sqlmock) {
		sqlMock.ExpectBegin()
		sqlMock.ExpectQuery(regexp.QuoteMeta(`
			SELECT * FROM "points" 
			WHERE level = $1
			AND "points"."deleted_at" IS NULL 
			ORDER BY "points"."id" 
			LIMIT 1
		`)).WithArgs("bronze").WillReturnError(errors.New("select error"))
		sqlMock.ExpectRollback()
	})

	repository := repository.NewPointRepository(db, time.Second, 3)

	err := repository.DecreaseBronzePoint(context.Background())
	suite.NotNil(err)
}

func (suite *PointRepositoryTestSuite) TestPointRepository_RemainingPointLessThanZero() {
	db := suite.setupDbMockCustomTrx("bronze", func(sqlMock sqlmock.Sqlmock) {
		sqlMock.ExpectBegin()

		rows := sqlmock.NewRows([]string{"id", "level", "remaining"}).AddRow(1, "bronze", 0)
		sqlMock.ExpectQuery(regexp.QuoteMeta(`
			SELECT * FROM "points" 
			WHERE level = $1
			AND "points"."deleted_at" IS NULL 
			ORDER BY "points"."id" 
			LIMIT 1
		`)).WithArgs("bronze").WillReturnRows(rows)

		sqlMock.ExpectRollback()
	})

	repository := repository.NewPointRepository(db, time.Second, 3)

	err := repository.DecreaseBronzePoint(context.Background())
	suite.NotNil(err)
}

func (suite *PointRepositoryTestSuite) TestPointRepository_UpdatePointError() {
	db := suite.setupDbMockCustomTrx("bronze", func(sqlMock sqlmock.Sqlmock) {
		sqlMock.ExpectBegin()

		rows := sqlmock.NewRows([]string{"id", "level", "remaining"}).AddRow(1, "bronze", 1000)
		sqlMock.ExpectQuery(regexp.QuoteMeta(`
			SELECT * FROM "points" 
			WHERE level = $1
			AND "points"."deleted_at" IS NULL 
			ORDER BY "points"."id" 
			LIMIT 1
		`)).WithArgs("bronze").WillReturnRows(rows)

		sqlMock.ExpectBegin()

		sqlMock.ExpectExec(regexp.QuoteMeta(`
			UPDATE "points" 
			SET "remaining"=$1,"updated_at"=$2 
			WHERE (updated_at = $3 AND level = $4) 
			AND "points"."deleted_at" IS NULL
		`)).WithArgs(999, sqlmock.AnyArg(), sqlmock.AnyArg(), "bronze").
			WillReturnError(errors.New("update error"))

		sqlMock.ExpectRollback()
	})

	repository := repository.NewPointRepository(db, time.Second, 3)

	err := repository.DecreaseBronzePoint(context.Background())
	suite.NotNil(err)
}

func (suite *PointRepositoryTestSuite) TestPointRepository_AttemptIsOne() {
	db := suite.setupDbMockCustomTrx("bronze", func(sqlMock sqlmock.Sqlmock) {
		sqlMock.ExpectBegin()

		firstRow := sqlmock.NewRows([]string{"id", "level", "remaining"}).AddRow(1, "bronze", 1000)
		sqlMock.ExpectQuery(regexp.QuoteMeta(`
			SELECT * FROM "points" 
			WHERE level = $1
			AND "points"."deleted_at" IS NULL 
			ORDER BY "points"."id" 
			LIMIT 1
		`)).WithArgs("bronze").WillReturnRows(firstRow)

		sqlMock.ExpectBegin()

		sqlMock.ExpectExec(regexp.QuoteMeta(`
			UPDATE "points" 
			SET "remaining"=$1,"updated_at"=$2 
			WHERE (updated_at = $3 AND level = $4) 
			AND "points"."deleted_at" IS NULL
		`)).WithArgs(999, sqlmock.AnyArg(), sqlmock.AnyArg(), "bronze").
			WillReturnResult(sqlmock.NewResult(0, 0))

		sqlMock.ExpectCommit()

		secondRow := sqlmock.NewRows([]string{"id", "level", "remaining"}).AddRow(1, "bronze", 1000)
		sqlMock.ExpectQuery(regexp.QuoteMeta(`
			SELECT * FROM "points" 
			WHERE level = $1
			AND "points"."deleted_at" IS NULL 
			ORDER BY "points"."id" 
			LIMIT 1
		`)).WithArgs("bronze").WillReturnRows(secondRow)
	})

	repository := repository.NewPointRepository(db, time.Second, 1)

	err := repository.DecreaseBronzePoint(context.Background())
	suite.NotNil(err)
}

func (suite *PointRepositoryTestSuite) TestPointRepository_AttemptIsTwo() {
	db := suite.setupDbMockCustomTrx("bronze", func(sqlMock sqlmock.Sqlmock) {
		sqlMock.ExpectBegin()

		firstRow := sqlmock.NewRows([]string{"id", "level", "remaining"}).AddRow(1, "bronze", 1000)
		sqlMock.ExpectQuery(regexp.QuoteMeta(`
			SELECT * FROM "points" 
			WHERE level = $1
			AND "points"."deleted_at" IS NULL 
			ORDER BY "points"."id" 
			LIMIT 1
		`)).WithArgs("bronze").WillReturnRows(firstRow)

		sqlMock.ExpectBegin()

		sqlMock.ExpectExec(regexp.QuoteMeta(`
			UPDATE "points" 
			SET "remaining"=$1,"updated_at"=$2 
			WHERE (updated_at = $3 AND level = $4) 
			AND "points"."deleted_at" IS NULL
		`)).WithArgs(999, sqlmock.AnyArg(), sqlmock.AnyArg(), "bronze").
			WillReturnResult(sqlmock.NewResult(0, 0))

		sqlMock.ExpectCommit()

		secondRow := sqlmock.NewRows([]string{"id", "level", "remaining"}).AddRow(1, "bronze", 1000)
		sqlMock.ExpectQuery(regexp.QuoteMeta(`
			SELECT * FROM "points" 
			WHERE level = $1
			AND "points"."deleted_at" IS NULL 
			ORDER BY "points"."id" 
			LIMIT 1
		`)).WithArgs("bronze").WillReturnRows(secondRow)
	})

	repository := repository.NewPointRepository(db, time.Second, 2)

	err := repository.DecreaseBronzePoint(context.Background())
	suite.NotNil(err)
}

func TestPointRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(PointRepositoryTestSuite))
}
