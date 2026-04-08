package rustfs

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	EnvEndpoint  = "RUSTFS_ENDPOINT_URL"
	EnvRegion    = "RUSTFS_REGION"
	EnvAccessKey = "RUSTFS_ACCESS_KEY_ID"
	EnvSecretKey = "RUSTFS_SECRET_ACCESS_KEY"
)

var errMissingConfig = errors.New("missing RustFS configuration")

type Config struct {
	EndpointURL     string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	UsePathStyle    bool
}

type Client struct {
	s3 *s3.Client
}

type BucketInfo struct {
	Name string
}

type ObjectInfo struct {
	Key  string
	Size int64
}

func New(ctx context.Context, cfg Config) (*Client, error) {
	if cfg.EndpointURL == "" || cfg.Region == "" || cfg.AccessKeyID == "" || cfg.SecretAccessKey == "" {
		return nil, fmt.Errorf("%w: %s, %s, %s, %s", errMissingConfig, EnvEndpoint, EnvRegion, EnvAccessKey, EnvSecretKey)
	}

	if !cfg.UsePathStyle {
		cfg.UsePathStyle = true
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(
		ctx,
		awsconfig.WithRegion(cfg.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = cfg.UsePathStyle
		o.BaseEndpoint = aws.String(cfg.EndpointURL)
	})

	return &Client{s3: client}, nil
}

func NewFromEnv(ctx context.Context) (*Client, error) {
	loadDotEnv(".env")

	return New(ctx, Config{
		EndpointURL:     os.Getenv(EnvEndpoint),
		Region:          os.Getenv(EnvRegion),
		AccessKeyID:     os.Getenv(EnvAccessKey),
		SecretAccessKey: os.Getenv(EnvSecretKey),
		UsePathStyle:    true,
	})
}

func loadDotEnv(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)
		if key == "" {
			continue
		}

		if os.Getenv(key) == "" {
			_ = os.Setenv(key, value)
		}
	}
}

func (c *Client) ListBuckets(ctx context.Context) ([]BucketInfo, error) {
	resp, err := c.s3.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("list buckets: %w", err)
	}

	out := make([]BucketInfo, 0, len(resp.Buckets))
	for _, bucket := range resp.Buckets {
		out = append(out, BucketInfo{
			Name: aws.ToString(bucket.Name),
		})
	}
	return out, nil
}

func (c *Client) ListObjects(ctx context.Context, bucket string) ([]ObjectInfo, error) {
	resp, err := c.s3.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		return nil, fmt.Errorf("list objects in bucket %q: %w", bucket, err)
	}

	out := make([]ObjectInfo, 0, len(resp.Contents))
	for _, object := range resp.Contents {
		out = append(out, ObjectInfo{
			Key:  aws.ToString(object.Key),
			Size: aws.ToInt64(object.Size),
		})
	}
	return out, nil
}

func (c *Client) ReadObject(ctx context.Context, bucket, key string) ([]byte, error) {
	resp, err := c.s3.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("get object %q from bucket %q: %w", key, bucket, err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read object body for %q from bucket %q: %w", key, bucket, err)
	}
	return data, nil
}

func (c *Client) DownloadObject(ctx context.Context, bucket, key, outputPath string) error {
	resp, err := c.s3.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("get object %q from bucket %q: %w", key, bucket, err)
	}
	defer resp.Body.Close()

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create output file %q: %w", outputPath, err)
	}
	defer file.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
		return fmt.Errorf("write object %q to %q: %w", key, outputPath, err)
	}
	return nil
}
