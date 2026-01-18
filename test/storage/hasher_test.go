package storage

import (
	"testing"

	"github.com/irensaltali/object-storage-gateway/internal/storage"
)

func TestConsistentHasher(t *testing.T) {
	instances := []string{"instance-1", "instance-2", "instance-3"}

	hasher, err := storage.NewConsistentHasher(instances)
	if err != nil {
		t.Fatalf("failed to create hasher: %v", err)
	}

	// Test that same ID always maps to same instance
	objectID := "test-object-123"
	selectedInstance1, err := hasher.SelectInstance(objectID)
	if err != nil {
		t.Fatalf("failed to select instance: %v", err)
	}

	selectedInstance2, err := hasher.SelectInstance(objectID)
	if err != nil {
		t.Fatalf("failed to select instance: %v", err)
	}

	if selectedInstance1 != selectedInstance2 {
		t.Errorf("inconsistent hashing: got %s and %s for same object ID", selectedInstance1, selectedInstance2)
	}

	// Test that different IDs can map to different instances
	objectID2 := "test-object-456"
	selectedInstance3, err := hasher.SelectInstance(objectID2)
	if err != nil {
		t.Fatalf("failed to select instance: %v", err)
	}

	// Note: They might be the same by chance, but we can't assert they're different
	if selectedInstance3 == "" {
		t.Error("selected instance is empty")
	}

	// Test empty ID
	_, err = hasher.SelectInstance("")
	if err == nil {
		t.Error("expected error for empty object ID")
	}

	// Test no instances
	_, err = storage.NewConsistentHasher([]string{})
	if err == nil {
		t.Error("expected error for no instances")
	}
}

func TestConsistentHasherUpdateInstances(t *testing.T) {
	instances := []string{"instance-1", "instance-2"}

	hasher, err := storage.NewConsistentHasher(instances)
	if err != nil {
		t.Fatalf("failed to create hasher: %v", err)
	}

	objectID := "test-object"
	selectedInstance1, _ := hasher.SelectInstance(objectID)

	// Update instances
	newInstances := []string{"instance-1", "instance-2", "instance-3"}
	err = hasher.UpdateInstances(newInstances)
	if err != nil {
		t.Fatalf("failed to update instances: %v", err)
	}

	selectedInstance2, _ := hasher.SelectInstance(objectID)

	// Note: Hashing might be different after adding instances
	// This is expected behavior for consistent hashing
	if selectedInstance2 == "" {
		t.Error("selected instance is empty after update")
	}

	// Verify instance is in the new list
	found := false
	for _, inst := range newInstances {
		if inst == selectedInstance2 {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("selected instance %s not in updated list", selectedInstance2)
	}

	_ = selectedInstance1 // selectedInstance1 is used for comparison logic
}
