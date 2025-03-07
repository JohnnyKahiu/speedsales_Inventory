package products

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/JohnnyKahiu/speedsales_inventory/database"
	"github.com/JohnnyKahiu/speedsales_inventory/pkg/variables"
)

type ProdDB struct {
	ProductDB map[string]StockMaster    `json:"product_db"`
	Recipe    map[string][]Recipe       `json:"recipe"`
	Codes     map[string]CodeTranslator `json:"codes"`
	Vats      map[string]float64        `json:"vats"`
	mx        sync.RWMutex
}

var ProdMaster ProdDB

func (arg *ProdDB) LoadStockMaster() error {
	sql := `SELECT 
				st_mas.item_code as item_code
				, st_mas.item_name as item_name
				, cast(st_mas.item_sellingprice as float) as item_sellingprice 
				, cast(st_mas.item_cost as float) as item_cost
				, cast(st_mas.item_wholesaleprice as float) as item_wholesaleprice
				, cast(coalesce(st_mas.item_offerprice, 0) as float) as item_offerprice
				, st_mas.offer_start as offer_start
				, st_mas.offer_end as offer_end
				, cast(coalesce(st_mas.offer_qty, 0) as float) as offer_qty
				, st_mas.vat_alpha as Vat_alpha
				, coalesce(st_mas.units_per_pack::int, 1) as units
				, st_mas.dept_code
				, st_mas.dept_name as dept
				, coalesce(st_mas.manufucturer_code, 1)  as manufucturer_code
				, st_mas.supplier_code as supp_code
				, coalesce(supp.name, 'uknown') as supp_name
				, st_mas.is_batched
				, st_mas.is_serial
				, st_mas.is_return
				, coalesce(st_mas.return_code, '') as return_code
				, cast(coalesce(st_mas.price_effect_time, '2020-01-01 00:00:00') as timestamp) as price_effect_time
				, cast(st_mas.kg_weight as float) as kg_weight
				, st_mas.is_produced 
				, st_mas.units_per_recipe
				, coalesce(st_mas.image, '../../static/images/')
				, st_mas.is_active
				, st_mas.is_inventory
				
			FROM stock_master st_mas
			LEFT JOIN suppliers as supp ON supp.cr_id = st_mas.supplier_code
			`

	rows, err := database.PgPool.Query(context.Background(), sql)
	if err != nil {
		log.Println("error. failed to fetch products from database     err =", err)
		return err
	}
	defer rows.Close()

	arg.ProductDB = make(map[string]StockMaster)
	for rows.Next() {
		var r StockMaster
		err := rows.Scan(&r.ItemCode, &r.ItemName, &r.ItemSellingprice, &r.ItemCost, &r.ItemWholesaleprice, &r.ItemOfferprice, &r.OfferStart, &r.OfferEnd, &r.OfferQty,
			&r.VatAlpha, &r.UnitsPerPack, &r.DeptCode, &r.DeptName, &r.ManufucturerCode, &r.SupplierCode, &r.ManufucturerName, &r.IsBatched, &r.IsSerial, &r.IsReturn, &r.ReturnCode,
			&r.PriceEffectTime, &r.KgWeight, &r.IsProduced, &r.UnitsPerRecipe, &r.Image, &r.IsActive, &r.IsInventory)
		if err != nil {
			log.Println("error. failed to scan stock master    err =", err)
			return err
		}
		r.Label = r.ItemCode

		arg.ProductDB[r.ItemCode] = r
	}

	err = arg.Pickle()
	if err != nil {
		log.Println("error failed to pickle    err =", err)
		return err
	}
	return nil
}

func (arg *ProdDB) LoadRecipe() error {
	sql := `SELECT 
				
				
			FROM stock_master st_mas
			LEFT JOIN suppliers as supp ON supp.cr_id = st_mas.supplier_code
			`

	rows, err := database.PgPool.Query(context.Background(), sql)
	if err != nil {
		log.Println("error. failed to fetch products from database     err =", err)
		return err
	}
	defer rows.Close()

	arg.ProductDB = make(map[string]StockMaster)
	for rows.Next() {
		var r StockMaster
		err := rows.Scan(&r.ItemCode, &r.ItemName, &r.ItemSellingprice, &r.ItemCost, &r.ItemWholesaleprice, &r.ItemOfferprice, &r.OfferStart, &r.OfferEnd, &r.OfferQty,
			&r.VatAlpha, &r.UnitsPerPack, &r.DeptCode, &r.DeptName, &r.ManufucturerCode, &r.SupplierCode, &r.ManufucturerName, &r.IsBatched, &r.IsSerial, &r.IsReturn, &r.ReturnCode,
			&r.PriceEffectTime, &r.KgWeight, &r.IsProduced, &r.UnitsPerRecipe, &r.Image, &r.IsActive)
		if err != nil {
			log.Println("error. failed to scan stock master    err =", err)
			return err
		}
		r.Label = r.ItemCode

		arg.ProductDB[r.ItemCode] = r
	}

	err = arg.Pickle()
	if err != nil {
		log.Println("error failed to pickle    err =", err)
		return err
	}
	return nil
}

func (arg *ProdDB) LoadCodeTranslator() error {
	sql := `SELECT
				master_code
				, link_code
				, pkg_qty
				, discount
			FROM code_translator`

	rows, err := database.PgPool.Query(context.Background(), sql)
	if err != nil {
		log.Println("error. failed to get code_translator    err =", err)
		return err
	}
	defer rows.Close()

	arg.Codes = make(map[string]CodeTranslator)
	for rows.Next() {
		var r CodeTranslator
		err := rows.Scan(&r.MasterCode, &r.LinkCode, &r.PkgQty, &r.Discount)
		if err != nil {
			log.Println("LoadCodeTranslator: error. failed to scan from code_translator    err =", err)
			return err
		}

		arg.Codes[r.LinkCode] = r
	}

	err = arg.Pickle()
	if err != nil {
		log.Println("LoadCodeTranslator: error. failed to pickle    err =", err)
		return err
	}
	return nil
}

// AddProduct adds a new product to Cache
func (arg *ProdDB) AddProduct(val StockMaster) error {
	start := time.Now()
	defer fmt.Printf("\t PridMaster AddProduct took %v\n", time.Since(start))

	update := false
	var offerStart time.Time
	var offerEnd time.Time

	price := float64(0)
	offer := float64(0)
	offerQty := float64(0)

	if val.ItemCode == "" {
		return fmt.Errorf("cannot add a null item. item_code is empty")
	}

	var comboItems []Combo
	if val, ok := arg.ProductDB[val.ItemCode]; ok {
		update = true
		price = val.ItemSellingprice
		offer = val.ItemOfferprice
		offerEnd = val.OfferEnd
		offerStart = val.OfferStart
		offerQty = float64(val.OfferQty)
		comboItems = val.ComboItems
	}

	if update {
		val.ItemSellingprice = price
		val.ItemOfferprice = offer
		val.OfferEnd = offerEnd
		val.OfferStart = offerStart
		val.OfferQty = offerQty
		val.Label = val.ItemName
		val.ComboItems = comboItems
	}
	arg.ProductDB[val.ItemCode] = val

	var ct CodeTranslator
	ct.MasterCode = val.ItemCode
	ct.LinkCode = val.ItemCode
	ct.PkgQty = 1
	ct.Discount = 0

	arg.Codes[val.ItemCode] = ct
	fmt.Printf("Code translation  %v\n", arg.Codes[val.ItemCode])

	// pickle and save changes
	err := arg.Pickle()
	if err != nil {
		fmt.Println("error failed to pickle    err =", err)
		return err
	}
	fmt.Println("\t AddProduct completed  successfully")
	return nil
}

func (arg *ProdDB) UpdateLinks(ct CodeTranslator) error {
	start := time.Now()
	defer fmt.Printf("\tStockMaster UpdateLinks took %v\n", time.Since(start))

	arg.Codes[ct.LinkCode] = ct

	err := arg.Pickle()
	if err != nil {
		log.Println("error. UpdateLinks failed to pickle    err =", err)
		return err
	}
	return nil
}

func (arg *ProdDB) BulkUpdateLinks(cts []CodeTranslator) error {
	start := time.Now()
	defer fmt.Printf("\tStockMaster UpdateLinks took %v\n", time.Since(start))
	if cts == nil {
		return nil
	}

	for _, ct := range cts {
		if arg.Codes == nil {
			arg.Codes = make(map[string]CodeTranslator)
		}
		arg.Codes[ct.LinkCode] = ct

		err := arg.Pickle()
		if err != nil {
			log.Println("error. UpdateLinks failed to pickle    err =", err)
			return err
		}
	}
	return nil
}

func (arg *ProdDB) FetchAll() ([]StockMaster, error) {
	var vals []StockMaster

	for _, item := range arg.ProductDB {
		vals = append(vals, item)
	}
	return vals, nil
}

// DelLink function removes link between code and master code
func (arg *ProdDB) DelLink(link string) error {
	start := time.Now()
	defer fmt.Printf("\tStockMaster delLink took %v\n", time.Since(start))

	delete(arg.Codes, link)

	err := arg.Pickle()
	if err != nil {
		log.Println("error failed to pickle delLink    err =", err)
		return err
	}

	return nil
}

func (arg *ProdDB) Pickle() error {
	buf := &bytes.Buffer{}
	err := gob.NewEncoder(buf).Encode(arg)
	if err != nil {
		log.Println("error'    failed to pickle products database     err =", err)
	}

	arg.mx.Lock()
	defer arg.mx.Unlock()
	picklePath := filepath.Join(variables.FDBPath, "products.bin")
	fmt.Println("pickle path =", picklePath)

	err = os.WriteFile(picklePath, buf.Bytes(), 0666)
	if err != nil {
		log.Println("error.    failed to write products file     err =", err)
		return err
	}
	return nil
}

func (arg *ProdDB) LoadFromDB() error {
	err := arg.LoadStockMaster()
	if err != nil {
		log.Println("error. failed to load from stock_master    err =", err)
		return err
	}
	fmt.Println("\t\t Loaded stock_master")

	err = arg.LoadCodeTranslator()
	if err != nil {
		log.Println("error. failed to load code translator    err =", err)
		return err
	}
	fmt.Println("\t\t Loaded code translations")

	err = arg.Pickle()
	if err != nil {
		log.Println("error failed to pickle    err =", err)
		return err
	}
	fmt.Println("\t\t pickled Products_DB")
	return nil
}

func (arg *ProdDB) Read() error {
	picklePath := filepath.Join(variables.FDBPath, "products.bin")

	rd, err := os.ReadFile(picklePath)
	if err != nil {
		err = arg.LoadFromDB()
		if err != nil {
			log.Println("error. failed to Load data from db    err =", err)
			return err
		}
	}

	if rd == nil || len(rd) <= 1 {
		err = arg.LoadStockMaster()
		if err != nil {
			log.Println("error. failed to load stock master    err =", err)
			return err
		}
	}

	err = gob.NewDecoder(bytes.NewReader(rd)).Decode(&arg)
	if err != nil {
		log.Println("error. failed to decode prods    err =", err)
		return err
	}
	return nil
}

func (arg *ProdDB) NewVat(vals map[string]float64) error {
	arg.Vats = vals

	arg.Pickle()
	return nil
}

func (arg *ProdDB) Del(key string) error {
	delete(arg.ProductDB, key)

	err := arg.Pickle()
	if err != nil {
		return err
	}
	return nil
}

func (arg *ProdDB) Merge(key string) error {
	delete(arg.ProductDB, key)
	return nil
}

func createIfNotExists(path string) bool {
	isCreate := true
	// check and create folder
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		isCreate = false
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}

	return isCreate
}

func (arg *ProdDB) SampleDB() error {
	jsonFile, err := os.Open("/home/johnny/Dropbo/projects/speed_sales/server/products/stock_master_model.json")
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println("\n\njson open error    err =", err)
		return err
	}
	byteValue, _ := io.ReadAll(jsonFile)

	json.Unmarshal(byteValue, &arg)

	return nil
}

// 671 667
