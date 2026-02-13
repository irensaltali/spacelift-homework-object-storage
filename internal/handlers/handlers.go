package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/irensaltali/object-storage-gateway/internal/storage"
)

// ObjectGateway captures the storage behavior handlers depend on.
type ObjectGateway interface {
	PutObject(ctx context.Context, objectKey string, data io.Reader, size int64) error
	GetObject(ctx context.Context, objectKey string) (io.ReadCloser, error)
}

// PutObject handles the PUT /object/{id} endpoint
func PutObject(w http.ResponseWriter, r *http.Request, gateway ObjectGateway) {
	objectKey := mux.Vars(r)["id"]
	log.Printf("PUT /object/%s - received request, content-length: %d", objectKey, r.ContentLength)

	if err := storage.ValidateObjectID(objectKey); err != nil {
		log.Printf("PUT /object/%s - invalid object id: %v", objectKey, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get content length from request header
	contentLength := r.ContentLength
	if contentLength < 0 {
		log.Printf("PUT /object/%s - error: content-length header is required", objectKey)
		http.Error(w, "content-length header is required", http.StatusBadRequest)
		return
	}

	// Store object by gateway
	if err := gateway.PutObject(r.Context(), objectKey, r.Body, contentLength); err != nil {
		if errors.Is(err, storage.ErrInvalidObjectID) {
			log.Printf("PUT /object/%s - invalid object id: %v", objectKey, err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Printf("PUT /object/%s - error storing object: %v", objectKey, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("PUT /object/%s - object stored successfully", objectKey)

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]any{
		"id":      objectKey,
		"status":  "stored",
		"message": "object stored successfully",
	}

	json.NewEncoder(w).Encode(response)
}

// GetObject handles the GET /object/{id} endpoint
func GetObject(w http.ResponseWriter, r *http.Request, gateway ObjectGateway) {
	objectKey := mux.Vars(r)["id"]
	log.Printf("GET /object/%s - received request", objectKey)

	if err := storage.ValidateObjectID(objectKey); err != nil {
		log.Printf("GET /object/%s - invalid object id: %v", objectKey, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Retrieve object from gateway
	object, err := gateway.GetObject(r.Context(), objectKey)
	if err != nil {
		if errors.Is(err, storage.ErrInvalidObjectID) {
			log.Printf("GET /object/%s - invalid object id: %v", objectKey, err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if errors.Is(err, storage.ErrObjectNotFound) {
			log.Printf("GET /object/%s - object not found", objectKey)
			http.Error(w, "object not found", http.StatusNotFound)
			return
		}
		log.Printf("GET /object/%s - error retrieving object: %v", objectKey, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer object.Close()

	log.Printf("GET /object/%s - object found, streaming to client", objectKey)

	// Set response headers
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)

	// Stream object data to response
	if _, err := io.Copy(w, object); err != nil {
		log.Printf("GET /object/%s - error streaming object: %v", objectKey, err)
		return
	}

	log.Printf("GET /object/%s - object streamed successfully", objectKey)
}
