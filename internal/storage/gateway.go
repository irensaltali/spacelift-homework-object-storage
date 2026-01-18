package storage

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/irensaltali/object-storage-gateway/internal/discovery"
	"github.com/minio/minio-go/v7"
)

// Gateway provides the main object storage gateway functionality.
type Gateway struct {
	hasher  *ConsistentHasher
	clients *MinioClientManager
}

// NewGateway creates a new object storage gateway.
func NewGateway(instances []discovery.MinioInstance) (*Gateway, error) {
	if len(instances) == 0 {
		return nil, fmt.Errorf("at least one minio instance is required")
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
		hasher:  hasher,
		clients: clients,
	}, nil
}

// PutObject stores an object in the gateway.
func (g *Gateway) PutObject(ctx context.Context, objectKey string, data io.Reader, size int64) error {
	if objectKey == "" {
		return fmt.Errorf("object id cannot be empty")
	}

	if data == nil {
		return fmt.Errorf("data cannot be nil")
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

	bucketName := "objects"

	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		log.Fatalln("Failed to check bucket existence:", err)
	}

	if !exists {
		err = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			log.Fatalln("Failed to create bucket:", err)
		}
	}

	_, err = client.PutObject(ctx, bucketName, objectKey, data, size, minio.PutObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to put object in minio: %w", err)
	}

	return nil
}

// GetObject retrieves an object from the gateway.
func (g *Gateway) GetObject(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	if objectKey == "" {
		return nil, fmt.Errorf("object id cannot be empty")
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
	bucketName := "objects"

	_, err = client.StatObject(ctx, bucketName, objectKey, minio.StatObjectOptions{})

	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		if errResponse.Code == "NoSuchKey" || errResponse.Code == "NoSuchBucket" || errResponse.Code == "NoSuchObject" {
			return nil, fmt.Errorf("object not found or error reading: %w", err)
		}
		log.Printf("GET /object/%s - error stat object: %v, code: %s", objectKey, errResponse, errResponse.Code)
		return nil, fmt.Errorf("failed to stat object: %w", err)
	}

	object, err := client.GetObject(ctx, bucketName, objectKey, minio.GetObjectOptions{})
	if err != nil {
		object.Close()
		return nil, fmt.Errorf("object not found or error reading: %w", err)
	}

	return object, nil
}

// Close closes the gateway and all connections.
func (g *Gateway) Close() error {
	return g.clients.Close()
}
