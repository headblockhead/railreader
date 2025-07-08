package darwin

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

type Connection struct {
	log               *slog.Logger
	connectionContext context.Context
	fetcherContext    context.Context
	reader            *kafka.Reader
}

type MessageCapsule struct {
	Destination struct {
		Name            string `json:"name"`
		DestinationType string `json:"destinationType"`
	} `json:"destination"`
	MessageID    string `json:"messageID"`
	Priority     int    `json:"priority"`
	Redelivered  bool   `json:"redelivered"`
	MessageType  string `json:"messageType"`
	DeliveryMode int    `json:"deliveryMode"`
	Bytes        string `json:"bytes"`
	ReplyTo      string `json:"replyTo"`
	Expiration   int64  `json:"expiration"`
}

type PushPortMessage struct {
	XMLName          xml.Name  `xml:"Pport"`
	Timestamp        time.Time `xml:"ts,attr"`
	Version          string    `xml:"version,attr"`
	UpdateResponse   *Response `xml:"uR"`
	SnapshotResponse *Response `xml:"sR"`
}

type Response struct {
	UpdateOrigin  string `xml:"updateOrigin,attr"`
	RequestSource string `xml:"requestSource,attr"`
	RequestID     string `xml:"requestID,attr"`

	Schedules                          []ScheduleInformation                `xml:"schedule"`
	Deactivations                      []DeactivationInformation            `xml:"deactivated"`
	Associations                       []AssociationInformation             `xml:"association"`
	TrainFormations                    []TrainFormationInformation          `xml:"scheduleFormations"`
	ActualAndForecasts                 []ActualAndForecastInformation       `xml:"TS"`
	TrainLoadings                      []TrainLoadingInformation            `xml:"formationLoadings"`
	TableSuppressionAndStationMessages []TableSuppressionAndStationMessages `xml:"OW"`
	TrainOrders                        []TrainOrderInformation              `xml:"trainOrder"`
	TrainAlertMessages                 []TrainAlertMessages                 `xml:"trainAlert"`
	TrackingIDChanges                  []TrackingIDChanges                  `xml:"trackingID"`
	Alarms                             []Alarms                             `xml:"alarm"`
}

type ScheduleInformation struct {
	// RID stands for 'Real-Time Train Information Identity' - the unique ID for this specific train running this route, now.
	RID string `xml:"rid,attr"`
	// ID for this route at this time of day
	UID           string `xml:"uid,attr"`
	TrainID       string `xml:"trainID,attr"`
	RSID          string `xml:"rsid,attr"`
	SSD           string `xml:"ssd,attr"`
	TOC           string `xml:"toc,attr"`
	Status        string `xml:"status,attr"`
	TrainCategory string `xml:"trainCat,attr"`
}

type DeactivationInformation struct {
}

type AssociationInformation struct {
}

type TrainFormationInformation struct {
}

type ActualAndForecastInformation struct {
}

type TrainLoadingInformation struct {
}

type TableSuppressionAndStationMessages struct {
}

type TrainOrderInformation struct {
}

type TrainAlertMessages struct {
}

type TrackingIDChanges struct {
}

type Alarms struct {
}

func NewConnection(connectionContext context.Context, fetcherContext context.Context, log *slog.Logger, bootstrapServer string, groupID string, username string, password string) *Connection {
	return &Connection{
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

func (dc *Connection) Close() error {
	dc.log.Info("closing connection...")
	return dc.reader.Close()
}

func (dc *Connection) FetchKafkaMessage() (msg kafka.Message, err error) {
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

func (dc *Connection) ProcessKafkaMessage(msg kafka.Message) error {
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
	var capsule MessageCapsule
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

func (dc *Connection) ProcessMessageCapsule(msg MessageCapsule) error {
	log := dc.log.With(slog.String("messageID", string(msg.MessageID)))

	os.WriteFile(filepath.Join("capture", msg.MessageID+".xml"), []byte(msg.Bytes), 0644)

	var pport PushPortMessage
	if err := xml.Unmarshal([]byte(msg.Bytes), &pport); err != nil {
		return fmt.Errorf("failed to unmarshal message XML: %w", err)
	}
	log.Debug("unmarshalled", slog.String("ts", pport.Timestamp.Format(time.Stamp)))

	// TODO: check common fields are always as we expect

	return nil
}
