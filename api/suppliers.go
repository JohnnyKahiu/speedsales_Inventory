package api

import (
	"encoding/json"
	"net/http"

	"github.com/JohnnyKahiu/speedsales_inventory/internal/supplier"
)

func PostSuppliers(w http.ResponseWriter, r *http.Request) {
	respMap := supplier.POST(w, r)

	EnableCors(&w)
	jstr, err := json.Marshal(respMap)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"response": "error"}`))
		return
	}

	w.Write(jstr)
}

func GetSuppliers(w http.ResponseWriter, r *http.Request) {
	respMap := supplier.GET(w, r)

	EnableCors(&w)
	jstr, err := json.Marshal(respMap)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"response": "error"}`))
		return
	}

	w.Write(jstr)
}
