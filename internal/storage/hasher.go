package storage

import (
	"fmt"
	"hash/crc32"
)

// ConsistentHasher provides deterministic mapping of IDs to instances.
type ConsistentHasher struct {
	instances []string
}

// NewConsistentHasher creates a new hasher for the given instances.
func NewConsistentHasher(instances []string) (*ConsistentHasher, error) {
	if len(instances) == 0 {
		return nil, fmt.Errorf("at least one instance is required")
	}

	return &ConsistentHasher{
		instances: instances,
	}, nil
}

// SelectInstance returns the instance ID for a given object ID.
// The same object ID always maps to the same instance.
func (ch *ConsistentHasher) SelectInstance(objectID string) (string, error) {
	if objectID == "" {
		return "", fmt.Errorf("object id cannot be empty")
	}

	if len(ch.instances) == 0 {
		return "", fmt.Errorf("no instances available")
	}

	// Use CRC32 for deterministic hashing
	hash := crc32.ChecksumIEEE([]byte(objectID))
	index := hash % uint32(len(ch.instances))

	return ch.instances[int(index)], nil
}

// UpdateInstances updates the list of available instances.
// This should be called when instances are added or removed.
func (ch *ConsistentHasher) UpdateInstances(instances []string) error {
	if len(instances) == 0 {
		return fmt.Errorf("at least one instance is required")
	}

	ch.instances = instances
	return nil
}
