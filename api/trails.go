package api

import (
	"encoding/json"
	"net/http"

	"github.com/JohnnyKahiu/speedsales_inventory/internal/ledger"
)

func GetTrails(w http.ResponseWriter, r *http.Request) {
	respMap := ledger.GET(w, r)

	EnableCors(&w)

	w.WriteHeader(http.StatusOK)
	jstr, err := json.Marshal(respMap)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"error converting to json"}`))
		return
	}
	if respMap["response"] == "forbidden" {
		w.WriteHeader(http.StatusForbidden)
	}
	if respMap["response"] == "error" {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.Write(jstr)
}
