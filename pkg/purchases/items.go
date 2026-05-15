package purchases

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/JohnnyKahiu/speedsales_inventory/database"
	"github.com/JohnnyKahiu/speedsales_inventory/pkg/products"
)

type GrnItem struct {
	table          string    `name:"grn_items" type:"table"`
	AutoID         int64     `json:"auto_id" name:"auto_id" type:"field" sql:"BIGSERIAL PRIMARY KEY"`
	TransDate      time.Time `json:"trans_date" name:"trans_date" type:"field" sql:"TIMESTAMP NOT NULL DEFAULT NOW()"`
	GrnNum         int64     `json:"grn_num" name:"grn_num" type:"field" sql:"BIGINT NOT NULL"`
	CompanyID      int64     `json:"company_id" name:"company_id" type:"field" sql:"BIGINT NOT NULL DEFAULT '0'"`
	InvType        string    `json:"vat_type" type:"field" sql:"VARCHAR NOT NULL"`
	ScanCode       string    `json:"scan_code" type:"field" sql:"VARCHAR NOT NULL"`
	ItemCode       string    `json:"item_code" type:"field" sql:"VARCHAR NOT NULL"`
	ItemName       string    `json:"item_name" type:"field" sql:"VARCHAR"`
	CtnCharged     float64   `json:"ctn_charged" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	DozCharged     float64   `json:"doz_charged" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	PcsCharged     float64   `json:"pcs_charged" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	QtyCharged     float64   `json:"qty_charged" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	QtyReceived    float64   `json:"qty_received" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	ItemCost       float64   `json:"item_cost" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	OldCost        float64   `json:"old_cost" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	ItemPrice      float64   `json:"item_price" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	ItemVat        float64   `json:"item_vat" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	Vat            float64   `json:"vat" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	Vatable        float64   `json:"vatable" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	Excempt        float64   `json:"excempt" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	VatAlpha       string    `json:"vat_alpha" type:"field" sql:"VARCHAR(2) NOT NULL"`
	DiscountType   string    `json:"discount_type" type:"field" sql:"VARCHAR"`
	Discount       float64   `json:"discount" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	DiscountVal    float64   `json:"discount_val" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	QtyDiscount    float64   `json:"qty_discount" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	TotalAmount    float64   `json:"total_amount" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	TotalAmountInc float64   `json:"total_amount_inc" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	FinID          int64     `json:"fin_id" type:"field" sql:"BIGINT"`
	ReturnNum      string    `json:"return_num" type:"field" sql:"VARCHAR"`
	ReturnQty      float64   `json:"return_qty" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	NetQty         float64   `json:"net_qty" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	NetAmount      float64   `json:"net_amount" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	Poster         string    `json:"poster" type:"field" sql:"VARCHAR"`
	State          string    `json:"state" type:"field" sql:"VARCHAR NOT NULL DEFAULT 'pending'"`
	InvState       string    `json:"inv_state" type:"field" sql:"VARCHAR NOT NULL DEFAULT 'pending'"`
	LocationID     int64     `json:"location_id" type:"field" sql:"INT NOT NULL DEFAULT '1'"`
	ReceivedValue  float64   `json:"received_value" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	grnfkey        string    `name:"grn_foreign_key" type:"constraint" sql:"FOREIGN KEY (grn_num) REFERENCES grn_log(grn_num) "`
	locationFkey   string    `name:"loc_foreign_key" type:"constraint" sql:"FOREIGN KEY (location_id) REFERENCES stock_locations(auto_id)"`
	OverallTotal   float64   `json:"overall_total"`
	VatPercent     float64   `json:"vat_percent"`
	Label          string    `json:"label"`
	CostTolerance  float64
}

func genGrnItems() error {
	return database.CreateFromStruct(GrnItem{})
}

// Write - adds data to database
// Inserts data to grn_items table
// returns an error if it fails
func (a *GrnItem) Write(ctxt context.Context) error {
	sql := `
			INSERT INTO grn_items(grn_num, vat_type, 
				item_code, scan_code, item_name, qty_charged, item_cost, old_cost, item_price, 
				item_vat, vat_alpha, vat, 
				discount_type, discount, qty_discount, discount_val, total_amount, total_amount_inc, 
				net_qty, vatable, excempt, return_qty, 
				state, location_id)
			VALUES($1, $2,
					$3, $4, $5, $6, $7, $8, $9, 
					$10, $11, $12, 
					$13, $14, $15, $16, $17, $18, 
					$19, $20, $21, $22, 
					$23, $24)
			RETURNING auto_id`

	ctx, cancel := context.WithTimeout(ctxt, 20*time.Second)
	defer cancel()

	rows, err := database.PgPool.Query(ctx, sql, a.GrnNum, a.InvType,
		a.ItemCode, a.ScanCode, a.ItemName, a.QtyCharged, a.ItemCost, a.OldCost, a.ItemPrice,
		a.ItemVat, a.VatAlpha, a.Vat,
		a.DiscountType, a.Discount, a.QtyDiscount, a.DiscountVal, a.TotalAmount, a.TotalAmountInc,
		a.NetQty, a.Vatable, a.Excempt, a.ReturnQty,
		a.State, a.LocationID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&a.AutoID)
		if err != nil {
			log.Println("scan error    failed to scan grn_item     err =", err)
			return err
		}
	}

	return nil
}

// AddItem adds a user input into database
// resolves item and adds into grn_items table
// returns an error if it fails
func (a *GrnItem) AddItem(ctxt context.Context) error {
	if err := a.enrichFromGrnLog(ctxt); err != nil {
		return err
	}

	if err := a.resolveProduct(ctxt); err != nil {
		return err
	}

	if err := a.validateQty(ctxt); err != nil {
		return err
	}

	// calculate costs
	if err := a.calculateCosts(ctxt); err != nil {
		return err
	}
	// calculate VAT
	if err := a.calculateVATS(ctxt); err != nil {
		return err
	}

	// determine invoice state
	if err := a.determineState(ctxt); err != nil {
		return err
	}

	// write to database
	return a.Write(ctxt)
}

func (a *GrnItem) enrichFromGrnLog(ctxt context.Context) error {
	if a.GrnNum == 0 {
		return errors.New("invalid grn_num")
	}
	grnLog := GrnLog{GrnNum: a.GrnNum}
	if err := grnLog.Details(ctxt); err != nil {
		log.Println("failed to get grn_log details     err =", err)
		return err
	}

	a.InvType = grnLog.InvType
	return nil
}

// resolveProduct - populates grn_item from products
func (a *GrnItem) resolveProduct(ctxt context.Context) error {
	if a.ItemCode == "" {
		return errors.New("item's code is null")
	}

	prDets, err := products.GetByCode(a.ItemCode, true, a.LocationID)
	if err != nil {
		return errors.New("failed to resolve item code")
	}

	// enrich details
	a.ItemName = prDets.ItemName
	a.OldCost = prDets.ItemCost
	a.ItemPrice = prDets.TillPrice
	a.VatAlpha = prDets.VatAlpha
	a.VatPercent = prDets.VatPercent

	if err := a.calculateCosts(ctxt); err != nil {
		return err
	}

	return nil
}

// resolveProduct - populates grn_item from products
func (a *GrnItem) calculateCosts(ctxt context.Context) error {
	if a.NetQty == 0 {
		a.NetQty = a.QtyCharged + a.QtyDiscount - a.ReturnQty
	}

	a.ItemCost = a.TotalAmountInc / a.QtyCharged

	// calculate totals based on invType
	amountInc := a.TotalAmount
	if a.InvType == "exclusive" {
		amountInc = (a.TotalAmount * (100 + a.VatPercent)) / 100
		return nil
	}
	if a.InvType == "return" {
		a.ItemCost = a.OldCost
		amountInc = a.ItemCost * a.ReturnQty
		a.TotalAmount = amountInc
		return nil
	}

	a.TotalAmountInc = amountInc
	return nil
}

// resolveProduct - populates grn_item from products
func (arg *GrnItem) calculateVATS(ctxt context.Context) error {
	// calculate vat from total amount that is inclusive VAT
	if arg.VatPercent == 0 {
		arg.Excempt = arg.TotalAmount
		arg.Vatable = 0
		arg.Vat = 0
		return nil
	}

	arg.Excempt = 0
	arg.Vatable = (arg.TotalAmountInc * 100) / (100 + arg.VatPercent)
	arg.Vat = (arg.TotalAmountInc * arg.VatPercent) / (100 + arg.VatPercent)

	return nil
}

// resolveProduct - determines grn_item from products
func (arg *GrnItem) determineState(ctxt context.Context) error {
	arg.State = "pending"

	discountedAmount := (arg.Discount / 100) * arg.ItemCost

	calculatedCost := arg.ItemCost + discountedAmount
	withinTolerance := calculatedCost >= (arg.OldCost-arg.CostTolerance) && calculatedCost <= (arg.OldCost+arg.CostTolerance)
	if !withinTolerance {
		arg.State = "PRICE CHANGE"
	}
	return nil
}

// validate quantities
func (arg *GrnItem) validateQty(ctxt context.Context) error {
	if arg.InvType == "purchase_simplified" {
		arg.NetQty = arg.QtyCharged
	}

	if arg.NetQty > arg.QtyCharged {
		arg.DiscountType = "quantity"
		arg.QtyDiscount = arg.NetQty - arg.QtyCharged
	}

	return nil
}

// Delete - removes item from cart
// Updates state to VOIDED
// returns an error if fails
func (arg *GrnItem) Delete(ctxt context.Context) error {
	ctx, cancel := context.WithTimeout(ctxt, 15*time.Second)
	defer cancel()

	sql := ` DELETE FROM grn_items WHERE auto_id = $1 `
	_, err := database.PgPool.Exec(ctx, sql, arg.AutoID)
	if err != nil {
		log.Println("error.  failed to delete grn_item     err =", err)
		return err
	}

	fmt.Printf("deleted item auto_id = %v successfully\n", arg.AutoID)
	return nil
}
