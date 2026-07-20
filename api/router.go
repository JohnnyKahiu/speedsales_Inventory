package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/JohnnyKahiu/speedsales_inventory/pkg/authentication"
	"github.com/JohnnyKahiu/speedsales_inventory/pkg/products"
	"github.com/gorilla/mux"
)

var mySigningKey = []byte(os.Getenv("SPEEDSALESJWTKEY"))

func EnableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Headers", "*")
}

func NewRouter() *mux.Router {
	r := mux.NewRouter()

	// Public: image serving bypasses JWT.
	r.HandleFunc("/products/image/{item_code}", GetItemImage).Methods("GET", "OPTIONS")

	// All other routes require a valid JWT token.
	p := r.NewRoute().Subrouter()
	p.Use(JwtMiddleware)

	p.HandleFunc("/app/{module}", AppGet).Methods("GET", "OPTIONS")
	p.HandleFunc("/app/{module}", AppPost).Methods("POST", "OPTIONS")

	p.HandleFunc("/products/search/{module}", SearchGet).Methods("GET", "OPTIONS")
	p.HandleFunc("/products/balance/{module}", BalanceGet).Methods("GET", "OPTIONS")
	p.HandleFunc("/products/catalogue/{supplier}", CatalogueGet).Methods("GET", "OPTIONS")
	p.HandleFunc("/products/groups/{key}", GetGroups).Methods("GET", "OPTIONS")

	p.HandleFunc("/products/image/{item_code}", PostItemImage).Methods("POST", "OPTIONS")

	p.HandleFunc("/products/{module}", PostProducts).Methods("POST", "OPTIONS")
	p.HandleFunc("/products/update/{module}", UpdateProducts).Methods("POST", "OPTIONS")

	p.HandleFunc("/products/locations/{module}", locationsPost).Methods("POST", "OPTIONS")
	p.HandleFunc("/products/locations/{module}", locationsGet).Methods("GET", "OPTIONS")

	p.HandleFunc("/products/{module}/{code}", DelProducts).Methods("DELETE", "OPTIONS")

	p.HandleFunc("/products/stock_take/{module}", PostCounts).Methods("POST", "OPTIONS")
	p.HandleFunc("/products/stock_take/{module}", GetCounts).Methods("GET", "OPTIONS")

	p.HandleFunc("/suppliers/{module}", PostSuppliers).Methods("POST", "OPTIONS")
	p.HandleFunc("/suppliers/{module}", GetSuppliers).Methods("GET", "OPTIONS")

	p.HandleFunc("/aquisition/purchase/{module}", PostPurchase).Methods("POST", "OPTIONS")
	p.HandleFunc("/aquisition/purchase/{module}", GetPurchase).Methods("GET", "OPTIONS")
	p.HandleFunc("/aquisition/purchase/{module}", DelPurchase).Methods("DELETE", "OPTIONS")

	p.HandleFunc("/reports/trail/{code}", GetTrails).Methods("GET", "OPTIONS")

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

		loc := products.Locations{StoreName: user.Branch, StorageLoc: user.StkLocation}

		err := loc.GetLocID(context.Background())
		if err != nil {
			log.Println("\n\t token string is not valid")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"response": "error", "message": "unauthorized"}`))
			return
		}

		// r.Header.Set("", fmt.Sprintf("%v", loc.AutoID))
		next.ServeHTTP(w, r)
	})
}
