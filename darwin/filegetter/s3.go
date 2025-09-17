package filegetter

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"

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

func (c S3) Get(name string) (io.ReadCloser, error) {
	filepath := c.prefix + name
	object, err := c.s3Client.GetObject(c.ctx, &s3.GetObjectInput{
		Bucket: &c.bucket,
		Key:    &filepath,
	})
	if err != nil {
		return nil, err
	}
	return object.Body, nil
}

func (c S3) FindNewestWithSuffix(suffix string) (string, error) {
	paginator := s3.NewListObjectsV2Paginator(c.s3Client, &s3.ListObjectsV2Input{
		Bucket: &c.bucket,
		Prefix: &c.prefix,
	})

	var latestPath string
	var latestModTime int64

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(c.ctx)
		if err != nil {
			return "", err
		}
		for _, obj := range page.Contents {
			if obj.Key == nil || !strings.HasSuffix(*obj.Key, suffix) {
				continue
			}
			if obj.LastModified == nil {
				c.log.Warn("skipping over object with missing modification time", slog.String("key", *obj.Key))
				continue
			}
			if obj.LastModified.Unix() > latestModTime {
				latestModTime = obj.LastModified.Unix()
				latestPath = *obj.Key
			}
		}
	}
	if latestPath == "" {
		return "", errors.New("no file found with the specified suffix")
	}
	latestPath = strings.TrimPrefix(latestPath, c.prefix)
	return latestPath, nil
}
