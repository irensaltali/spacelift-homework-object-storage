package storage

import (
	"fmt"
	"sync"

	"github.com/irensaltali/object-storage-gateway/internal/discovery"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinioClientManager manages connections to Minio instances.
type MinioClientManager struct {
	clients   map[string]*minio.Client
	instances map[string]discovery.MinioInstance
	mu        sync.RWMutex
}

// NewMinioClientManager creates a new Minio client manager.
func NewMinioClientManager() *MinioClientManager {
	return &MinioClientManager{
		clients:   make(map[string]*minio.Client),
		instances: make(map[string]discovery.MinioInstance),
	}
}

// UpdateInstances updates the set of known Minio instances and creates clients.
func (mcm *MinioClientManager) UpdateInstances(instances []discovery.MinioInstance) error {
	mcm.mu.Lock()
	defer mcm.mu.Unlock()

	// Close clients for removed instances
	for id := range mcm.clients {
		found := false
		for _, inst := range instances {
			if inst.ID == id {
				found = true
				break
			}
		}
		if !found {
			delete(mcm.clients, id)
		}
	}

	// Create clients for new instances
	for _, inst := range instances {
		if _, exists := mcm.clients[inst.ID]; !exists {
			client, err := mcm.createClient(inst)
			if err != nil {
				return fmt.Errorf("failed to create client for instance %s: %w", inst.ID, err)
			}
			mcm.clients[inst.ID] = client
		}

		mcm.instances[inst.ID] = inst
	}

	return nil
}

// createClient creates a Minio client for the given instance.
func (mcm *MinioClientManager) createClient(inst discovery.MinioInstance) (*minio.Client, error) {
	endpoint := fmt.Sprintf("%s:%s", inst.Host, inst.Port)

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(inst.AccessKey, inst.SecretKey, ""),
		Secure: false,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	return client, nil
}

// GetClient returns the Minio client for the given instance ID.
func (mcm *MinioClientManager) GetClient(instanceID string) (*minio.Client, error) {
	mcm.mu.RLock()
	defer mcm.mu.RUnlock()

	client, exists := mcm.clients[instanceID]
	if !exists {
		return nil, fmt.Errorf("no client found for instance %s", instanceID)
	}

	return client, nil
}

// GetInstance returns the instance details for the given instance ID.
func (mcm *MinioClientManager) GetInstance(instanceID string) (discovery.MinioInstance, error) {
	mcm.mu.RLock()
	defer mcm.mu.RUnlock()

	inst, exists := mcm.instances[instanceID]
	if !exists {
		return discovery.MinioInstance{}, fmt.Errorf("no instance found with id %s", instanceID)
	}

	return inst, nil
}

// Close closes all Minio client connections.
func (mcm *MinioClientManager) Close() error {
	mcm.mu.Lock()
	defer mcm.mu.Unlock()

	for _, client := range mcm.clients {
		// Minio client doesn't have explicit close, but we clear references
		_ = client
	}

	mcm.clients = make(map[string]*minio.Client)
	mcm.instances = make(map[string]discovery.MinioInstance)

	return nil
}
