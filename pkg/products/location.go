package products

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/JohnnyKahiu/speedsales_inventory/database"
)

type Locations struct {
	table      string        `name:"stock_locations" type:"table"`
	AutoID     int64         `json:"auto_id" name:"auto_id" type:"field" sql:"BIGSERIAL NOT NULL"`
	StoreNum   int           `json:"store_num" type:"field" sql:"INT NOT NULL"`
	StoreName  string        `json:"store_name" type:"field" sql:"VARCHAR NOT NULL"`
	StorageLoc string        `json:"storage_location" type:"field" sql:"VARCHAR NOT NULL"`
	Aisle      string        `json:"aisle" type:"field" sql:"VARCHAR(30) NOT NULL"`
	Level      string        `json:"level" type:"field" sql:"VARCHAR(30) NOT NULL"`
	Bin        string        `json:"bin" type:"field" sql:"VARCHAR(50) NOT NULL"`
	StockList  []string      `json:"stock_list" type:"field" sql:"VARCHAR(50)[]"`
	IsSaleLoc  bool          `json:"is_sale_loc" type:"field" sql:"BOOLEAN NOT NULL DEFAULT false"`
	constraint string        `name:"pk_stk_loc" type:"constraint" sql:"PRIMARY KEY(auto_id)"`
	ItemCode   string        `json:"item_code" `
	Products   []StockMaster `json:"products"`
	IDS        []int         `json:"ids"`
}

type StoreID struct {
	table     string `name:"store"`
	AutoID    int64  `json:"auto_id" type:"field" sql:"BIGSERIAL NOT NULL"`
	StoreName string `json:"store_name" type:"field" sql:"VARCHAR NOT NULL"`
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

// GetLocID fetches stock location data
// returns an error if it fails
func (arg *Locations) GetAllLocInBranch(ctx context.Context) error {
	sql := `SELECT 
				auto_id
				, store_num
				, store_name
				, storage_location
			FROM stock_locations
			WHERE store_name = $1 `

	rows, err := database.PgPool.Query(ctx, sql, arg.StoreName)
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

		arg.IDS = append(arg.IDS, int(arg.AutoID))
	}

	return nil
}

// GetSaleLoc fetches stock location data
// following a hierarchy and stock availability to get where sale was deducted
// returns an error if it fails
func (arg *Locations) GetSaleLoc(ctx context.Context) error {
	fmt.Println("\n\t store_name =", arg.StoreName)
	sql := `SELECT 
				auto_id
				, store_num
				, store_name
				, storage_location
			FROM stock_locations
			WHERE store_name = $1 
				AND $2 = ANY(stock_list) 
				AND is_sale_loc = true
			ORDER BY auto_id ASC 
			LIMIT 1`

	rows, err := database.PgPool.Query(ctx, sql, arg.StoreName, arg.ItemCode)
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
		// log.Fatalf("\t location_id = %v\n", arg.AutoID)
	}

	return nil
}

// GenNew creates a new Location
// inserts into stock_locations table
// returns an error if it fails
func (arg *Locations) GenNew(ctx context.Context) error {
	sql := `INSERT INTO stock_locations (store_num, store_name, storage_location, aisle, level)
			VALUES($1, $2, $3, $4, $5)`

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	_, err := database.PgPool.Exec(ctx, sql, arg.StoreNum, arg.StoreName, arg.StorageLoc, arg.Aisle, arg.Level)
	if err != nil {
		log.Println("error failed to add new stock_location    err =", err)
		return err
	}

	return nil
}

// Fetch gets all location for a stire name or number
// queries stock_locations database
// returns a slice of locations or error if it fails
func (arg *Locations) Fetch() ([]Locations, error) {
	sql := `SELECT 
				auto_id, store_num, 
				store_name, storage_location, 
				aisle, level, 
				stock_list
			FROM stock_locations 
			WHERE store_name = $1 OR store_num = $2
			ORDER BY storage_location, auto_id`

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	fmt.Printf("sql = %s \n store_name = %s \n", sql, arg.StoreName)
	rows, err := database.PgPool.Query(ctx, sql, arg.StoreName, arg.StoreNum)
	if err != nil {
		log.Println("postgres error.    err =", err)
		return []Locations{}, nil
	}
	defer rows.Close()

	vals := []Locations{}
	for rows.Next() {
		r := Locations{}
		err := rows.Scan(&r.AutoID, &r.StoreNum, &r.StoreName, &r.StorageLoc, &r.Aisle, &r.Level, &r.StockList)
		if err != nil {
			log.Println("error scannig locations data    err =", err)
			return vals, nil
		}

		if r.StockList == nil {
			r.StockList = []string{}
		}

		r.Products, err = GetMultiCodes(r.StockList)
		if err != nil {
			return vals, nil
		}

		vals = append(vals, r)
	}

	return vals, nil
}

// Details gets details of a location
// queries information of stockLocation
// populates Location and returns an error when it fails
func (arg *Locations) Details(ctxt context.Context) error {
	ctx, cancel := context.WithTimeout(ctxt, 15*time.Second)
	defer cancel()

	sql := `SELECT 
				auto_id, store_num, 
				store_name, storage_location, 
				aisle, level, 
				stock_list
			FROM stock_locations 
			WHERE auto_id = $1`

	rows, err := database.PgPool.Query(ctx, sql, arg.AutoID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&arg.AutoID, &arg.StoreNum, &arg.StoreName, &arg.StorageLoc, &arg.Aisle, &arg.Level, &arg.StockList)
		if err != nil {
			return err
		}

		fmt.Println("stock_list =", arg.StockList)
		arg.Products, err = GetMultiCodes(arg.StockList)
		if err != nil {
			log.Println("error failed to get products     err =", err)
			return err
		}
		fmt.Println("products = ", arg.Products)
	}

	return nil
}

// Adds a stock_list item to location
// appends item_code to stock_list
// updates item to stock_locations database
// returns an error if it fails
func (arg *Locations) AddToStockList(cntxt context.Context) error {
	ctx, cancel := context.WithTimeout(cntxt, 15*time.Second)
	defer cancel()

	sql := `UPDATE stock_locations
			SET
				stock_list = $1
			WHERE auto_id = $2`

	fmt.Printf("\n\t stock_list = %v \n\t location_id = %v\n", arg.StockList, arg.AutoID)

	_, err := database.PgPool.Exec(ctx, sql, arg.StockList, arg.AutoID)
	if err != nil {
		log.Println("postgres error. failed to update 'stock_list' in 'stock_locations'    err =", err)
		return err
	}

	fmt.Println("\t updated stock_list successfully")
	return nil
}

// FetchMultiIDs
func (arg *Locations) FetchMultiIDs(cntxt context.Context) ([]Locations, error) {
	sql := `SELECT 
				auto_id, store_num, 
				store_name, storage_location, 
				aisle, level, 
				stock_list
			FROM stock_locations 
			WHERE auto_id = ANY($1)
			ORDER BY storage_location, auto_id`

	ctx, cancel := context.WithTimeout(cntxt, 15*time.Second)
	defer cancel()

	fmt.Printf("multi sql = %s \n store_name = %s \n", sql, arg.IDS)
	rows, err := database.PgPool.Query(ctx, sql, arg.IDS)
	if err != nil {
		log.Println("postgres error.    err =", err)
		return []Locations{}, nil
	}
	defer rows.Close()

	vals := []Locations{}
	for rows.Next() {
		r := Locations{}
		err := rows.Scan(&r.AutoID, &r.StoreNum, &r.StoreName, &r.StorageLoc, &r.Aisle, &r.Level, &r.StockList)
		if err != nil {
			log.Println("error scannig locations data    err =", err)
			return vals, nil
		}

		if r.StockList == nil {
			r.StockList = []string{}
		}

		r.Products, err = GetMultiCodes(r.StockList)
		if err != nil {
			return vals, nil
		}

		vals = append(vals, r)
	}

	return vals, nil
}
