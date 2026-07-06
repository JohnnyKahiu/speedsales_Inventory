package products

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/JohnnyKahiu/speedsales_inventory/database"
	"github.com/JohnnyKahiu/speedsales_inventory/pkg/variables"
)

// ImageDir returns the path to the images directory, creating it if needed.
func ImageDir() string {
	dir := filepath.Join(variables.Fpath, "images")
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, os.ModePerm)
	}
	return dir
}

// SaveImage writes image bytes to disk, persists the path in the DB and cache,
// and returns the stored file path (relative filename).
func SaveImage(itemCode string, data []byte, contentType string) (string, error) {
	if itemCode == "" {
		return "", fmt.Errorf("item_code is required")
	}

	ext := extensionFromContentType(contentType)
	if ext == "" {
		return "", fmt.Errorf("unsupported image type: %s", contentType)
	}

	filename := itemCode + ext
	fullPath := filepath.Join(ImageDir(), filename)

	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		log.Println("error saving image file    err =", err)
		return "", err
	}

	if err := updateImageInDB(itemCode, filename); err != nil {
		return "", err
	}

	updateImageInCache(itemCode, filename)
	return filename, nil
}

// ImagePath returns the full filesystem path for a given item_code's image,
// or an empty string if no image is stored.
func ImagePath(itemCode string) string {
	item, ok := ProdMaster.ProductDB[itemCode]
	if !ok || item.Image == "" {
		return ""
	}
	return filepath.Join(ImageDir(), filepath.Base(item.Image))
}

func updateImageInDB(itemCode, filename string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sql := `UPDATE stock_master SET image = $1 WHERE item_code = $2`
	_, err := database.PgPool.Exec(ctx, sql, filename, itemCode)
	if err != nil {
		log.Println("error updating image in DB    err =", err)
		return err
	}
	return nil
}

func updateImageInCache(itemCode, filename string) {
	item, ok := ProdMaster.ProductDB[itemCode]
	if !ok {
		return
	}
	item.Image = filename
	ProdMaster.ProductDB[itemCode] = item
	if err := ProdMaster.Pickle(); err != nil {
		log.Println("error pickling after image update    err =", err)
	}
}

func extensionFromContentType(ct string) string {
	ct = strings.ToLower(strings.TrimSpace(ct))
	switch {
	case strings.Contains(ct, "jpeg") || strings.Contains(ct, "jpg"):
		return ".jpg"
	case strings.Contains(ct, "png"):
		return ".png"
	case strings.Contains(ct, "webp"):
		return ".webp"
	default:
		return ""
	}
}
