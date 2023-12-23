package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"point-service/app/internal/handler"
	"point-service/app/internal/model"
	"point-service/app/internal/repository"
	"point-service/app/internal/service"
	"point-service/app/pkg/kafka"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	// postgresql config
	dsn        = "host=localhost user=postgresusr password=1234 dbname=songvutdb port=5432 sslmode=disable TimeZone=Asia/Bangkok"
	waitTime   = time.Millisecond * 100
	maxAttempt = 1000

	// kafka config
	brokerAddress             = []string{"localhost:9092"}
	consumerGroupId           = "point-service"
	topicSuccessOrder         = "success.order"
	topicDecreasePointSuccess = "decrease.point.success"
)

func main() {
	// DATABASE
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Panicf("connect to database error: %s", err.Error())
	}
	log.Println("connect database success")

	err = db.AutoMigrate(&model.Point{}, &model.Product{})
	if err != nil {
		log.Panicf("auto migration error: %s", err.Error())
	}
	log.Println("database auto migration success")

	// KAFKA PRODUCER
	producer, err := kafka.NewProducer(brokerAddress)
	if err != nil {
		log.Panicf("new producer error: %s", err.Error())
	}
	log.Println("kafka producer is ready...")

	// REPOSITORY, SERVICE, HANDLER
	productRepository := repository.NewProductRepository(db)
	pointRepository := repository.NewPointRepository(db, waitTime, uint(maxAttempt))
	pointService := service.NewPointService(pointRepository, productRepository, producer, topicDecreasePointSuccess)
	pointHandler := handler.NewPointHandler(pointService)

	// KAFKA CONSUMER
	log.Println("Starting a new Sarama consumer")
	kafkaCtx := context.Background()
	consumerGroup, err := kafka.NewConsumerGroup(kafkaCtx, consumerGroupId, brokerAddress)

	consumer := kafka.NewConsumer(pointHandler.SuccessOrderProcess)
	go func() {
		for {
			err = consumerGroup.Consume(kafkaCtx, []string{topicSuccessOrder}, &consumer)
			if err != nil {
				if errors.Is(err, sarama.ErrClosedConsumerGroup) {
					return
				}
				log.Panicf("consume message error: %s", err.Error())
			}

			if kafkaCtx.Err() != nil {
				return
			}
		}
	}()
	log.Println("kafka consumer up and running!...")

	// GRACEFUL SHUTDOWN
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-signalChannel:
		log.Println("terminating: via signal")
	case <-kafkaCtx.Done():
		log.Println("terminating: kafka context cancelled")
	}

	err = consumerGroup.Close()
	if err != nil {
		log.Panicf("closing consumer group error: %s", err.Error())
	}

	err = producer.CloseConnection()
	if err != nil {
		log.Panicf("closing producer error: %s", err.Error())
	}

	log.Println("graceful shutdown complete")
}
