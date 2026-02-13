package handlers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/irensaltali/object-storage-gateway/internal/storage"
)

type mockGateway struct {
	putObjectFn func(ctx context.Context, objectKey string, data io.Reader, size int64) error
	getObjectFn func(ctx context.Context, objectKey string) (io.ReadCloser, error)
}

func (m *mockGateway) PutObject(ctx context.Context, objectKey string, data io.Reader, size int64) error {
	if m.putObjectFn != nil {
		return m.putObjectFn(ctx, objectKey, data, size)
	}
	return nil
}

func (m *mockGateway) GetObject(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	if m.getObjectFn != nil {
		return m.getObjectFn(ctx, objectKey)
	}
	return io.NopCloser(strings.NewReader("ok")), nil
}

func TestPutObject_InvalidObjectID(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/object/not-valid!", strings.NewReader("content"))
	req.ContentLength = int64(len("content"))
	req = mux.SetURLVars(req, map[string]string{"id": "not-valid!"})

	rr := httptest.NewRecorder()

	PutObject(rr, req, &mockGateway{})

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestPutObject_MissingContentLength(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/object/object1", strings.NewReader("content"))
	req.ContentLength = -1
	req = mux.SetURLVars(req, map[string]string{"id": "object1"})

	rr := httptest.NewRecorder()

	PutObject(rr, req, &mockGateway{})

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestPutObject_StorageError(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/object/object1", strings.NewReader("content"))
	req.ContentLength = int64(len("content"))
	req = mux.SetURLVars(req, map[string]string{"id": "object1"})

	rr := httptest.NewRecorder()

	PutObject(rr, req, &mockGateway{
		putObjectFn: func(ctx context.Context, objectKey string, data io.Reader, size int64) error {
			return errors.New("backend down")
		},
	})

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}
}

func TestGetObject_InvalidObjectID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/object/not-valid!", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "not-valid!"})

	rr := httptest.NewRecorder()

	GetObject(rr, req, &mockGateway{})

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestGetObject_NotFoundError(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/object/object1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "object1"})

	rr := httptest.NewRecorder()

	GetObject(rr, req, &mockGateway{
		getObjectFn: func(ctx context.Context, objectKey string) (io.ReadCloser, error) {
			return nil, fmt.Errorf("%w: %s", storage.ErrObjectNotFound, objectKey)
		},
	})

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestGetObject_Success(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/object/object1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "object1"})

	rr := httptest.NewRecorder()

	GetObject(rr, req, &mockGateway{
		getObjectFn: func(ctx context.Context, objectKey string) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader("payload")), nil
		},
	})

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if rr.Body.String() != "payload" {
		t.Fatalf("body = %q, want %q", rr.Body.String(), "payload")
	}
}
