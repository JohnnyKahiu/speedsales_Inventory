package api

import (
	"encoding/json"
	"net/http"

	"github.com/JohnnyKahiu/speedsales_inventory/internal/products"
)

func PostProducts(w http.ResponseWriter, r *http.Request) {
	respMap := products.PostRoutes(w, r)

	EnableCors(&w)
	jstr, err := json.Marshal(respMap)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("json error"))
		return
	}

	w.Write(jstr)
}
