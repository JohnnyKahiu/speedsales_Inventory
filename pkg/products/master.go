package products

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/JohnnyKahiu/speedsales_inventory/database"
	"github.com/jackc/pgx/v5"
)

// StockMaster Holds information about a stock item
type StockMaster struct {
	table              string            `type:"table" name:"stock_master"`
	ItemCode           string            `json:"item_code" type:"field" sql:"VARCHAR PRIMARY KEY"`
	ItemName           string            `json:"item_name" type:"field" sql:"VARCHAR NOT NULL"`
	Description        Description       `json:"description"`
	ItemCost           float64           `json:"item_cost" type:"field" sql:"DECIMAL NOT NULL DEFAULT '0.0'"`
	ItemSellingprice   float64           `json:"item_sellingprice" type:"field" sql:"DECIMAL NOT NULL DEFAULT '0.0'"`
	ItemWholesaleprice float64           `json:"item_wholesaleprice" type:"field" sql:"DECIMAL NOT NULL DEFAULT '0.0'"`
	ItemOfferprice     float64           `json:"item_offerprice" type:"field" sql:"DECIMAL NOT NULL DEFAULT '0.0'"`
	OfferStart         time.Time         `json:"offer_start" type:"field" sql:"TIMESTAMP NOT NULL DEFAULT NOW()"`
	OfferEnd           time.Time         `json:"offer_end" type:"field" sql:"TIMESTAMP NOT NULL DEFAULT NOW()"`
	OfferQty           float64           `json:"offer_qty" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	VatAlpha           string            `json:"vat_alpha" type:"field" sql:"VARCHAR(5) NOT NULL"`
	HsCode             string            `json:"hs_code" type:"hs_code" sql:"VARCHAR NOT NULL DEFAULT ''"`
	DeptName           string            `json:"dept_name" type:"field" sql:"VARCHAR NOT NULL"`
	DeptCode           int32             `json:"dept_code" type:"field" sql:"INTEGER NOT NULL"`
	SupplierCode       int64             `json:"supplier_code" type:"field" sql:"INTEGER NOT NULL DEFAULT '1'"`
	ManufucturerCode   int64             `json:"manufucturer_code" type:"field" sql:"INTEGER NOT NULL DEFAULT '0'"`
	ManufucturerName   string            `json:"manufucturer_name" type:"field" sql:"VARCHAR NOT NULL DEFAULT 'undefined'"`
	UnitsPerPack       int               `json:"units_per_pack" type:"field" sql:"VARCHAR NOT NULL"`
	KgWeight           float64           `json:"kg_weight" type:"field" sql:"FLOAT NOT NULL"`
	OldPrice           bool              `json:"old_price" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	PriceChange        bool              `json:"price_change" type:"field" sql:"BOOLEAN NOT NULL DEFAULT False"`
	PriceChangeDate    time.Time         `json:"price_change_date" type:"field" sql:"TIMESTAMP"`
	PriceEffectTime    time.Time         `json:"price_effect_time" type:"field" sql:"TIMESTAMP"`
	PriceEffectLead    int32             `json:"price_effect_lead" type:"field" sql:"INTEGER"`
	IsInventory        bool              `json:"is_inventory" type:"field" sql:"BOOLEAN NOT NULL DEFAULT True"`
	IsSerial           bool              `json:"is_serial" type:"field" sql:"BOOLEAN NOT NULL DEFAULT False"`
	IsBatched          bool              `json:"is_batched" type:"field" sql:"BOOLEAN NOT NULL DEFAULT False"`
	IsReturn           bool              `json:"is_return" type:"field" sql:"BOOLEAN NOT NULL DEFAULT False"`
	ReturnCode         string            `json:"return_code" type:"field" sql:"VARCHAR"`
	UpdateTime         time.Time         `json:"update_time" type:"field" sql:"TIMESTAMP NOT NULL DEFAULT NOW()"`
	MinMargin          float32           `json:"min_margin" type:"field" sql:"FLOAT NOT NULL DEFAULT '0.05'"`
	IsActive           bool              `json:"is_active" type:"field" sql:"BOOL NOT NULL DEFAULT True"`
	IsProduced         bool              `json:"is_produced" type:"field" sql:"BOOL NOT NULL DEFAULT False"`
	Category           string            `json:"category" type:"field" sql:"VARCHAR NOT NULL DEFAULT 'sale item'"`
	UnitsPerRecipe     float64           `json:"units_per_recipe" type:"field" sql:"FLOAT NOT NULL DEFAULT '1'"`
	ReorderLevel       float64           `json:"reorder_level" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	Image              string            `json:"image" type:"field" sql:"VARCHAR"`
	IsCombo            bool              `json:"is_combo" type:"field" sql:"BOOLEAN NOT NULL DEFAULT False"`
	ComboItems         []Combo           `json:"combo_items" type:"field" sql:"JSONB"`
	Balance            map[int64]float64 `json:"balance"`
	OnOffer            bool              `json:"on_offer"`
	TillPrice          float64           `json:"till_price"`
	VatPercent         float64           `json:"vat_percent"`
	Margin             float64           `json:"margin"`
	Markup             float64           `json:"mark_up"`
	PkgQty             float64           `json:"pkg_qty"`
	Disc               float64           `json:"Disc"`
	Label              string            `json:"label"`
	Bal                float64           `json:"Bal"`
}

type Combo struct {
	MasterCode string  `json:"master_code"`
	ItemCode   string  `json:"item_code"`
	ItemName   string  `json:"item_name"`
	Quantity   float64 `json:"quantity"`
}

// Categories Holds information on stock categories
type Categories struct {
	table       string `name:"categories" type:"table"`
	AutoID      int64  `json:"auto_id" name:"auto_id" type:"field" sql:"BIGSERIAL PRIMARY KEY"`
	Name        string `json:"name" name:"name" type:"field" sql:"VARCHAR NOT NULL"`
	SubCategory string `json:"sub_category" name:"sub_category" type:"field" sql:"VARCHAR"`
	label       string `name:""`
}

// BalanceLog structure holds stock master Balance
type BalanceLog struct {
	table       string    `name:"balance_log" type:"table"`
	AutoID      int64     `json:"auto_id" name:"auto_id" type:"field" sql:"BIGSERIAL PRIMARY KEY"`
	TransDate   time.Time `json:"trans_date" name:"trans_date" type:"field" sql:"TIMESTAMP NOT NULL DEFAULT now()"`
	DoneBy      string    `json:"done_by" name:"done_by" type:"field" sql:"VARCHAR NOT NULL"`
	Type        string    `json:"type" name:"type" type:"field" sql:"VARCHAR NOT NULL"`
	ItemCode    string    `json:"item_code" name:"item_code" type:"field" sql:"VARCHAR NOT NULL"`
	ItemName    string    `json:"item_name" name:"item_name" type:"field" sql:"VARCHAR NOT NULL"`
	Old         float64   `json:"old" name:"old" type:"field" sql:"Float NOT NULL Default '0.0'"`
	New         float64   `json:"new" name:"new" type:"field" sql:"Float NOT NULL Default '0.0'"`
	State       string    `json:"state" name:"state" type:"field" sql:"VARCHAR NOT NULL DEFAULT 'pending'"`
	AdoptedBy   string    `json:"adopted_by" name:"adopted_by" type:"field" sql:"VARCHAR"`
	Branch      string    `json:"branch" name:"branch" type:"field" sql:"VARCHAR NOT NULL"`
	StkLocation string    `json:"stock_location" name:"stock_location" type:"field" sql:"VARCHAR NOT NULL"`
}

// CountedStock structer holding all items that have been counted
type CountedStock struct {
	ItemCode    string  `json:"item_code"`
	ItemName    string  `json:"item_name"`
	CtnUnits    float32 `json:"ctn_units"`
	CtnQty      float32 `json:"ctn_qty"`
	DozQty      float32 `json:"doz_qty"`
	PcsQty      float32 `json:"pcs_qty"`
	TotalPcs    float32 `json:"total_pcs"`
	Poster      string  `json:"poster"`
	Branch      string  `json:"branch"`
	StkLocation string  `json:"stk_location"`
}

// StkBals Holds live stock balances
type StkBals struct {
	Mu           sync.Mutex
	Log          map[string]float64
	ReorderLevel map[string]float64
}

func createStockMasterTbl() error {
	var tblStruct StockMaster
	return database.CreateFromStruct(tblStruct)
}

// GetByCode fetches product by the item code
// Receives a string that represents the item code in search
// and a boolean that chooses wheather to return all or only active stock items
// Queries data from cache
// It returns StockMaster and an error if exists
func GetByCode(key string, all bool, locID int64) (StockMaster, error) {
	start := time.Now()
	defer fmt.Printf("\t\t stockMaster GetByCode took %v \n", time.Since(start))

	var arg StockMaster
	if key == "" {
		return arg, fmt.Errorf("empty search key")
	}

	fmt.Println("\t key =", key)
	// for key, _ := range ProdMaster.ProductDB {
	// 	fmt.Println("\t", key)
	// }
	fmt.Printf("\n\t code trans = %v\n", ProdMaster.Codes)

	// get master code from code translation
	ct, ok := ProdMaster.Codes[key]
	if !ok {
		ct.MasterCode = key
	}
	fmt.Println("\t code translator =", ct.MasterCode)

	// Get combo items
	for i, itm := range ProdMaster.ProductDB[ct.MasterCode].ComboItems {
		ProdMaster.ProductDB[ct.MasterCode].ComboItems[i].ItemName = ProdMaster.ProductDB[itm.ItemCode].ItemName
	}

	// fetch item from product
	arg = ProdMaster.ProductDB[ct.MasterCode]

	arg.PkgQty = ct.PkgQty
	if !all && !arg.IsActive {
		var arg StockMaster
		return arg, nil
	}

	if arg.UnitsPerPack == 0 {
		arg.UnitsPerPack = 1
	}
	fmt.Println("item =", arg)

	arg.VatPercent = ProdMaster.Vats[arg.VatAlpha]
	err := arg.StockCalcs()
	if err != nil {
		return arg, err
	}

	arg.Disc = ct.Discount
	fmt.Println("stock master combo_items =", arg.ComboItems)

	return arg, nil
}

// GetMultiCodes fetches products from a list of codes
// receives a list of codes
// returns a slice of StockMaster and error if fails
func GetMultiCodes(keys []string) ([]StockMaster, error) {
	vals := []StockMaster{}

	if len(keys) == 0 {
		return vals, nil
	}

	for _, val := range ProdMaster.ProductDB {
		for _, code := range keys {
			if val.ItemCode == code {
				vals = append(vals, val)
			}
		}
	}

	return vals, nil
}

// SearchDescription fetches product by the item code
// Queries data from cache
// Receives a string that represents the item name in search
// It returns a slice of StockMaster and an error if exists
func SearchDescription(key string, locID int64) ([]StockMaster, error) {
	// start := time.Now()
	// defer fmt.Printf("\t\t stockMaster SearchName took %v \n", time.Since(start))

	var args []StockMaster
	word_item := strings.Split(key, " ")
	for _, val := range ProdMaster.ProductDB {
		c := 0.0
		for _, name := range word_item {
			if strings.Contains(strings.ToLower(val.ItemName), strings.ToLower(name)) {
				c += 1
			}
		}

		if c == float64(len(word_item)) {
			val.StockCalcs()
			val.Bal = val.Balance[locID]

			args = append(args, val)
		}

		if len(args) >= 50 {
			break
		}
	}

	return args, nil
}

// SearchByCategory Queries a products from a category
// Receives a string param representing category_code
// Returns a slice of products or an error
func SearchByCategory(key string, locID int64) ([]StockMaster, error) {
	var vals []StockMaster

	for _, item := range ProdMaster.ProductDB {
		if fmt.Sprintf("%v", item.DeptCode) == key && item.IsActive {
			item.StockCalcs()

			if item.ItemCode == "" {
				fmt.Println("\t null item")
				continue
			}
			item.Bal = item.Balance[locID]
			if locID == 0 {
				item.Bal = float64(0)
				for _, val := range item.Balance {
					item.Bal += val
				}
			}

			vals = append(vals, item)
		}
	}

	sort.Slice(vals, func(i, j int) bool {
		return vals[i].ItemName < vals[j].ItemName
	})

	return vals, nil
}

// All fetches all products in a given limit
// Receives a string param representing category_code
// Returns a slice of products or an error
func All(limit int) ([]StockMaster, error) {
	var vals []StockMaster

	i := 0
	for _, item := range ProdMaster.ProductDB {

		item.StockCalcs()
		vals = append(vals, item)

		if i >= limit && limit > 0 {
			sort.Slice(vals, func(i, j int) bool { return vals[i].ItemName < vals[j].ItemName })
			return vals, nil
		}

		i++
	}

	sort.Slice(vals, func(i, j int) bool { return vals[i].ItemName < vals[j].ItemName })
	return vals, nil
}

// StockCalcs - calculates product's margins and markups
// returns an error if fails
func (val *StockMaster) StockCalcs() error {
	val.TillPrice = val.ItemSellingprice
	if val.OfferEnd.After(time.Now()) {
		val.TillPrice = val.ItemOfferprice
		val.OnOffer = true
	}

	if val.TillPrice == 0 {
		// val.OnOffer = false
		val.TillPrice = val.ItemSellingprice
	}

	val.Margin = (val.TillPrice / val.ItemCost) - 1
	if val.ItemSellingprice == 0 {
		val.Margin = 1 - (val.ItemCost / 1)
		val.Markup = 1 - (val.ItemCost / 1)
	}
	if val.ItemCost == 0 {
		val.Margin = 1
	}

	val.Markup = (val.ItemCost / val.TillPrice) - 1
	if val.TillPrice == 0 {
		val.Markup = (val.ItemCost / val.ItemSellingprice) - 1

		if val.ItemSellingprice == 0 {
			val.Markup = (val.ItemCost / 1) - 1
		}
	}

	val.Label = val.ItemName

	return nil
}

// CreateNew() creates a new product
// Receives a StockMaster struct
// Adds to database and cache
// Returns an error if fails
func (arg *StockMaster) CreateNew() error {
	fmt.Println("Item Code = ", arg.ItemCode)
	if arg.ItemCode == "" {
		log.Println("failed to create new item. item code is required")
		return errors.New("item code is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tx, err := database.PgPool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// adds to database
	err = arg.AddDB(ctx, tx)
	if err != nil {
		log.Println("error. Unable to Add new Item     err =", err)
		return err
	}

	// add description to database
	err = arg.Description.SQLAdd(ctx, tx)
	if err != nil {
		log.Println("error. Unable to Add new description     err =", err)
		return err
	}

	// add item_name to cache / file in-memory db
	err = ProdMaster.AddProduct(*arg)
	if err != nil {
		log.Println("error. Unable to cache new Item     err =", err)
		return err
	}

	return tx.Commit(ctx)
}

// AddDB adds a new stock_item to database
// returns an error if fails
func (arg *StockMaster) AddDB(ctx context.Context, tx pgx.Tx) error {
	sql := `INSERT INTO stock_master (item_code, item_name, item_cost, item_sellingprice,
				item_wholesaleprice, item_offerprice, offer_end, offer_qty, vat_alpha,
				dept_name, dept_code, manufucturer_code, supplier_code, units_per_pack,
				kg_weight, min_margin, is_serial, is_batched, is_return, return_code, units_per_recipe, is_inventory, sync_servers)
			VALUES($1, $2, $3, $4, $5, $6, now(), $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $22, '{}')
			ON CONFLICT ON CONSTRAINT stock_master_pkey 
			DO UPDATE SET 
			vat_alpha = $8, dept_name = $9, dept_code = $10
			, kg_weight = $14, is_serial = $16, is_batched = $17
			, is_return = $18, units_per_recipe = $20,  is_active = $21, item_name = EXCLUDED.item_name 
			, is_inventory = $22`

	_, err := tx.Exec(ctx, sql, arg.ItemCode, arg.ItemName, arg.ItemCost, arg.ItemSellingprice,
		arg.ItemWholesaleprice, arg.ItemOfferprice, arg.OfferQty, arg.VatAlpha,
		arg.DeptName, arg.DeptCode, arg.ManufucturerCode, arg.SupplierCode, arg.UnitsPerPack,
		arg.KgWeight, arg.MinMargin, arg.IsSerial, arg.IsBatched, arg.IsReturn, arg.ReturnCode, arg.UnitsPerRecipe, arg.IsActive, arg.IsInventory)
	if err != nil {
		log.Printf("error failed to insert item to database    err = %v\n", err)
		return err
	}

	return nil
}

// AddDescription defines a product's identifiables
// Adds into Database product, brand, size, color
