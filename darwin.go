package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
)

type DarwinConnection struct {
	log               *slog.Logger
	connectionContext context.Context
	fetcherContext    context.Context
	reader            *kafka.Reader
}

type DarwinMessageCapsuleDestination struct {
	Name            string `json:"name"`
	DestinationType string `json:"destinationType"`
}

type DarwinMessageCapsule struct {
	Destination  DarwinMessageCapsuleDestination `json:"destination"`
	MessageID    string                          `json:"messageID"`
	Priority     int                             `json:"priority"`
	Redelivered  bool                            `json:"redelivered"`
	MessageType  string                          `json:"messageType"`
	DeliveryMode int                             `json:"deliveryMode"`
	Bytes        string                          `json:"bytes"`
	ReplyTo      string                          `json:"replyTo"`
	Expiration   int64                           `json:"expiration"`
}

type DarwinPport struct {
	XMLName          xml.Name        `xml:"Pport"`
	Timestamp        time.Time       `xml:"ts,attr"`
	Version          string          `xml:"version,attr"`
	UpdateResponse   *DarwinResponse `xml:"uR"`
	SnapshotResponse *DarwinResponse `xml:"sR"`
}

type DarwinResponse struct {
	UpdateOrigin  string `xml:"updateOrigin,attr"`
	RequestSource string `xml:"requestSource,attr"`
	RequestID     string `xml:"requestID,attr"`
}

func NewDarwinConnection(connectionContext context.Context, fetcherContext context.Context, log *slog.Logger, bootstrapServer string, groupID string, username string, password string) *DarwinConnection {
	return &DarwinConnection{
		log:               log,
		connectionContext: connectionContext,
		fetcherContext:    fetcherContext,
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: []string{bootstrapServer},
			GroupID: groupID,
			Topic:   "prod-1010-Darwin-Train-Information-Push-Port-IIII2_0-XML",
			Dialer: &kafka.Dialer{
				Timeout:   10 * time.Second,
				DualStack: true,
				SASLMechanism: plain.Mechanism{
					Username: username,
					Password: password,
				},
				TLS: &tls.Config{},
			},
		}),
	}
}

func (dc *DarwinConnection) Close() error {
	dc.log.Info("closing connection...")
	return dc.reader.Close()
}

func (dc *DarwinConnection) FetchKafkaMessage() (msg kafka.Message, err error) {
	dc.log.Debug("waiting for a Kafka message to fetch...")
	if err := dc.fetcherContext.Err(); err != nil {
		return msg, fmt.Errorf("context error: %w", err)
	}
	msg, err = dc.reader.FetchMessage(dc.fetcherContext)
	if err != nil {
		return msg, fmt.Errorf("failed to fetch kafka message: %w", err)
	}
	var key struct {
		MessageID string `json:"messageID"`
	}
	if err := json.Unmarshal(msg.Key, &key); err != nil {
		return msg, fmt.Errorf("failed to unmarshal kafka message key: %w", err)
	}
	dc.log.Debug("recieved Kafka message", slog.String("messageID", key.MessageID))
	return msg, nil
}

func (dc *DarwinConnection) ProcessKafkaMessage(msg kafka.Message) error {
	if err := dc.connectionContext.Err(); err != nil {
		return fmt.Errorf("context error: %w", err)
	}

	var key struct {
		MessageID string `json:"messageID"`
	}
	if err := json.Unmarshal(msg.Key, &key); err != nil {
		return fmt.Errorf("failed to unmarshal kafka message key: %w", err)
	}

	log := dc.log.With(slog.String("messageID", string(key.MessageID)))

	log.Debug("unmarshaling Kafka message...")
	var capsule DarwinMessageCapsule
	if err := json.Unmarshal(msg.Value, &capsule); err != nil {
		return fmt.Errorf("failed to unmarshal kafka message: %w", err)
	}
	log.Debug("unmarshaled Kafka message")

	// TODO: check common fields are always as we expect

	if err := dc.connectionContext.Err(); err != nil {
		return fmt.Errorf("context error: %w", err)
	}

	if err := dc.ProcessMessageCapsule(capsule); err != nil {
		return fmt.Errorf("failed to process message capsule: %w", err)
	}

	if err := dc.reader.CommitMessages(dc.connectionContext, msg); err != nil {
		return fmt.Errorf("failed to commit message: %w", err)
	}

	log.Debug("processed a message")

	return nil
}

func (dc *DarwinConnection) ProcessMessageCapsule(msg DarwinMessageCapsule) error {
	log := dc.log.With(slog.String("messageID", string(msg.MessageID)))

	os.WriteFile(filepath.Join("capture", msg.MessageID+".xml"), []byte(msg.Bytes), 0644)

	var pport DarwinPport
	if err := xml.Unmarshal([]byte(msg.Bytes), &pport); err != nil {
		return fmt.Errorf("failed to unmarshal message XML: %w", err)
	}
	log.Debug("unmarshalled", slog.String("ts", pport.Timestamp.Format(time.Stamp)))

	// TODO: check common fields are always as we expect

	return nil
}
