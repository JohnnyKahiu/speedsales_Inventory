package products

import "github.com/JohnnyKahiu/speedsales_inventory/database"

type Locations struct {
	table      string `name:"stock_locations" type:"table"`
	AutoID     int64  `json:"auto_id" name:"auto_id" type:"field" sql:"BIGSERIAL NOT NULL"`
	StoreNum   int    `json:"store_num" type:"field" sql:"INT NOT NULL"`
	StoreName  string `json:"store_name" type:"field" sql:"VARCHAR NOT NULL"`
	StorageLoc string `json:"storage_location" type:"field" sql:"VARCHAR NOT NULL"`
	Aisle      string `json:"aisle" type:"field" sql:"VARCHAR NOT NULL"`
}

func genLocationsTbl() error {
	var tblStruct Locations
	return database.CreateFromStruct(tblStruct)
}
