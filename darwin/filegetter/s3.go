package filegetter

import (
	"context"
	"io"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3 struct {
	ctx      context.Context
	log      *slog.Logger
	s3Client *s3.Client
	bucket   string
	prefix   string
}

func NewS3(ctx context.Context, log *slog.Logger, s3Client *s3.Client, bucket string, prefix string) S3 {
	return S3{
		ctx:      ctx,
		log:      log,
		s3Client: s3Client,
		bucket:   bucket,
		prefix:   prefix,
	}
}

func (c S3) Get(name string) ([]byte, error) {
	filepath := c.prefix + name
	object, err := c.s3Client.GetObject(c.ctx, &s3.GetObjectInput{
		Bucket: &c.bucket,
		Key:    &filepath,
	})
	if err != nil {
		return nil, err
	}
	return io.ReadAll(object.Body)
}
