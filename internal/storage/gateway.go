package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/irensaltali/object-storage-gateway/internal/discovery"
	"github.com/minio/minio-go/v7"
)

const defaultBucketName = "objects"

// Gateway provides the main object storage gateway functionality.
type Gateway struct {
	hasher     *ConsistentHasher
	clients    *MinioClientManager
	bucketName string
}

type gatewayConfig struct {
	bucketName string
}

// GatewayOption configures gateway construction.
type GatewayOption func(*gatewayConfig)

// WithBucketName configures the bucket used for object storage.
func WithBucketName(bucketName string) GatewayOption {
	return func(cfg *gatewayConfig) {
		cfg.bucketName = bucketName
	}
}

// NewGateway creates a new object storage gateway.
func NewGateway(instances []discovery.MinioInstance, opts ...GatewayOption) (*Gateway, error) {
	if len(instances) == 0 {
		return nil, fmt.Errorf("at least one minio instance is required")
	}

	cfg := gatewayConfig{
		bucketName: defaultBucketName,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	if strings.TrimSpace(cfg.bucketName) == "" {
		return nil, fmt.Errorf("bucket name cannot be empty")
	}

	// Extract instance IDs for hashing
	instanceIDs := make([]string, len(instances))
	for i, inst := range instances {
		instanceIDs[i] = inst.ID
	}

	hasher, err := NewConsistentHasher(instanceIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to create hasher: %w", err)
	}

	clients := NewMinioClientManager()
	if err := clients.UpdateInstances(instances); err != nil {
		return nil, fmt.Errorf("failed to initialize clients: %w", err)
	}

	return &Gateway{
		hasher:     hasher,
		clients:    clients,
		bucketName: cfg.bucketName,
	}, nil
}

// PutObject stores an object in the gateway.
func (g *Gateway) PutObject(ctx context.Context, objectKey string, data io.Reader, size int64) error {
	if err := ValidateObjectID(objectKey); err != nil {
		return err
	}

	if data == nil {
		return fmt.Errorf("data cannot be nil")
	}
	if size < 0 {
		return fmt.Errorf("size cannot be negative")
	}

	// Select instance based on object ID
	instanceID, err := g.hasher.SelectInstance(objectKey)
	if err != nil {
		return fmt.Errorf("failed to select instance: %w", err)
	}

	// Get the Minio client for the selected instance
	client, err := g.clients.GetClient(instanceID)
	if err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}

	exists, err := client.BucketExists(ctx, g.bucketName)
	if err != nil {
		return fmt.Errorf("failed to check bucket %q existence: %w", g.bucketName, err)
	}

	if !exists {
		err = client.MakeBucket(ctx, g.bucketName, minio.MakeBucketOptions{})
		if err != nil {
			errResp := minio.ToErrorResponse(err)
			if errResp.Code != "BucketAlreadyOwnedByYou" && errResp.Code != "BucketAlreadyExists" {
				return fmt.Errorf("failed to create bucket %q: %w", g.bucketName, err)
			}
		}
	}

	_, err = client.PutObject(ctx, g.bucketName, objectKey, data, size, minio.PutObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to put object in minio: %w", err)
	}

	return nil
}

// GetObject retrieves an object from the gateway.
func (g *Gateway) GetObject(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	if err := ValidateObjectID(objectKey); err != nil {
		return nil, err
	}

	// Select instance based on object ID
	instanceID, err := g.hasher.SelectInstance(objectKey)
	if err != nil {
		return nil, fmt.Errorf("failed to select instance: %w", err)
	}

	// Get the Minio client for the selected instance
	client, err := g.clients.GetClient(instanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	// Retrieve object from Minio
	_, err = client.StatObject(ctx, g.bucketName, objectKey, minio.StatObjectOptions{})

	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		if errResponse.Code == "NoSuchKey" || errResponse.Code == "NoSuchBucket" || errResponse.Code == "NoSuchObject" {
			return nil, fmt.Errorf("%w: %s", ErrObjectNotFound, objectKey)
		}
		log.Printf("GET /object/%s - error stat object: %v, code: %s", objectKey, errResponse, errResponse.Code)
		return nil, fmt.Errorf("failed to stat object: %w", err)
	}

	object, err := client.GetObject(ctx, g.bucketName, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}

	return object, nil
}

// Close closes the gateway and all connections.
func (g *Gateway) Close() error {
	return g.clients.Close()
}
