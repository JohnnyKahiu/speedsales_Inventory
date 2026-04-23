package api

import (
	"encoding/json"
	"net/http"

	"github.com/JohnnyKahiu/speedsales_inventory/internal/counts"
)

func PostCounts(w http.ResponseWriter, r *http.Request) {
	respMap := counts.POST(w, r)

	EnableCors(&w)
	jstr, err := json.Marshal(respMap)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("json error"))
		return
	}

	w.Write(jstr)
}

func GetCounts(w http.ResponseWriter, r *http.Request) {
	respMap := counts.GET(w, r)

	EnableCors(&w)
	jstr, err := json.Marshal(respMap)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("json error"))
		return
	}

	w.Write(jstr)
}
