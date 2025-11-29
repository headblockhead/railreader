package darwin

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"log/slog"
	"strings"

	"github.com/headblockhead/railreader/darwin/interpreter"
	"github.com/headblockhead/railreader/darwin/unmarshaller"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"
)

// Ingester implements the interface railreader.Ingester[kafka.Message].
type Ingester struct {
	ctx    context.Context
	cancel context.CancelFunc
	log    *slog.Logger
	reader *kafka.Reader
	dbpool *pgxpool.Pool
	fs     fs.ReadDirFS
}

func NewIngester(ctx context.Context, log *slog.Logger, reader *kafka.Reader, dbpool *pgxpool.Pool, fs fs.ReadDirFS) (*Ingester, error) {

	// Grab the latest timetable and/or reference files from the SFTP directory.
	files, err := fs.ReadDir("PPTimetable") // fs.ReadDir returns entries in sorted order
	if err != nil {
		return nil, err
	}

	var foundReferenceFile bool
	// Search for the most recent reference file:
	for _, fileEntry := range files {
		if fileEntry.IsDir() || !strings.HasSuffix(fileEntry.Name(), unmarshaller.ExpectedReferenceFileSuffix) {
			continue
		}

		file, err := fs.Open("./PPTimetable/" + fileEntry.Name())
		if err != nil {
			return nil, err
		}

		ur, err := interpreter.NewUnitOfWork(ctx, log, dbpool, fs, nil, nil)
		if err != nil {
			return nil, err
		}
		err = ur.InterpretReferenceFile(file)
		if err != nil {
			_ = ur.Rollback()
			return nil, err
		}
		err = ur.Commit()
		if err != nil {
			_ = ur.Rollback()
			return nil, err
		}

		foundReferenceFile = true
		break
	}

	if !foundReferenceFile {
		return nil, errors.New("no reference file found in SFTP directory")
	}

	// Search for the most recent timetable file:
	/* for _, fileEntry := range files {*/
	/*if fileEntry.IsDir() || !strings.HasSuffix(fileEntry.Name(), unmarshaller.ExpectedTimetableFileSuffix) {*/
	/*continue*/
	/*}*/

	/*file, err := fs.Open("./PPTimetable/" + fileEntry.Name())*/
	/*if err != nil {*/
	/*return nil, err*/
	/*}*/

	/*ur, err := interpreter.NewUnitOfWork(ctx, log, dbpool, fs, nil, nil)*/
	/*if err != nil {*/
	/*return nil, err*/
	/*}*/
	/*err = ur.InterpretTimetableFile(file)*/
	/*if err != nil {*/
	/*_ = ur.Rollback()*/
	/*return nil, err*/
	/*}*/
	/*err = ur.Commit()*/
	/*if err != nil {*/
	/*_ = ur.Rollback()*/
	/*return nil, err*/
	/*}*/

	/*break*/
	/*}*/

	newCtx, cancel := context.WithCancel(ctx)
	return &Ingester{
		ctx:    newCtx,
		cancel: cancel,
		log:    log,
		reader: reader,
		dbpool: dbpool,
		fs:     fs,
	}, nil
}

func (i *Ingester) Close() error {
	i.cancel()
	i.dbpool.Close()
	return i.reader.Close()
}

// Fetch blocks until a message is available, or the provided context is cancelled.
func (i *Ingester) Fetch(ctx context.Context) (kafka.Message, error) {
	msg, err := i.reader.FetchMessage(ctx)
	if err != nil {
		return msg, err
	}
	i.log.Info("fetched message", slog.Int64("offset", msg.Offset))
	return msg, nil
}

// messageCapsule is the raw JSON structure as received from the Rail Data Marketplace's Kafka topic.
// The JSON itself has a lot of useless data, so I cherry-pick out what I want.
type messageCapsule struct {
	MessageID  string `json:"messageID"`
	Properties struct {
		PushPortSequence struct {
			String string `json:"string"`
		} `json:"PushPortSequence"`
	} `json:"properties"`
	XML string `json:"bytes"`
}

func (i *Ingester) ProcessAndCommit(msg kafka.Message) error {
	// Unmarshal the message capsule from JSON to extract its fields.
	var capsule messageCapsule
	err := json.Unmarshal(msg.Value, &capsule)
	if err != nil {
		return err
	}
	messageLog := i.log.With(slog.String("messageID", capsule.MessageID))

	// Unit of work 1: Insert the message XML record.
	u1, err := interpreter.NewUnitOfWork(i.ctx, messageLog, i.dbpool, i.fs, &capsule.MessageID, nil)
	if err != nil {
		return err
	}
	err = u1.InsertMessageXMLRecord(interpreter.MessageXMLRecord{
		ID:            capsule.MessageID,
		KafkaOffset:   msg.Offset,
		PPortSequence: capsule.Properties.PushPortSequence.String,
		XML:           capsule.XML,
	})
	if err != nil {
		_ = u1.Rollback()
		return err
	}
	err = u1.Commit()
	if err != nil {
		_ = u1.Rollback()
		return err
	}

	// Unmarshal the whole PushPort message's XML.
	pport, err := unmarshaller.NewPushPortMessage(capsule.XML)
	if err != nil {
		return err
	}

	// Unit of work 2: Insert the message's data into the various tables.
	u2, err := interpreter.NewUnitOfWork(i.ctx, messageLog, i.dbpool, i.fs, &capsule.MessageID, nil)
	if err != nil {
		return err
	}
	err = u2.InterpretPushPortMessage(pport)
	if err != nil {
		_ = u2.Rollback()
		return err
	}
	err = u2.Commit()
	if err != nil {
		_ = u2.Rollback()
		return err
	}

	// Mark the message as committed in Kafka.
	err = i.reader.CommitMessages(i.ctx, msg)
	if err != nil {
		return err
	}
	i.log.Info("committed message", slog.Int64("offset", msg.Offset))
	return nil
}
