package products

import (
	"context"
	"log"

	"github.com/JohnnyKahiu/speedsales_inventory/database"
)

type Locations struct {
	table      string `name:"stock_locations" type:"table"`
	AutoID     int64  `json:"auto_id" name:"auto_id" type:"field" sql:"BIGSERIAL NOT NULL"`
	StoreNum   int    `json:"store_num" type:"field" sql:"INT NOT NULL"`
	StoreName  string `json:"store_name" type:"field" sql:"VARCHAR NOT NULL"`
	StorageLoc string `json:"storage_location" type:"field" sql:"VARCHAR NOT NULL"`
	Aisle      string `json:"aisle" type:"field" sql:"VARCHAR NOT NULL"`
	Level      string `json:"level" type:"field" sql:"VARCHAR NOT NULL"`
	constraint string `name:"pk_stk_loc" type:"constraint" sql:"PRIMARY KEY(auto_id)"`
}

func genLocationsTbl() error {
	var tblStruct Locations
	return database.CreateFromStruct(tblStruct)
}

// GetLocID fetches stock location data
// returns an error if it fails
func (arg *Locations) GetLocID(ctx context.Context) error {
	sql := `SELECT 
				auto_id
				, store_num
				, store_name
				, storage_location
			FROM stock_locations
			WHERE store_name = $1 AND storage_location = $2 `

	rows, err := database.PgPool.Query(ctx, sql, arg.StoreName, arg.StorageLoc)
	if err != nil {
		log.Println("sql error   failed to fetch stock_location")
		return err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&arg.AutoID, &arg.StoreNum, &arg.StoreName, &arg.StorageLoc)
		if err != nil {
			log.Println("scan error    failed to scan row    err =", err)
			return err
		}
	}

	return nil
}
