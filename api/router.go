package api

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/JohnnyKahiu/speedsales_inventory/pkg/authentication"
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
	r.Use(JwtMiddleware)

	r.HandleFunc("/products/search/{module}", SearchGet).Methods("GET", "OPTIONS")
	r.HandleFunc("/products/balance/{module}", BalanceGet).Methods("GET", "OPTIONS")
	r.HandleFunc("/products/catalogue/{supplier}", CatalogueGet).Methods("GET", "OPTIONS")
	r.HandleFunc("/products/groups/{key}", GetGroups).Methods("GET", "OPTIONS")

	r.HandleFunc("/products/{module}", PostProducts).Methods("POST", "OPTIONS")
	r.HandleFunc("/products/update/{module}", UpdateProducts).Methods("POST", "OPTIONS")

	r.HandleFunc("/products/{module}/{code}", DelProducts).Methods("DELETE", "OPTIONS")

	// r.HandleFunc("/sms", sms.Post).Methods("POST", "OPTIONS")

	return r
}

func JwtMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		EnableCors(&w)

		tokenString := r.Header.Get("token")
		if tokenString == "" {
			log.Println("\n\t token string not provided")

			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"response": "error", "message": "unauthorized"}`))
			return
		}

		user, authentic := authentication.ValidateJWT(tokenString)
		if !authentic {
			log.Println("\n\t token string is not valid")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"response": "error", "message": "unauthorized"}`))
			return
		}

		juser, _ := json.Marshal(user)

		r.Header.Set("user_details", string(juser))
		next.ServeHTTP(w, r)
	})
}
