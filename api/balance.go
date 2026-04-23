package api

import (
	"encoding/json"
	"log"
	"net/http"
)

func BalanceGet(w http.ResponseWriter, r *http.Request) {
	respMap := make(map[string]interface{})

	jstr, err := json.Marshal(respMap)
	if err != nil {
		log.Println("error failed to marshal json response    err =", err)
	}

	EnableCors(&w)
	w.Write(jstr)
}
