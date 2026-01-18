package api

import (
	"net/http"

	"github.com/irensaltali/object-storage-gateway/internal/handlers"
	"github.com/irensaltali/object-storage-gateway/internal/storage"

	"github.com/gorilla/mux"
)

func NewRouter(gateway *storage.Gateway) *mux.Router {
	router := mux.NewRouter()

	// User routes
	router.HandleFunc("/object/{id}", func(w http.ResponseWriter, r *http.Request) {
		handlers.PutObject(w, r, gateway)
	}).Methods("PUT")
	router.HandleFunc("/object/{id}", func(w http.ResponseWriter, r *http.Request) {
		handlers.GetObject(w, r, gateway)
	}).Methods("GET")

	return router
}
