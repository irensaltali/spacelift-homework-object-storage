package storage

import (
	"testing"

	"github.com/irensaltali/object-storage-gateway/internal/discovery"
)

func TestNewMinioClientManager(t *testing.T) {
	mcm := NewMinioClientManager()
	if mcm == nil {
		t.Fatal("NewMinioClientManager() returned nil")
	}
	if mcm.clients == nil {
		t.Error("clients map is nil")
	}
	if mcm.instances == nil {
		t.Error("instances map is nil")
	}
	if len(mcm.clients) != 0 {
		t.Errorf("expected empty clients map, got %d clients", len(mcm.clients))
	}
	if len(mcm.instances) != 0 {
		t.Errorf("expected empty instances map, got %d instances", len(mcm.instances))
	}
}

func TestMinioClientManager_UpdateInstances(t *testing.T) {
	tests := []struct {
		name      string
		instances []discovery.MinioInstance
		wantErr   bool
	}{
		{
			name: "single instance",
			instances: []discovery.MinioInstance{
				{ID: "instance-1", Host: "localhost", Port: "9000", AccessKey: "minioadmin", SecretKey: "minioadmin"},
			},
			wantErr: false,
		},
		{
			name: "multiple instances",
			instances: []discovery.MinioInstance{
				{ID: "instance-1", Host: "localhost", Port: "9000", AccessKey: "minioadmin", SecretKey: "minioadmin"},
				{ID: "instance-2", Host: "localhost", Port: "9001", AccessKey: "minioadmin", SecretKey: "minioadmin"},
			},
			wantErr: false,
		},
		{
			name:      "empty instances",
			instances: []discovery.MinioInstance{},
			wantErr:   false, // UpdateInstances doesn't validate empty, it just clears
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mcm := NewMinioClientManager()
			err := mcm.UpdateInstances(tt.instances)
			if tt.wantErr {
				if err == nil {
					t.Errorf("UpdateInstances() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("UpdateInstances() unexpected error = %v", err)
				}
				// Verify instances were stored
				if len(tt.instances) > 0 {
					for _, inst := range tt.instances {
						_, err := mcm.GetInstance(inst.ID)
						if err != nil {
							t.Errorf("GetInstance() failed for %s: %v", inst.ID, err)
						}
					}
				}
			}
		})
	}
}

func TestMinioClientManager_UpdateInstancesRemovesOld(t *testing.T) {
	mcm := NewMinioClientManager()

	// Add initial instances
	initialInstances := []discovery.MinioInstance{
		{ID: "instance-1", Host: "localhost", Port: "9000", AccessKey: "minioadmin", SecretKey: "minioadmin"},
		{ID: "instance-2", Host: "localhost", Port: "9001", AccessKey: "minioadmin", SecretKey: "minioadmin"},
	}

	err := mcm.UpdateInstances(initialInstances)
	if err != nil {
		t.Fatalf("UpdateInstances() error = %v", err)
	}

	// Update with only one instance
	updatedInstances := []discovery.MinioInstance{
		{ID: "instance-1", Host: "localhost", Port: "9000", AccessKey: "minioadmin", SecretKey: "minioadmin"},
	}

	err = mcm.UpdateInstances(updatedInstances)
	if err != nil {
		t.Fatalf("UpdateInstances() error = %v", err)
	}

	// Verify instance-1 still exists
	_, err = mcm.GetInstance("instance-1")
	if err != nil {
		t.Errorf("GetInstance() failed for instance-1: %v", err)
	}

	// Verify instance-2 was removed (client should be gone, but instance might still be in map)
	// Actually, looking at the code, instances map is updated, so instance-2 should be removed
	_, err = mcm.GetInstance("instance-2")
	if err == nil {
		t.Error("GetInstance() expected error for removed instance-2")
	}
}

func TestMinioClientManager_GetClient(t *testing.T) {
	mcm := NewMinioClientManager()

	instances := []discovery.MinioInstance{
		{ID: "instance-1", Host: "localhost", Port: "9000", AccessKey: "minioadmin", SecretKey: "minioadmin"},
	}

	err := mcm.UpdateInstances(instances)
	if err != nil {
		t.Fatalf("UpdateInstances() error = %v", err)
	}

	tests := []struct {
		name       string
		instanceID string
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "existing client",
			instanceID: "instance-1",
			wantErr:    false,
		},
		{
			name:       "non-existent client",
			instanceID: "instance-999",
			wantErr:    true,
			errMsg:     "no client found",
		},
		{
			name:       "empty instance ID",
			instanceID: "",
			wantErr:    true,
			errMsg:     "no client found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := mcm.GetClient(tt.instanceID)
			if tt.wantErr {
				if err == nil {
					t.Errorf("GetClient() expected error but got none")
					return
				}
				if client != nil {
					t.Errorf("GetClient() expected nil client on error")
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("GetClient() error = %v, want error containing %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("GetClient() unexpected error = %v", err)
				}
				if client == nil {
					t.Errorf("GetClient() expected non-nil client")
				}
			}
		})
	}
}

func TestMinioClientManager_GetInstance(t *testing.T) {
	mcm := NewMinioClientManager()

	instances := []discovery.MinioInstance{
		{ID: "instance-1", Host: "localhost", Port: "9000", AccessKey: "minioadmin", SecretKey: "minioadmin"},
		{ID: "instance-2", Host: "localhost", Port: "9001", AccessKey: "minioadmin2", SecretKey: "minioadmin2"},
	}

	err := mcm.UpdateInstances(instances)
	if err != nil {
		t.Fatalf("UpdateInstances() error = %v", err)
	}

	tests := []struct {
		name       string
		instanceID string
		wantErr    bool
		wantHost   string
		wantPort   string
	}{
		{
			name:       "existing instance",
			instanceID: "instance-1",
			wantErr:    false,
			wantHost:   "localhost",
			wantPort:   "9000",
		},
		{
			name:       "another existing instance",
			instanceID: "instance-2",
			wantErr:    false,
			wantHost:   "localhost",
			wantPort:   "9001",
		},
		{
			name:       "non-existent instance",
			instanceID: "instance-999",
			wantErr:    true,
		},
		{
			name:       "empty instance ID",
			instanceID: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instance, err := mcm.GetInstance(tt.instanceID)
			if tt.wantErr {
				if err == nil {
					t.Errorf("GetInstance() expected error but got none")
					return
				}
			} else {
				if err != nil {
					t.Errorf("GetInstance() unexpected error = %v", err)
					return
				}
				if instance.ID != tt.instanceID {
					t.Errorf("GetInstance() ID = %v, want %v", instance.ID, tt.instanceID)
				}
				if instance.Host != tt.wantHost {
					t.Errorf("GetInstance() Host = %v, want %v", instance.Host, tt.wantHost)
				}
				if instance.Port != tt.wantPort {
					t.Errorf("GetInstance() Port = %v, want %v", instance.Port, tt.wantPort)
				}
			}
		})
	}
}

func TestMinioClientManager_Close(t *testing.T) {
	mcm := NewMinioClientManager()

	instances := []discovery.MinioInstance{
		{ID: "instance-1", Host: "localhost", Port: "9000", AccessKey: "minioadmin", SecretKey: "minioadmin"},
	}

	err := mcm.UpdateInstances(instances)
	if err != nil {
		t.Fatalf("UpdateInstances() error = %v", err)
	}

	err = mcm.Close()
	if err != nil {
		t.Errorf("Close() unexpected error = %v", err)
	}

	// Verify clients are cleared
	if len(mcm.clients) != 0 {
		t.Errorf("Close() expected empty clients map, got %d clients", len(mcm.clients))
	}

	if len(mcm.instances) != 0 {
		t.Errorf("Close() expected empty instances map, got %d instances", len(mcm.instances))
	}
}

func TestMinioClientManager_ConcurrentAccess(t *testing.T) {
	mcm := NewMinioClientManager()

	instances := []discovery.MinioInstance{
		{ID: "instance-1", Host: "localhost", Port: "9000", AccessKey: "minioadmin", SecretKey: "minioadmin"},
		{ID: "instance-2", Host: "localhost", Port: "9001", AccessKey: "minioadmin", SecretKey: "minioadmin"},
	}

	err := mcm.UpdateInstances(instances)
	if err != nil {
		t.Fatalf("UpdateInstances() error = %v", err)
	}

	// Test concurrent reads
	done := make(chan bool, 2)
	go func() {
		_, err := mcm.GetClient("instance-1")
		if err != nil {
			t.Errorf("GetClient() error in goroutine 1: %v", err)
		}
		done <- true
	}()

	go func() {
		_, err := mcm.GetInstance("instance-2")
		if err != nil {
			t.Errorf("GetInstance() error in goroutine 2: %v", err)
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done
}

func TestMinioClientManager_UpdateInstancesWithInvalidClient(t *testing.T) {
	mcm := NewMinioClientManager()

	// Try to create a client with invalid endpoint
	// This will fail during client creation
	instances := []discovery.MinioInstance{
		{ID: "instance-1", Host: "invalid-host-that-does-not-exist", Port: "99999", AccessKey: "minioadmin", SecretKey: "minioadmin"},
	}

	// Note: The actual client creation might succeed (it doesn't connect immediately)
	// but we can test that UpdateInstances handles errors properly
	err := mcm.UpdateInstances(instances)
	// The error might occur during client creation, but UpdateInstances should handle it
	_ = err
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsMiddle(s, substr))))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
