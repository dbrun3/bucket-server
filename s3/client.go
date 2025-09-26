package s3

import (
	"context"
	"errors"
	"io"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

var (
	ErrNoBucket = errors.New("NoBucket")
	ErrNoKey    = errors.New("NoKey")
)

type Client struct {
	s3Client *s3.Client
}

func NewClient() *Client {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	client := s3.NewFromConfig(cfg)
	return &Client{
		s3Client: client,
	}
}

func (c *Client) DownloadFile(ctx context.Context, bucketName string, objectKey string) ([]byte, error) {
	objectKey = strings.TrimLeft(objectKey, "/")
	if objectKey == "" {
		return nil, ErrNoKey
	}
	result, err := c.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		var noBucket *types.NoSuchBucket
		if errors.As(err, &noBucket) {
			return nil, ErrNoBucket
		}
		var noKey *types.NoSuchKey
		if errors.As(err, &noKey) {
			return nil, ErrNoKey
		}
		log.Printf("Couldn't get object %v:%v. Here's why: %v\n", bucketName, objectKey, err)
		return nil, err
	}
	defer result.Body.Close()

	body, err := io.ReadAll(result.Body)
	if err != nil {
		log.Printf("Couldn't read object body from %v. Here's why: %v\n", objectKey, err)
	}
	return body, nil
}
