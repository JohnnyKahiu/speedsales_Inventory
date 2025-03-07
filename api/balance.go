package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/JohnnyKahiu/speedsales_inventory/internal/search"
)

func BalancehGet(w http.ResponseWriter, r *http.Request) {
	respMap := search.GetRoutes(w, r)

	jstr, err := json.Marshal(respMap)
	if err != nil {
		log.Println("error failed to marshal json response    err =", err)
	}

	EnableCors(&w)
	w.Write(jstr)
}
