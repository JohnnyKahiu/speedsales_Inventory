package api

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

var mySigningKey = []byte(os.Getenv("SPEEDSALESJWTKEY"))

func EnableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Headers", "*")
}

func NewRouter() *mux.Router {
	// rentals.CreateTables()

	r := mux.NewRouter()

	// r.HandleFunc("/ws", socketHandler)

	r.HandleFunc("/products/search/{module}", SearchGet).Methods("GET", "OPTIONS")
	r.HandleFunc("/products/balance/{module}", SearchGet).Methods("GET", "OPTIONS")

	r.HandleFunc("/products/{module}", PostProducts).Methods("POST", "OPTIONS")

	// r.HandleFunc("/sms", sms.Post).Methods("POST", "OPTIONS")

	return r
}
