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

// MessageCapsule is the raw JSON structure as received from the Rail Data Marketplace's Kafka topic.
// It contains a ridiculous amount of completely useless data and is practically fully undocumented, so I ignore everything but the message data inside, and the message's ID.
type MessageCapsule struct {
	MessageID string `json:"messageID"`
	Bytes     string `json:"bytes"`
}

// PushPortMessage is the root node of Darwin messages.
type PushPortMessage struct {
	Timestamp time.Time `xml:"ts,attr"`
	Version   string    `xml:"version,attr"`

	UpdateResponse   *Response `xml:"uR"`
	SnapshotResponse *Response `xml:"sR"`
}

// Response to a request made to update Darwin's data, by broadcasting the new state(s) of the data to all subscribers.
type Response struct {
	// UpdateOrigin is optionally provided by the requestor to indicate which system the update originated from.
	UpdateOrigin string `xml:"updateOrigin,attr"`
	// RequestSource is optionally provided by the requestor to indicate who they are.
	RequestSource string `xml:"requestSource,attr"`
	// RequestID is optionally provided by the requestor to identify their request.
	RequestID string `xml:"requestID,attr"`

	// 0 or more of any of these updated elements can be present in a response.
	// This includes 0 of all, which is a valid response.

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

type TransportType string

/*func (tt TransportType) String() string {*/
/*return map[TransportType]string{*/
/*PassengerAndParcelTrain: "Passenger and Parcel Train",*/
/*}[tt]*/
/*}*/

const (
	PassengerAndParcelTrain     TransportType = "P"
	Bus                         TransportType = "B"
	Ship                        TransportType = "S"
	Trip                        TransportType = "T"
	Freight                     TransportType = "F"
	PassengerAndParcelShortTerm TransportType = "1"
	BusShortTerm                TransportType = "5"
	ShipShortTerm               TransportType = "4"
	TripShortTerm               TransportType = "3"
	FreightShortTerm            TransportType = "2"
)

type ScheduleInformation struct {
	// RID is the unique 16-character ID for this specific train, running this schedule, at this time.
	RID string `xml:"rid,attr"`
	// UID is (despite the name) a non-unique 6-character ID for this route at this time of day.
	UID string `xml:"uid,attr"`
	// TrainID is the 4-character headcode of the train, with the format:
	// [0-9][A-Z][0-9][0-9]
	TrainID string `xml:"trainID,attr"`
	// RSID is the optionally provided Retail Service ID, either as an:
	// 8 character "portion identifier" (including a leading TOC code),
	// or a 6 character "base identifier" (without a TOC code).
	RSID string `xml:"rsid,attr"`
	// SSD is the scheduled start date of the train, in YYYY-MM-DD format.
	SSD string `xml:"ssd,attr"`
	// TOC is the Rail Delivery Group's 2-character code for the train operating company.
	TOC string `xml:"toc,attr"`
	// Status is the 1-character code for the type of transport.
	// If not provided, it defaults to P (Passenger and Parcel Train).
	Status TransportType `xml:"status,attr"`
	// TrainCategory is a 2-character code for the type of train.
	// Values that indicate a passenger service are:
	// OL, OO, OW, XC, XD, XI, XR, XX, XZ.
	// If not provided, it defaults to OO.
	TrainCategory string `xml:"trainCat,attr"`
	// IsPassengerService is true if not provided. This can be set to false based on the value of the TrainCategory.
	IsPassengerService bool `xml:"isPassengerSvc,attr"`
	// IsActive is only present in snapshots, used to indicate a service has been deactivated by a DeactivationInformation element.
	IsActive bool `xml:"isActive,attr"`
	// Deleted defaults to false. If true, do not use or display this schedule.
	Deleted bool `xml:"deleted,attr"`
	// IsCharter defaults to false.
	IsCharter bool `xml:"isCharter,attr"`
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
	dc.log.Debug("received Kafka message", slog.String("messageID", key.MessageID))
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
	var c MessageCapsule
	if err := json.Unmarshal(msg.Value, &c); err != nil {
		return fmt.Errorf("failed to unmarshal kafka message: %w", err)
	}
	log.Debug("unmarshaled Kafka message")

	if err := dc.connectionContext.Err(); err != nil {
		return fmt.Errorf("context error: %w", err)
	}

	if err := dc.ProcessMessageCapsule(c); err != nil {
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
