package api

import (
	"encoding/json"
	"net/http"

	"github.com/JohnnyKahiu/speedsales_inventory/internal/purchase"
)

func PostPurchase(w http.ResponseWriter, r *http.Request) {
	respMap := purchase.POST(w, r)

	jstr, err := json.Marshal(respMap)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jstr)
}

func GetPurchase(w http.ResponseWriter, r *http.Request) {
	respMap := purchase.GET(w, r)

	jstr, err := json.Marshal(respMap)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
		return
	}

	w.WriteHeader(http.StatusOK)
	if respMap["response"] == "error" {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Write(jstr)
}

func DelPurchase(w http.ResponseWriter, r *http.Request) {
	respMap := purchase.DELETE(w, r)

	jstr, err := json.Marshal(respMap)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jstr)

}
