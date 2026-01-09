package products

import (
	"context"
	"fmt"
	"log"

	"github.com/JohnnyKahiu/speedsales_inventory/database"
	"github.com/jackc/pgx/v5"
)

type Description struct {
	table     string `name:"product_description" type:"table"`
	ID        int64  `json:"id" type:"field" sql:"BIGSERIAL NOT NULL"`
	ItemCode  string `json:"item_code" type:"field" sql:"VARCHAR NOT NULL UNIQUE"`
	ItemName  string `json:"item_name" type:"field" sql:"VARCHAR NOT NULL"`
	Product   string `json:"product" type:"field" sql:"VARCHAR NOT NULL"`
	BrandName string `json:"brand_name" type:"field" sql:"VARCHAR NOT NULL"`
	Category  string `json:"category" type:"field" sql:"VARCHAR NOT NULL DEFAULT ''"`
	Category1 string `json:"category_1" type:"field" sql:"VARCHAR NOT NULL DEFAULT ''"`
	Category2 string `json:"category_2" type:"field" sql:"VARCHAR NOT NULL DEFAULT ''"`
	Category3 string `json:"category_3" type:"field" sql:"VARCHAR NOT NULL DEFAULT ''"`
	Size      string `json:"size" type:"field" sql:"VARCHAR NOT NULL DEFAULT ''"`
	Color     string `json:"color" type:"field" sql:"VARCHAR NOT NULL DEFAULT ''"`
	pKey      string `type:"constraint" name:"description_pkey" sql:"PRIMARY KEY (id)"`
}

// genDescriptionTbl creates the description table
func genDescriptionTbl() error {
	var tblStruct Description
	return database.CreateFromStruct(tblStruct)
}

// AddDB insert description into database
func AddDB(d Description) error {
	return database.InsertFromStruct(d)
}

// SQLAdd inserts description into database
// returns an error if fails
func (a *Description) SQLAdd(ctx context.Context, tx pgx.Tx) error {
	fmt.Println("\t Description =", *a)
	sql := `INSERT INTO product_description(
				item_code, item_name, product, brand_name, category, category_1, 
				category_2, category_3, size, color
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) 
			ON CONFLICT ON CONSTRAINT product_description_item_code_key
			DO UPDATE SET 
				item_code = $1, item_name = $2, product = $3, brand_name = $4, category = $5, category_1 = $6, 
				category_2 = $7, category_3 = $8, size = $9, color = $10
			RETURNING id`
	_, err := tx.Exec(ctx, sql, a.ItemCode, a.ItemName, a.Product, a.BrandName,
		a.Category, a.Category1, a.Category2, a.Category3, a.Size, a.Color)
	if err != nil {
		log.Println("error, failed to execute row item     err =", err)
		return err
	}

	return nil
}
