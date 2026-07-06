package api

import (
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/JohnnyKahiu/speedsales_inventory/pkg/products"
	"github.com/gorilla/mux"
)

const maxImageSize = 5 << 20 // 5 MB

// PostItemImage handles multipart image uploads for a product.
// POST /products/image/{item_code}
// Form field: "image" (file)
func PostItemImage(w http.ResponseWriter, r *http.Request) {
	EnableCors(&w)

	itemCode := mux.Vars(r)["item_code"]
	if itemCode == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"response":"error","message":"item_code is required"}`))
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxImageSize)
	if err := r.ParseMultipartForm(maxImageSize); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"response":"error","message":"image too large or bad form"}`))
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"response":"error","message":"missing image field"}`))
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"response":"error","message":"failed to read image"}`))
		return
	}

	ct := header.Header.Get("Content-Type")
	if ct == "" {
		ct = http.DetectContentType(data)
	}

	filename, err := products.SaveImage(itemCode, data, ct)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"response":"error","message":"` + err.Error() + `"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"response":"success","image":"` + filename + `"}`))
}

// GetItemImage serves the stored image for a product.
// GET /products/image/{item_code}
func GetItemImage(w http.ResponseWriter, r *http.Request) {
	EnableCors(&w)

	itemCode := mux.Vars(r)["item_code"]
	if itemCode == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	imgPath := products.ImagePath(itemCode)
	if imgPath == "" {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"response":"error","message":"no image found"}`))
		return
	}

	f, err := os.Open(imgPath)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"response":"error","message":"image not found on disk"}`))
		return
	}
	defer f.Close()

	ct := mime.TypeByExtension(filepath.Ext(imgPath))
	if ct == "" {
		ct = "application/octet-stream"
	}
	w.Header().Set("Content-Type", ct)
	w.Header().Set("Cache-Control", "public, max-age=86400")
	io.Copy(w, f)
}
