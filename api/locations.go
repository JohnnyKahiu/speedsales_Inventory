package api

import (
	"encoding/json"
	"net/http"

	"github.com/JohnnyKahiu/speedsales_inventory/internal/locations"
)

func locationsPost(w http.ResponseWriter, r *http.Request) {
	respMap := locations.POST(w, r)

	EnableCors(&w)
	jstr, err := json.Marshal(respMap)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.Write(jstr)
}
