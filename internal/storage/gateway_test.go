package storage

import (
	"context"
	"strings"
	"testing"

	"github.com/irensaltali/object-storage-gateway/internal/discovery"
)

func TestNewGateway(t *testing.T) {
	tests := []struct {
		name      string
		instances []discovery.MinioInstance
		options   []GatewayOption
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "empty instances",
			instances: []discovery.MinioInstance{},
			wantErr:   true,
			errMsg:    "at least one minio instance is required",
		},
		{
			name: "default bucket name",
			instances: []discovery.MinioInstance{
				{ID: "instance-1", Host: "localhost", Port: "9000", AccessKey: "minioadmin", SecretKey: "minioadmin"},
			},
			wantErr: false,
		},
		{
			name: "custom bucket name",
			instances: []discovery.MinioInstance{
				{ID: "instance-1", Host: "localhost", Port: "9000", AccessKey: "minioadmin", SecretKey: "minioadmin"},
			},
			options: []GatewayOption{WithBucketName("custom-objects")},
			wantErr: false,
		},
		{
			name: "invalid bucket name",
			instances: []discovery.MinioInstance{
				{ID: "instance-1", Host: "localhost", Port: "9000", AccessKey: "minioadmin", SecretKey: "minioadmin"},
			},
			options: []GatewayOption{WithBucketName("   ")},
			wantErr: true,
			errMsg:  "bucket name cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gateway, err := NewGateway(tt.instances, tt.options...)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Fatalf("error = %v, want containing %q", err, tt.errMsg)
				}
				if gateway != nil {
					t.Fatal("expected nil gateway on error")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gateway == nil {
				t.Fatal("expected non-nil gateway")
			}

			_ = gateway.Close()
		})
	}
}

func TestGatewayPutObjectValidation(t *testing.T) {
	instances := []discovery.MinioInstance{
		{ID: "instance-1", Host: "localhost", Port: "9000", AccessKey: "minioadmin", SecretKey: "minioadmin"},
	}

	gateway, err := NewGateway(instances)
	if err != nil {
		t.Fatalf("failed to create gateway: %v", err)
	}
	defer gateway.Close()

	if err := gateway.PutObject(context.Background(), "invalid-id!", strings.NewReader("data"), 4); err == nil {
		t.Fatal("expected error for invalid object id")
	}

	if err := gateway.PutObject(context.Background(), "object1", nil, 0); err == nil {
		t.Fatal("expected error for nil data")
	}

	if err := gateway.PutObject(context.Background(), "object1", strings.NewReader("data"), -1); err == nil {
		t.Fatal("expected error for negative size")
	}
}

func TestGatewayGetObjectValidation(t *testing.T) {
	instances := []discovery.MinioInstance{
		{ID: "instance-1", Host: "localhost", Port: "9000", AccessKey: "minioadmin", SecretKey: "minioadmin"},
	}

	gateway, err := NewGateway(instances)
	if err != nil {
		t.Fatalf("failed to create gateway: %v", err)
	}
	defer gateway.Close()

	if _, err := gateway.GetObject(context.Background(), "invalid-id!"); err == nil {
		t.Fatal("expected error for invalid object id")
	}
}
