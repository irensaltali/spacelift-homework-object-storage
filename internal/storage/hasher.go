package storage

import (
	"fmt"
	"hash/fnv"
	"sort"
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

	sortedInstances := append([]string(nil), instances...)
	sort.Strings(sortedInstances)

	return &ConsistentHasher{
		instances: sortedInstances,
	}, nil
}

// SelectInstance returns the instance ID for a given object ID.
// The same object ID always maps to the same instance.
func (ch *ConsistentHasher) SelectInstance(objectKey string) (string, error) {
	if objectKey == "" {
		return "", fmt.Errorf("object id cannot be empty")
	}

	if len(ch.instances) == 0 {
		return "", fmt.Errorf("no instances available")
	}

	// Rendezvous hashing (highest-random-weight) keeps selection deterministic
	// while avoiding modulo based bucket assignment.
	selected := ""
	var maxScore uint64

	for idx, instance := range ch.instances {
		score := calculateRendezvousScore(objectKey, instance)
		if idx == 0 || score > maxScore || (score == maxScore && instance < selected) {
			selected = instance
			maxScore = score
		}
	}

	return selected, nil
}

// UpdateInstances updates the list of available instances.
// This should be called when instances are added or removed.
func (ch *ConsistentHasher) UpdateInstances(instances []string) error {
	if len(instances) == 0 {
		return fmt.Errorf("at least one instance is required")
	}

	sortedInstances := append([]string(nil), instances...)
	sort.Strings(sortedInstances)
	ch.instances = sortedInstances
	return nil
}

func calculateRendezvousScore(objectKey, instance string) uint64 {
	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte(objectKey))
	_, _ = hasher.Write([]byte{':'})
	_, _ = hasher.Write([]byte(instance))
	return hasher.Sum64()
}
