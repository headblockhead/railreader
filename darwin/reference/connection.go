package reference

import (
	"context"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Connection struct {
	ctx      context.Context
	log      *slog.Logger
	s3Client *s3.Client
	bucket   string
	prefix   string
}

func NewConnection(ctx context.Context, log *slog.Logger, s3Client *s3.Client, bucket string, prefix string) Connection {
	return Connection{
		ctx:      ctx,
		log:      log,
		s3Client: s3Client,
		bucket:   bucket,
		prefix:   prefix,
	}
}

func (c *Connection) Get(filepath string) ([]byte, error) {
	key := c.prefix + filepath
	object, err := c.s3Client.GetObject(c.ctx, &s3.GetObjectInput{
		Bucket: &c.bucket,
		Key:    &key,
	})
	if err != nil {
		return nil, err
	}
	var data []byte
	if _, err := object.Body.Read(data); err != nil {
		return nil, err
	}
	return data, nil
}
