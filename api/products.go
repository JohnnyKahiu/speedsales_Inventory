package api

import (
	"encoding/json"
	"net/http"

	"github.com/JohnnyKahiu/speedsales_inventory/internal/product"
)

func PostProducts(w http.ResponseWriter, r *http.Request) {
	respMap := product.PostRoutes(w, r)

	EnableCors(&w)
	jstr, err := json.Marshal(respMap)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("json error"))
		return
	}

	w.Write(jstr)
}

// UpdateProducts post router to update product
func UpdateProducts(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	if token == "" {
		w.WriteHeader(401)
		w.Write([]byte("unauthorized"))
		return
	}

	respMap := product.UpdateRoutes(w, r)

	EnableCors(&w)
	jstr, err := json.Marshal(respMap)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("json error"))
		return
	}

	w.Write(jstr)
}

func DelProducts(w http.ResponseWriter, r *http.Request) {
	respMap := product.DelRoutes(w, r)

	EnableCors(&w)
	jstr, err := json.Marshal(respMap)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("json error"))
		return
	}

	w.Write(jstr)
}

func CatalogueGet(w http.ResponseWriter, r *http.Request) {
	respMap := product.CatalogueGet(w, r)

	EnableCors(&w)
	jstr, err := json.Marshal(respMap)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("json error"))
		return
	}

	w.Write(jstr)
}

func GetGroups(w http.ResponseWriter, r *http.Request) {
	respMap := product.GetGroups(w, r)

	EnableCors(&w)
	jstr, err := json.Marshal(respMap)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("json error"))
		return
	}

	w.Write(jstr)
}
