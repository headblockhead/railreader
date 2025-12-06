package darwin

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"log/slog"
	"runtime"
	"strings"
	"sync"

	"github.com/headblockhead/railreader/ingesters/darwin/inserter"
	"github.com/headblockhead/railreader/ingesters/darwin/unmarshaller"
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
	newCtx, cancel := context.WithCancel(ctx)
	i := &Ingester{
		ctx:    newCtx,
		cancel: cancel,
		log:    log,
		reader: reader,
		dbpool: dbpool,
		fs:     fs,
	}
	err := i.importFiles()
	if err != nil {
		return nil, err
	}
	return i, nil
}

func (i *Ingester) Close() error {
	i.cancel()
	i.dbpool.Close()
	return i.reader.Close()
}

// Grab the latest timetable and/or reference files from the SFTP directory.
func (i *Ingester) importFiles() error {
	files, err := i.fs.ReadDir("PPTimetable") // fs.ReadDir returns entries sorted in alphabetical order (oldest to newest)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return errors.New("no files found in SFTP directory")
	}

	var referenceFiles chan string = make(chan string, 64)
	var timetableFiles chan string = make(chan string, 64)

	// Load reference and timetable files into a queue.
	u1, err := inserter.NewUnitOfWork(i.ctx, i.log, i.dbpool, i.fs, nil, nil)
	if err != nil {
		return err
	}
	defer u1.Rollback()

	var loaders sync.WaitGroup
	loaders.Go(func() {
		for _, fileEntry := range files {
			if fileEntry.IsDir() {
				continue
			}
			if strings.HasSuffix(fileEntry.Name(), unmarshaller.ExpectedReferenceFileSuffix) {
				// Encountered a reference file: check if it's already been imported.
				imported, err := u1.ReferenceFileAlreadyImported(fileEntry.Name())
				if err != nil {
					i.log.Error("failed to check if reference file already imported", slog.String("filename", fileEntry.Name()), slog.Any("error", err))
					return
				}
				if !imported {
					// New, add to the queue.
					referenceFiles <- fileEntry.Name()
					i.log.Debug("queued new reference file", slog.String("filename", fileEntry.Name()))
				} else {
					i.log.Debug("skipped already imported reference file", slog.String("filename", fileEntry.Name()))
				}
				continue
			}
			if strings.HasSuffix(fileEntry.Name(), unmarshaller.ExpectedTimetableFileSuffix) {
				// Encountered a timetable file: check if it's already been imported.
				imported, err := u1.TimetableAlreadyImported(fileEntry.Name())
				if err != nil {
					i.log.Error("failed to check if timetable file already imported", slog.String("filename", fileEntry.Name()), slog.Any("error", err))
					return
				}
				if !imported {
					// New, add to the queue.
					timetableFiles <- fileEntry.Name()
					i.log.Debug("queued new timetable file", slog.String("filename", fileEntry.Name()))
				} else {
					i.log.Debug("skipped already imported timetable file", slog.String("filename", fileEntry.Name()))
				}
				continue
			}
			i.log.Debug("skipped unknown file", slog.String("filename", fileEntry.Name()))
		}
		i.log.Debug("finished scanning SFTP directory for new files")
		close(referenceFiles)
		close(timetableFiles)
	})

	for range runtime.NumCPU() {
		loaders.Go(func() {
			// Process new reference files from the queue.
			for fileEntry := range referenceFiles {
				i.log.Debug("taken reference file from the queue", slog.String("filename", fileEntry))
				file, err := i.fs.Open("PPTimetable/" + fileEntry)
				if err != nil {
					i.log.Error("failed to open reference file", slog.String("filename", fileEntry), slog.Any("error", err))
					return
				}
				u2, err := inserter.NewUnitOfWork(i.ctx, i.log, i.dbpool, i.fs, nil, nil)
				if err != nil {
					i.log.Error("failed to create unit of work", slog.String("filename", fileEntry), slog.Any("error", err))
					return
				}
				err = u2.LoadReferenceFile(file)
				if err != nil {
					i.log.Error("failed to load reference file", slog.String("filename", fileEntry), slog.Any("error", err))
					_ = u2.Rollback()
					return
				}
				err = u2.Commit()
				if err != nil {
					i.log.Error("failed to commit reference file", slog.String("filename", fileEntry), slog.Any("error", err))
					_ = u2.Rollback()
					return
				}
				i.log.Info("loaded reference file", slog.String("filename", fileEntry))
			}
		})
		loaders.Go(func() {
			// Process new timetable files from the queue.
			for fileEntry := range timetableFiles {
				i.log.Debug("taken timetable file from the queue", slog.String("filename", fileEntry))
				file, err := i.fs.Open("PPTimetable/" + fileEntry)
				if err != nil {
					i.log.Error("failed to open timetable file", slog.String("filename", fileEntry), slog.Any("error", err))
					return
				}
				u2, err := inserter.NewUnitOfWork(i.ctx, i.log, i.dbpool, i.fs, nil, nil)
				if err != nil {
					i.log.Error("failed to create unit of work", slog.String("filename", fileEntry), slog.Any("error", err))
					return
				}
				err = u2.LoadTimetableFile(file)
				if err != nil {
					i.log.Error("failed to load timetable file", slog.String("filename", fileEntry), slog.Any("error", err))
					_ = u2.Rollback()
					return
				}
				err = u2.Commit()
				if err != nil {
					i.log.Error("failed to commit timetable file", slog.String("filename", fileEntry), slog.Any("error", err))
					_ = u2.Rollback()
					return
				}
				i.log.Info("loaded timetable file", slog.String("filename", fileEntry))
			}
		})
	}
	loaders.Wait()
	return nil
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
	u1, err := inserter.NewUnitOfWork(i.ctx, messageLog, i.dbpool, i.fs, &capsule.MessageID, nil)
	if err != nil {
		return err
	}
	err = u1.InsertMessageXMLRecord(inserter.MessageXMLRecord{
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
	u2, err := inserter.NewUnitOfWork(i.ctx, messageLog, i.dbpool, i.fs, &capsule.MessageID, nil)
	if err != nil {
		return err
	}
	err = u2.InsertPushPortMessage(*pport)
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
