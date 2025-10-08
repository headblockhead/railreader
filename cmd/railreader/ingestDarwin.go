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
	"github.com/headblockhead/railreader/darwin/interpreter"
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

	if err := loadNewestDarwinFiles(log, db, darwinFileGetter); err != nil {
		return nil, nil, fmt.Errorf("error loading newest darwin files: %w", err)
	}

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

func loadNewestDarwinFiles(log *slog.Logger, db database.Database, fg filegetter.FileGetter) error {
	var timetableFileExension = "_v8.xml.gz"
	var referenceFileExtension = "_ref_v4.xml.gz"
	referencePath, err := fg.FindNewestWithSuffix(referenceFileExtension)
	if err != nil {
		return fmt.Errorf("error getting latest darwin reference file path: %w", err)
	}
	timetablePath, err := fg.FindNewestWithSuffix(timetableFileExension)
	if err != nil {
		return fmt.Errorf("error getting latest darwin timetable file path: %w", err)
	}
	log.Debug("creating a new UnitOfWork for interpreting the data")
	u, err := interpreter.NewUnitOfWork(context.Background(), log, db, fg, nil, nil) // timetableID is set with interpreter.InterpretTimetable
	if err != nil {
		return fmt.Errorf("failed to create a new UnitOfWork: %w", err)
	}
	lastReference, err := u.GetLastReference()
	if err != nil {
		return fmt.Errorf("error getting last darwin reference data: %w", err)
	}
	lastTimetable, err := u.GetLastTimetable()
	if err != nil {
		return fmt.Errorf("error getting last darwin timetable data: %w", err)
	}
	if lastReference == nil || lastReference.Filename != referencePath {
		log.Info("getting darwin reference data", slog.String("path", referencePath))
		err = u.InterpretFromPath(referencePath)
		if err != nil {
			return fmt.Errorf("error getting darwin reference file: %w", err)
		}
	} else {
		log.Info("latest darwin reference file has already been processed, skipping", slog.String("path", referencePath))
	}
	if lastTimetable == nil || lastTimetable.Filename != timetablePath {
		log.Info("getting darwin timetable data", slog.String("path", timetablePath))
		err = u.InterpretFromPath(timetablePath)
		if err != nil {
			return fmt.Errorf("error getting darwin timetable file: %w", err)
		}
	} else {
		log.Info("latest darwin timetable file has already been processed, skipping", slog.String("path", timetablePath))
	}
	log.Debug("committing the UnitOfWork")
	if err := u.Commit(); err != nil {
		_ = u.Rollback()
		return fmt.Errorf("failed to commit UnitOfWork: %w", err)
	}
	log.Debug("interpreted the data")
	return nil
}
