package handlers

import (
	"net/http"
)

// PutObject handles the PUT /object/{id} endpoint
func PutObject(w http.ResponseWriter, r *http.Request) {
	// vars := mux.Vars(r)
	// id := vars["id"]

	w.WriteHeader(http.StatusCreated)
}

// GetObject handles the GET /object/{id} endpoint
func GetObject(w http.ResponseWriter, r *http.Request) {
	// vars := mux.Vars(r)
	// id := vars["id"]

	w.WriteHeader(http.StatusOK)
}
