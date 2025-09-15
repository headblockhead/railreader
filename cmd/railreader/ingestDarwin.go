package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/headblockhead/railreader/darwin"
	"github.com/headblockhead/railreader/darwin/fetchercommitter"
	"github.com/headblockhead/railreader/darwin/filegetter"
	"github.com/headblockhead/railreader/database"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
)

func (c IngestCommand) newDarwin(log *slog.Logger, db database.Database) (messageFetcherCommitter, messageHandler, error) {
	creds := credentials.NewStaticCredentialsProvider(c.Darwin.S3.AccessKey, c.Darwin.S3.SecretKey, "")
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(c.Darwin.S3.Region), config.WithCredentialsProvider(creds))
	if err != nil {
		return nil, nil, fmt.Errorf("error creating AWS config: %w", err)
	}
	darwinAWSS3Client := s3.NewFromConfig(cfg)
	darwinFileGetter := filegetter.NewS3(context.Background(), log.With(slog.String("package", "filegetter")), darwinAWSS3Client, c.Darwin.S3.Bucket, c.Darwin.S3.Prefix)

	kafkaContext := context.Background()
	darwinKafkaConnection := fetchercommitter.NewKafka(kafkaContext, log.With(slog.String("process", "messagefetchercommitter")), kafka.ReaderConfig{
		Brokers: c.Darwin.Kafka.Brokers,
		GroupID: c.Darwin.Kafka.Group,
		Topic:   c.Darwin.Kafka.Topic,
		Dialer: &kafka.Dialer{
			Timeout:   c.Darwin.Kafka.ConnectionTimeout,
			DualStack: true,
			SASLMechanism: plain.Mechanism{
				Username: c.Darwin.Kafka.Username,
				Password: c.Darwin.Kafka.Password,
			},
			TLS: &tls.Config{},
		},
	})

	messageHandlerContext := context.Background()
	darwinMessageHandler := darwin.NewMessageHandler(messageHandlerContext, log.With(slog.String("source", "darwin.handler")), db, darwinFileGetter)

	return darwinKafkaConnection, darwinMessageHandler, nil
}
