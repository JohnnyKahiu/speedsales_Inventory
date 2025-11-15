package products

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// StockMaster Holds information about a stock item
type StockMaster struct {
	table              string    `type:"table" name:"stock_master"`
	ItemCode           string    `json:"item_code" type:"field" name:"item_code" sql:"VARCHAR PRIMARY KEY"`
	ItemName           string    `json:"item_name" type:"field" name:"item_name" sql:"VARCHAR NOT NULL"`
	ItemCost           float64   `json:"item_cost" type:"field" name:"item_cost" sql:"DECIMAL NOT NULL DEFAULT '0.0'"`
	ItemSellingprice   float64   `json:"item_sellingprice" type:"field" name:"item_sellingprice" sql:"DECIMAL NOT NULL DEFAULT '0.0'"`
	ItemWholesaleprice float64   `json:"item_wholesaleprice" type:"field" name:"item_wholesaleprice" sql:"DECIMAL NOT NULL DEFAULT '0.0'"`
	ItemOfferprice     float64   `json:"item_offerprice" type:"field" name:"item_offerprice" sql:"DECIMAL NOT NULL DEFAULT '0.0'"`
	OfferStart         time.Time `json:"offer_start" type:"field" name:"offer_start" sql:"TIMESTAMP NOT NULL DEFAULT NOW()"`
	OfferEnd           time.Time `json:"offer_end" type:"field" name:"offer_end" sql:"TIMESTAMP NOT NULL DEFAULT NOW()"`
	OfferQty           float64   `json:"offer_qty" type:"field" name:"offer_qty" sql:"FLOAT NOT NULL DEFAULT '0'"`
	VatAlpha           string    `json:"vat_alpha" type:"field" name:"vat_alpha" sql:"VARCHAR(5) NOT NULL"`
	DeptName           string    `json:"dept_name" type:"field" name:"dept_name" sql:"VARCHAR NOT NULL"`
	DeptCode           int32     `json:"dept_code" type:"field" name:"dept_code" sql:"INTEGER NOT NULL"`
	SupplierCode       int64     `json:"supplier_code" type:"field" name:"supplier_code" sql:"INTEGER NOT NULL DEFAULT '1'"`
	ManufucturerCode   int64     `json:"manufucturer_code" type:"field" name:"manufucturer_code" sql:"INTEGER NOT NULL DEFAULT '0'"`
	ManufucturerName   string    `json:"manufucturer_name" type:"field" name:"manufucturer_name" sql:"VARCHAR NOT NULL DEFAULT 'undefined'"`
	UnitsPerPack       int       `json:"units_per_pack" type:"field" name:"units_per_pack" sql:"VARCHAR NOT NULL"`
	KgWeight           float64   `json:"kg_weight" name:"kg_weight" type:"field" sql:"FLOAT NOT NULL"`
	OldPrice           bool      `json:"old_price" name:"old_price" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	PriceChange        bool      `json:"price_change" type:"field" name:"price_change" sql:"BOOLEAN NOT NULL DEFAULT False"`
	PriceChangeDate    time.Time `json:"price_change_date" type:"field" name:"price_change_date" sql:"TIMESTAMP"`
	PriceEffectTime    time.Time `json:"price_effect_time" type:"field" name:"price_effect_time" sql:"TIMESTAMP"`
	PriceEffectLead    int32     `json:"price_effect_lead" type:"field" name:"price_effect_lead" sql:"INTEGER"`
	IsInventory        bool      `json:"is_inventory" type:"field" name:"is_inventory" sql:"BOOLEAN NOT NULL DEFAULT True"`
	IsSerial           bool      `json:"is_serial" type:"field" name:"is_serial" sql:"BOOLEAN NOT NULL DEFAULT False"`
	IsBatched          bool      `json:"is_batched" type:"field" name:"is_batched" sql:"BOOLEAN NOT NULL DEFAULT False"`
	IsReturn           bool      `json:"is_return" type:"field" name:"is_return" sql:"BOOLEAN NOT NULL DEFAULT False"`
	ReturnCode         string    `json:"return_code" type:"field" name:"return_code" sql:"VARCHAR"`
	UpdateTime         time.Time `json:"update_time" type:"field" name:"update_time" sql:"TIMESTAMP NOT NULL DEFAULT NOW()"`
	MinMargin          float32   `json:"min_margin" type:"field" name:"min_margin" sql:"FLOAT NOT NULL DEFAULT '0.05'"`
	IsActive           bool      `json:"is_active" type:"field" name:"is_active" sql:"BOOL NOT NULL DEFAULT True"`
	IsProduced         bool      `json:"is_produced" type:"field" name:"is_produced" sql:"BOOL NOT NULL DEFAULT False"`
	Category           string    `json:"category" name:"category" type:"field" sql:"VARCHAR NOT NULL DEFAULT 'sale item'"`
	UnitsPerRecipe     float64   `json:"units_per_recipe" name:"units_per_recipe" type:"field" sql:"FLOAT NOT NULL DEFAULT '1'"`
	ReorderLevel       float64   `json:"reorder_level" type:"field" name:"reorder_level" sql:"FLOAT NOT NULL DEFAULT '0'"`
	Image              string    `json:"image" name:"image" type:"field" sql:"VARCHAR"`
	SyncServers        []string  `json:"sync_servers" name:"sync_servers" type:"field" sql:"VARCHAR[]"`
	IsCombo            bool      `json:"is_combo" type:"field" name:"is_combo" sql:"BOOLEAN NOT NULL DEFAULT False"`
	ComboItems         []Combo   `json:"combo_items" name:"combo_items" type:"field" sql:"JSONB"`
	OnOffer            bool      `json:"on_offer"`
	TillPrice          float64   `json:"till_price"`
	VatPercent         float64   `json:"vat_percent"`
	Margin             float64   `json:"margin"`
	Markup             float64   `json:"mark_up"`
	PkgQty             float64   `json:"pkg_qty"`
	Disc               float64   `json:"Disc"`
	Label              string    `json:"label"`
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

// CodeTranslator structure holds translation of master and linked codes
type CodeTranslator struct {
	table       string  `name:"code_translator" type:"table"`
	AutoID      int64   `json:"auto_id" name:"auto_id" type:"field" sql:"BIGSERIAL PRIMARY KEY"`
	MasterCode  string  `json:"master_code" name:"master_code" type:"field" sql:"VARCHAR(30) NOT NULL"`
	LinkCode    string  `json:"link_code" name:"link_code" type:"field" sql:"VARCHAR(30) NOT NULL"`
	PkgQty      float64 `json:"pkg_qty" name:"pkg_qty" type:"field" sql:"INT NOT NULL DEFAULT '1'"`
	Discount    float64 `json:"discount" name:"discount" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	constraint  int     `name:"code_translator_pkey" type:"constraint" sql:"PRIMARY KEY(link_code, master_code)"`
	constraint2 string  `name:"code_translator_link_code_key" type:"constraint" sql:"UNIQUE(link_code)"`
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

// Departments structure holding all products departments
type Departments struct {
	table       string `name:"departments" type:"table"`
	DeptCode    int64  `json:"dept_code" name:"dept_code" type:"field" sql:"SERIAL UNIQUE"`
	DeptName    string `json:"dept_name" name:"dept_name" type:"field" sql:"VARCHAR NOT NULL"`
	SubDeptName string `json:"sub_dept_name" name:"sub_dept_name" type:"field" sql:"VARCHAR NOT NULL"`
	Label       string `json:"label"`
	MinMargin   string `json:"min_margin" name:"min_margin" type:"field" sql:"VARCHAR NOT NULL DEFAULT '0.5'"`
	composite   string `name:"departments_pkey" type:"constraint" sql:"PRIMARY KEY(dept_name, sub_dept_name)"`
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

// GetByCode fetches product by the item code
// Receives a string that represents the item code in search
// and a boolean that chooses wheather to return all or only active stock items
// Queries data from cache
// It returns StockMaster and an error if exists
func GetByCode(key string, all bool) (StockMaster, error) {
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

	// get master code from code translation
	ct, ok := ProdMaster.Codes[key]
	if !ok {
		ct.MasterCode = key
	}
	// fmt.Println("code translator =", ct)

	for i, itm := range ProdMaster.ProductDB[ct.MasterCode].ComboItems {
		ProdMaster.ProductDB[ct.MasterCode].ComboItems[i].ItemName = ProdMaster.ProductDB[itm.ItemCode].ItemName
	}

	arg = ProdMaster.ProductDB[ct.MasterCode]
	// fmt.Println("item =", arg)
	arg.PkgQty = ct.PkgQty
	if !all && !arg.IsActive {
		var arg StockMaster
		return arg, nil
	}

	if arg.UnitsPerPack == 0 {
		arg.UnitsPerPack = 1
	}

	arg.VatPercent = ProdMaster.Vats[arg.VatAlpha]
	arg.StockCalcs()
	arg.Disc = ct.Discount
	fmt.Println("stock master combo_items =", arg.ComboItems)

	return arg, nil
}

// SearchDescription fetches product by the item code
// Queries data from cache
// Receives a string that represents the item name in search
// It returns a slice of StockMaster and an error if exists
func SearchDescription(key string) ([]StockMaster, error) {
	// start := time.Now()
	// defer fmt.Printf("\t\t stockMaster SearchName took %v \n", time.Since(start))

	var args []StockMaster
	word_item := strings.Split(key, " ")
	for _, val := range ProdMaster.ProductDB {
		c := 0
		for _, name := range word_item {
			if strings.Contains(strings.ToLower(val.ItemName), strings.ToLower(name)) {
				c += 1
			}
		}

		if c == len(word_item) {
			val.StockCalcs()
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
func SearchByCategory(key string) ([]StockMaster, error) {
	var vals []StockMaster

	for _, item := range ProdMaster.ProductDB {
		if fmt.Sprintf("%v", item.DeptCode) == key && item.IsActive {
			item.StockCalcs()

			/*fmt.Printf("\n\t========== %v ========\n", item.ItemName)
			fmt.Printf("\titem_sellingprice = %v\n", item.ItemSellingprice)
			fmt.Printf("\titem_cost         = %v\n", item.ItemCost)
			fmt.Printf("\tmargin            = %v\n", item.Margin)
			fmt.Printf("\tmarkup            = %v\n", item.Markup)*/

			vals = append(vals, item)
		}
	}

	sort.Slice(vals, func(i, j int) bool { return vals[i].ItemName < vals[j].ItemName })

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
