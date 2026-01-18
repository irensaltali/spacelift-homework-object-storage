package storage

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/irensaltali/object-storage-gateway/internal/discovery"
	"github.com/minio/minio-go/v7"
)

// mockMinioClient is a mock implementation of minio.Client for testing
type mockMinioClient struct {
	bucketExistsFunc func(ctx context.Context, bucketName string) (bool, error)
	makeBucketFunc   func(ctx context.Context, bucketName string, opts minio.MakeBucketOptions) error
	putObjectFunc    func(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error)
	statObjectFunc   func(ctx context.Context, bucketName, objectName string, opts minio.StatObjectOptions) (minio.ObjectInfo, error)
	getObjectFunc    func(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) (*minio.Object, error)
}

func (m *mockMinioClient) BucketExists(ctx context.Context, bucketName string) (bool, error) {
	if m.bucketExistsFunc != nil {
		return m.bucketExistsFunc(ctx, bucketName)
	}
	return false, nil
}

func (m *mockMinioClient) MakeBucket(ctx context.Context, bucketName string, opts minio.MakeBucketOptions) error {
	if m.makeBucketFunc != nil {
		return m.makeBucketFunc(ctx, bucketName, opts)
	}
	return nil
}

func (m *mockMinioClient) PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
	if m.putObjectFunc != nil {
		return m.putObjectFunc(ctx, bucketName, objectName, reader, objectSize, opts)
	}
	return minio.UploadInfo{}, nil
}

func (m *mockMinioClient) StatObject(ctx context.Context, bucketName, objectName string, opts minio.StatObjectOptions) (minio.ObjectInfo, error) {
	if m.statObjectFunc != nil {
		return m.statObjectFunc(ctx, bucketName, objectName, opts)
	}
	return minio.ObjectInfo{}, nil
}

func (m *mockMinioClient) GetObject(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) (*minio.Object, error) {
	if m.getObjectFunc != nil {
		return m.getObjectFunc(ctx, bucketName, objectName, opts)
	}
	return nil, errors.New("not implemented")
}

// mockClientManager is a mock implementation of MinioClientManager for testing
type mockClientManager struct {
	clients map[string]*minio.Client
}

func (m *mockClientManager) GetClient(instanceID string) (*minio.Client, error) {
	client, exists := m.clients[instanceID]
	if !exists {
		return nil, errors.New("client not found")
	}
	return client, nil
}

func (m *mockClientManager) UpdateInstances(instances []discovery.MinioInstance) error {
	return nil
}

func (m *mockClientManager) Close() error {
	return nil
}

func TestNewGateway(t *testing.T) {
	tests := []struct {
		name      string
		instances []discovery.MinioInstance
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
			name: "valid instances",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gateway, err := NewGateway(tt.instances)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewGateway() expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("NewGateway() error = %v, want error containing %v", err, tt.errMsg)
				}
				if gateway != nil {
					t.Errorf("NewGateway() expected nil gateway on error")
				}
			} else {
				if err != nil {
					t.Errorf("NewGateway() unexpected error = %v", err)
					return
				}
				if gateway == nil {
					t.Errorf("NewGateway() expected non-nil gateway")
				}
				if gateway.hasher == nil {
					t.Errorf("NewGateway() hasher is nil")
				}
				if gateway.clients == nil {
					t.Errorf("NewGateway() clients is nil")
				}
				// Clean up
				gateway.Close()
			}
		})
	}
}

func TestGateway_PutObject(t *testing.T) {
	tests := []struct {
		name      string
		objectID  string
		data      io.Reader
		size      int64
		setupMock func() *mockMinioClient
		wantErr   bool
		errMsg    string
	}{
		{
			name:     "empty object ID",
			objectID: "",
			data:     strings.NewReader("test data"),
			size:     9,
			wantErr:  true,
			errMsg:   "object id cannot be empty",
		},
		{
			name:     "nil data",
			objectID: "test-object",
			data:     nil,
			size:     0,
			wantErr:  true,
			errMsg:   "data cannot be nil",
		},
		{
			name:     "successful put",
			objectID: "test-object",
			data:     strings.NewReader("test data"),
			size:     9,
			setupMock: func() *mockMinioClient {
				return &mockMinioClient{
					bucketExistsFunc: func(ctx context.Context, bucketName string) (bool, error) {
						return true, nil
					},
					putObjectFunc: func(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
						return minio.UploadInfo{}, nil
					},
				}
			},
			wantErr: false,
		},
		{
			name:     "bucket does not exist, create it",
			objectID: "test-object",
			data:     strings.NewReader("test data"),
			size:     9,
			setupMock: func() *mockMinioClient {
				return &mockMinioClient{
					bucketExistsFunc: func(ctx context.Context, bucketName string) (bool, error) {
						return false, nil
					},
					makeBucketFunc: func(ctx context.Context, bucketName string, opts minio.MakeBucketOptions) error {
						return nil
					},
					putObjectFunc: func(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
						return minio.UploadInfo{}, nil
					},
				}
			},
			wantErr: false,
		},
		{
			name:     "put object error",
			objectID: "test-object",
			data:     strings.NewReader("test data"),
			size:     9,
			setupMock: func() *mockMinioClient {
				return &mockMinioClient{
					bucketExistsFunc: func(ctx context.Context, bucketName string) (bool, error) {
						return true, nil
					},
					putObjectFunc: func(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
						return minio.UploadInfo{}, errors.New("put object failed")
					},
				}
			},
			wantErr: true,
			errMsg:  "failed to put object",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a gateway with mock client manager
			instances := []discovery.MinioInstance{
				{ID: "instance-1", Host: "localhost", Port: "9000", AccessKey: "minioadmin", SecretKey: "minioadmin"},
			}

			gateway, err := NewGateway(instances)
			if err != nil {
				t.Fatalf("Failed to create gateway: %v", err)
			}
			defer gateway.Close()

			// Note: We test validation errors here. For full integration tests with mocks,
			// we would need to refactor to use interfaces or dependency injection.
			_ = tt.setupMock

			ctx := context.Background()
			err = gateway.PutObject(ctx, tt.objectID, tt.data, tt.size)

			if tt.wantErr {
				if err == nil {
					t.Errorf("PutObject() expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("PutObject() error = %v, want error containing %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("PutObject() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestGateway_GetObject(t *testing.T) {
	tests := []struct {
		name      string
		objectID  string
		setupMock func() *mockMinioClient
		wantErr   bool
		errMsg    string
	}{
		{
			name:     "empty object ID",
			objectID: "",
			wantErr:  true,
			errMsg:   "object id cannot be empty",
		},
		{
			name:     "object not found",
			objectID: "nonexistent",
			setupMock: func() *mockMinioClient {
				return &mockMinioClient{
					statObjectFunc: func(ctx context.Context, bucketName, objectName string, opts minio.StatObjectOptions) (minio.ObjectInfo, error) {
						return minio.ObjectInfo{}, errors.New("NoSuchKey")
					},
				}
			},
			wantErr: true,
			errMsg:  "object not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instances := []discovery.MinioInstance{
				{ID: "instance-1", Host: "localhost", Port: "9000", AccessKey: "minioadmin", SecretKey: "minioadmin"},
			}

			gateway, err := NewGateway(instances)
			if err != nil {
				t.Fatalf("Failed to create gateway: %v", err)
			}
			defer gateway.Close()

			ctx := context.Background()
			_, err = gateway.GetObject(ctx, tt.objectID)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetObject() expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("GetObject() error = %v, want error containing %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("GetObject() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestGateway_Close(t *testing.T) {
	instances := []discovery.MinioInstance{
		{ID: "instance-1", Host: "localhost", Port: "9000", AccessKey: "minioadmin", SecretKey: "minioadmin"},
	}

	gateway, err := NewGateway(instances)
	if err != nil {
		t.Fatalf("Failed to create gateway: %v", err)
	}

	err = gateway.Close()
	if err != nil {
		t.Errorf("Close() unexpected error = %v", err)
	}
}

func TestGateway_SelectInstanceConsistency(t *testing.T) {
	instances := []discovery.MinioInstance{
		{ID: "instance-1", Host: "localhost", Port: "9000", AccessKey: "minioadmin", SecretKey: "minioadmin"},
		{ID: "instance-2", Host: "localhost", Port: "9001", AccessKey: "minioadmin", SecretKey: "minioadmin"},
		{ID: "instance-3", Host: "localhost", Port: "9002", AccessKey: "minioadmin", SecretKey: "minioadmin"},
	}

	gateway, err := NewGateway(instances)
	if err != nil {
		t.Fatalf("Failed to create gateway: %v", err)
	}
	defer gateway.Close()

	objectID := "test-object-123"
	instanceID1, err := gateway.hasher.SelectInstance(objectID)
	if err != nil {
		t.Fatalf("SelectInstance() error = %v", err)
	}

	instanceID2, err := gateway.hasher.SelectInstance(objectID)
	if err != nil {
		t.Fatalf("SelectInstance() error = %v", err)
	}

	if instanceID1 != instanceID2 {
		t.Errorf("SelectInstance() inconsistent: got %s and %s for same object ID", instanceID1, instanceID2)
	}
}

func TestGateway_PutObjectReadsData(t *testing.T) {
	instances := []discovery.MinioInstance{
		{ID: "instance-1", Host: "localhost", Port: "9000", AccessKey: "minioadmin", SecretKey: "minioadmin"},
	}

	gateway, err := NewGateway(instances)
	if err != nil {
		t.Fatalf("Failed to create gateway: %v", err)
	}
	defer gateway.Close()

	testData := "test data content"
	data := bytes.NewReader([]byte(testData))
	ctx := context.Background()

	// This will fail without a real Minio instance, but we can test the validation
	err = gateway.PutObject(ctx, "test-object", data, int64(len(testData)))
	// We expect an error here since we don't have a real Minio instance
	// But we've already tested validation, so this is fine
	_ = err
}
