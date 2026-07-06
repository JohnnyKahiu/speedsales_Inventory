package purchases

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/JohnnyKahiu/speedsales_inventory/database"
	"github.com/JohnnyKahiu/speedsales_inventory/pkg/balances"
	"github.com/jackc/pgx/v5"
)

type GrnLog struct {
	table          string    `name:"grn_log" type:"table"`
	TransDate      time.Time `json:"trans_date" type:"field" sql:"TIMESTAMPTZ NOT NULL DEFAULT now()"`
	GrnNum         int64     `json:"grn_num" type:"field" sql:"BIGINT NOT NULL"`
	SuppPin        string    `json:"supp_pin" type:"field" sql:"VARCHAR(20) NOT NULL"`
	SuppName       string    `json:"supp_name" type:"field" sql:"VARCHAR NOT NULL"`
	InvNum         string    `json:"inv_num" type:"field" sql:"VARCHAR NOT NULL"`
	InvType        string    `json:"inv_type" type:"field" sql:"VARCHAR NOT NULL"`
	TimsCUIN       string    `json:"tims_cuin" type:"field" sql:"VARCHAR NOT NULL DEFAULT ''"`
	TotalExc       float64   `json:"total_exc" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	TotalAmount    float64   `json:"total_amount" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	TotalAmountInc float64   `json:"total_amount_inc" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	TotalVat       float64   `json:"total_vat" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	Discount       float64   `json:"discount" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	Vatable        float64   `json:"vatable" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	Exempt         float64   `json:"exempt" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	VehicleNum     string    `json:"vehicle_num" type:"field" sql:"VARCHAR NOT NULL DEFAULT 'nan'"`
	DriverName     string    `json:"driver_name" type:"field" sql:"VARCHAR NOT NULL DEFAULT 'nan'"`
	RegisteredBy   string    `json:"registered_by" type:"field" sql:"VARCHAR NOT NULL"`
	Poster         string    `json:"poster" type:"field" sql:"VARCHAR NOT NULL DEFAULT 'nan'"`
	InvDate        string    `json:"inv_date" type:"field" sql:"TIMESTAMPTZ"`
	RecvDate       string    `json:"recv_date" type:"field" sql:"TIMESTAMPTZ"`
	State          string    `json:"state" type:"field" sql:"VARCHAR NOT NULL DEFAULT 'pending'"`
	pKey           string    `name:"grn_pkey" type:"constraint" sql:"PRIMARY KEY (grn_num)"`
	Items          []GrnItem `json:"items"`
	Branch         string    `json:"branch"`
}

func GenPurchaseTbl() error {
	if err := genGrnLog(); err != nil {
		log.Println("error. failed to gen grn_log table      err =", err)
		return err
	}

	return genGrnItems()
}

func genGrnLog() error {
	return database.CreateFromStruct(GrnLog{})
}

func (arg *GrnLog) getGrnNum(ctxt context.Context) error {
	sql := `SELECT 
				CONCAT(
					EXTRACT(YEAR FROM now())
					,EXTRACT(MONTH FROM now())
					, EXTRACT(DAY FROM now())
					,'03'
					, COUNT(*)+1
				)::BIGINT as grn_num
			FROM grn_log `

	ctx, cancel := context.WithTimeout(ctxt, 15*time.Second)
	defer cancel()

	rows, err := database.PgPool.Query(ctx, sql)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&arg.GrnNum)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetGrn
func (arg *GrnLog) GetGrn(ctxt context.Context) error {
	exists, err := arg.invoiceIsExists(ctxt)
	if err != nil {
		return err
	}
	// reject if exists
	if exists {
		return fmt.Errorf("invoice exists")
	}

	layout := "2006-01-02"

	invdate, err := time.Parse(layout, arg.InvDate)
	recvdate, err := time.Parse(layout, arg.RecvDate)

	// reject if invoice received befor it was invoiced
	if invdate.After(recvdate) {
		return fmt.Errorf("invoice date is after receive date")
	}

	if err := arg.getGrnNum(ctxt); err != nil {
		log.Println("error. failed to generate grn_num     err =", err)
		return err
	}

	if err := arg.WriteNew(ctxt); err != nil {
		log.Println("error. failed to writeNew log     err =", err)
		return err
	}

	return nil
}

// GetReceipt = fetches full receipt details from Grn
// fetches data from grn_log and grn_items
// populates struct and
// returns an error if it fails
func (arg *GrnLog) GetReceipt(ctxt context.Context) error {
	if err := arg.Details(ctxt); err != nil {
		log.Println("failed to fetch grn_log details    err =", err)
		return err
	}

	if err := arg.GetItems(ctxt); err != nil {
		log.Println("failed to fetch grn_items     err =", err)
		return err
	}
	return nil
}

// WriteNew - creates a new purchase_log
// Inserts into purchase_log table
// returns an error if it fails
func (arg *GrnLog) WriteNew(ctxt context.Context) error {
	sql := `
		INSERT INTO grn_log(grn_num, inv_type, trans_date, inv_date, recv_date, 
							supp_pin, supp_name, inv_num, tims_cuin, total_exc, total_amount_inc, 
							total_vat, registered_by)
		VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	ctx, cancel := context.WithTimeout(ctxt, 20*time.Second)
	defer cancel()

	_, err := database.PgPool.Exec(ctx, sql, arg.GrnNum, arg.InvType, arg.TransDate, arg.InvDate, arg.RecvDate,
		arg.SuppPin, arg.SuppName, arg.InvNum, arg.TimsCUIN, arg.TotalExc, arg.TotalAmountInc,
		arg.TotalVat, arg.RegisteredBy)
	if err != nil {
		log.Println("postgresql error.  failed to register new invoice     err =", err)
		return err
	}
	return nil
}

// invoiceIsExists - checks whether an invoice exists or not
// Queries a count by supplier_pin and invoice_num from purchase_log
// returns a bool and an error if it fails
func (arg *GrnLog) invoiceIsExists(ctxt context.Context) (bool, error) {
	sql := `SELECT
				count(*) 
			FROM grn_log 
			WHERE supp_pin = $1 
				AND inv_num = $2 
				AND state NOT IN ('VOIDED', 'DELETED')`

	ctx, cancel := context.WithTimeout(ctxt, 15*time.Second)
	defer cancel()

	rows, err := database.PgPool.Query(ctx, sql, arg.SuppPin, arg.InvNum)
	if err != nil {
		return true, nil
	}
	defer rows.Close()

	var cnt int
	for rows.Next() {
		err = rows.Scan(&cnt)
		if err != nil {
			log.Printf("error scanning invoice from grn log err = %v\n", err)
		}
	}

	if cnt > 0 {
		return true, nil
	}
	return false, nil
}

// Details fetches
func (arg *GrnLog) Details(ctxt context.Context) error {
	sql := `
		SELECT 
			trans_date, grn_num, supp_name, inv_num, supp_pin
			, inv_type, tims_cuin, total_exc, total_amount_inc, total_vat, discount
			, vatable, exempt, vehicle_num, driver_name, registered_by, poster
			, inv_date, recv_date, state
		FROM grn_log WHERE grn_num = $1		
	`

	ctx, cancel := context.WithTimeout(ctxt, 20*time.Second)
	defer cancel()

	rows, err := database.PgPool.Query(ctx, sql, arg.GrnNum)
	if err != nil {
		log.Println("postgresql error.   failed to query grn_log details     err =", err)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&arg.TransDate, &arg.GrnNum, &arg.SuppName, &arg.InvNum, &arg.SuppPin,
			&arg.InvType, &arg.TimsCUIN, &arg.TotalExc, &arg.TotalAmountInc, &arg.TotalVat, &arg.Discount,
			&arg.Vatable, &arg.Exempt, &arg.VehicleNum, &arg.DriverName, &arg.RegisteredBy, &arg.Poster,
			&arg.InvDate, &arg.RecvDate, &arg.State); err != nil {
			log.Println("scan error.    err =", err)
			return err
		}
	}

	return nil
}

// Pending fetches pending grns
// returns a slice of GrnLog and an error if it fails
func (arg *GrnLog) Pending(ctxt context.Context) ([]GrnLog, error) {
	sql := `
		SELECT 
			trans_date, grn_num, supp_name, inv_num, supp_pin, inv_type
			, tims_cuin, total_exc, total_amount_inc, total_vat, discount
			, vatable, exempt, vehicle_num, driver_name, registered_by, poster
			, inv_date, recv_date, state
		FROM grn_log 
		WHERE state IN ('pending', 'suspended') `

	ctx, cancel := context.WithTimeout(ctxt, 20*time.Second)
	defer cancel()

	rows, err := database.PgPool.Query(ctx, sql)
	if err != nil {
		log.Println("postgresql error,  failed to query pending grns    err =", err)
		return []GrnLog{}, err
	}
	defer rows.Close()

	vals := []GrnLog{}
	for rows.Next() {
		r := GrnLog{}

		err := rows.Scan(&r.TransDate, &r.GrnNum, &r.SuppName, &r.InvNum, &r.SuppPin, &r.InvType,
			&r.TimsCUIN, &r.TotalExc, &r.TotalAmountInc, &r.TotalVat, &r.Discount,
			&r.Vatable, &r.Exempt, &r.VehicleNum, &r.DriverName, &r.RegisteredBy, &r.Poster,
			&r.InvDate, &r.RecvDate, &r.State)
		if err != nil {
			log.Println("scan error.    err =", err)
			return vals, err
		}

		vals = append(vals, r)
	}

	return vals, nil
}

// GetItems fetches grn_items
// queries data from grn_items and populates Items field
// returns an error if it fails
func (arg *GrnLog) GetItems(ctxt context.Context) error {
	sql := `
		SELECT 
			auto_id, trans_date, grn_num, company_id, vat_type, 
			scan_code, item_code, item_name, 
			ctn_charged, doz_charged, pcs_charged, qty_charged, 
			item_cost, old_cost, item_price, vat, vatable, excempt, 
			vat_alpha, discount_type, discount, discount_val, qty_discount, 
			total_amount, total_amount_inc, net_qty, net_amount, state, location_id
		FROM grn_items 
		WHERE grn_num = $1`

	ctx, cancel := context.WithTimeout(ctxt, 15*time.Second)
	defer cancel()

	rows, err := database.PgPool.Query(ctx, sql, arg.GrnNum)
	if err != nil {
		return err
	}
	defer rows.Close()

	arg.Items = []GrnItem{}
	for rows.Next() {
		r := GrnItem{}
		err = rows.Scan(&r.AutoID, &r.TransDate, &r.GrnNum, &r.CompanyID, &r.InvType,
			&r.ScanCode, &r.ItemCode, &r.ItemName,
			&r.CtnCharged, &r.DozCharged, &r.PcsCharged, &r.QtyCharged,
			&r.ItemCost, &r.OldCost, &r.ItemPrice, &r.Vat, &r.Vatable, &r.Excempt,
			&r.VatAlpha, &r.DiscountType, &r.Discount, &r.DiscountVal, &r.QtyDiscount,
			&r.TotalAmount, &r.TotalAmountInc, &r.NetQty, &r.NetAmount, &r.State, &r.LocationID,
		)
		if err != nil {
			return err
		}

		arg.Items = append(arg.Items, r)
	}

	return nil
}

func (arg *GrnLog) priceChageExists(ctxt context.Context) (bool, error) {
	sql := `
		SELECT 
			count(*)
		FROM grn_items 
		WHERE grn_num = $1 
			AND LOWER(state) = 'price change' `

	ctx, cancel := context.WithTimeout(ctxt, 20*time.Second)
	defer cancel()

	rows, err := database.PgPool.Query(ctx, sql, arg.GrnNum)
	if err != nil {
		log.Println("postgresql error.  failed to cquery price_change rows     err =", err)
		return false, err
	}

	counted := 0
	for rows.Next() {
		err = rows.Scan(&counted)
		if err != nil {
			return false, err
		}
		if counted > 0 {
			return true, nil
		}
	}

	return false, nil
}

// checks if the grn being posted has matching amounts
func (arg *GrnLog) comparePostingGRN(ctxt context.Context) (bool, bool, bool) {
	// sql := `SELECT coun(*) FROM grn_items WHERE `
	sql := `
		SELECT
			CASE
				WHEN inv_type  in ('delivery', 'return', 'purchase_simple') THEN True
				WHEN ((inv_total - (SELECT discount FROM grn_log WHERE grn_num = $1)) >= (recv_total - 2) AND (inv_total - (SELECT discount FROM grn_log WHERE grn_num = $1)) <= (recv_total + 2) AND inv_type not in ('delivery', 'return', 'purchase_simple') ) THEN True 
				ELSE False
			END as total_match,
			CASE
				WHEN (total_vat >= (recv_vat - 1) AND total_vat <= (recv_vat + 1) AND inv_type not in ('delivery', 'return', 'purchase_simple') ) OR inv_type in ('delivery', 'return', 'purchase_simple')  
				THEN True ELSE False
			END as vat_match,
			CASE
				WHEN (SELECT COUNT(*) FROM grn_items WHERE state = 'PRICE CHANGE' AND grn_num = $1 ) > 0 AND inv_type not in ('delivery', 'return', 'purchase_simple') THEN True ELSE False
			END as price_change
		FROM (
			(SELECT grn_num, sum(total_amount_inc) inv_total, total_vat, inv_type FROM grn_log WHERE state != 'VOIDED' AND grn_num = $1  GROUP BY grn_num, total_vat, inv_type) as inv
				LEFT JOIN
			(SELECT grn_num, sum(total_amount_inc) recv_total, sum(vat) recv_vat FROM grn_items WHERE state NOT IN ('DELETED') AND grn_num = $1 AND inv_state != 'VOIDED' GROUP BY grn_num) as recv
			ON recv.grn_num = inv.grn_num
		)as a
	`

	ctx, cancel := context.WithTimeout(ctxt, 15*time.Second)
	defer cancel()

	rows, err := database.PgPool.Query(ctx, sql, arg.GrnNum)
	if err != nil {
		log.Printf("error getting comparison err = %v\n", err)
		return false, false, false
	}
	defer rows.Close()

	var totalMatch, vatMatch, priceChange bool
	for rows.Next() {
		rows.Scan(&totalMatch, &vatMatch, &priceChange)
	}

	// log.Fatalln("inv_type = ", arg.InvType)

	switch arg.InvType {
	case "purchase_simple", "delivery", "return":
		return true, true, priceChange
	}

	return totalMatch, vatMatch, priceChange
}

// completeItemsWithoutPriceChange
// updates item's state to
func (arg *GrnLog) completeItemsWithoutPriceChange(ctx context.Context, tx pgx.Tx) error {
	sql := `UPDATE grn_items
		 	SET
			 	state = 'COMPLETED'
		  	WHERE grn_num = $1 AND state = 'pending'`

	_, err := tx.Exec(ctx, sql, arg.GrnNum)
	if err != nil {
		log.Println("postgresql error. update grn_items     err =", err)
		return err
	}
	return nil
}

func (arg *GrnLog) completeGrnLog(ctx context.Context, tx pgx.Tx) error {
	// update grn logs
	sql := `UPDATE grn_log g
			SET 
				state = $3
				, poster = $2
				, trans_date = now()
			WHERE g.grn_num = $1
				AND state in ('pending', 'pending_pc', 'returning')`

	_, err := tx.Exec(ctx, sql, arg.GrnNum, arg.Poster, arg.State)
	if err != nil {
		log.Printf("failed to complete G.R.N. log     err = %v\n", err)
		return err
	}
	return nil
}

func (arg *GrnLog) completeWithCalcCost(ctx context.Context, tx pgx.Tx) error {
	// update grn logs
	sql := `UPDATE grn_log g
			SET
				total_amount_inc = a.total
				, total_vat = a.total_vat
				, vatable = a.vatable
			FROM 
				(SELECT 
					grn_num
					, SUM(total_amount_inc) as total 
					, SUM(vat) as total_vat
					, SUM(vatable) as vatable
				 FROM grn_items 
				 WHERE grn_num = $1 AND state = 'COMPLETED' 
				 GROUP BY grn_num
				) as a
			WHERE a.grn_num = g.grn_num 
				AND inv_type in ('return', 'delivery', 'purchase_simple')`

	_, err := tx.Exec(ctx, sql, arg.GrnNum)
	if err != nil {
		log.Printf("failed to complete G.R.N. log     err = %v\n", err)
		return err
	}
	return nil
}

func (arg *GrnLog) Complete(ctxt context.Context) error {
	// validate totals
	totalMatch, vatMatch, priceChange := arg.comparePostingGRN(ctxt)
	if !totalMatch {
		return fmt.Errorf(`failed to complete "Invoice totals do not match"`)
	} else if !vatMatch {
		return fmt.Errorf(`failed to complete "Invoice Vat do not match"`)
	}

	priceChange, err := arg.priceChageExists(ctxt)
	fmt.Printf("\t is_price_change = %v\n", priceChange)
	arg.State = "POSTED"
	if priceChange {
		arg.State = "pending_pc"
	}

	ctx, cancel := context.WithTimeout(ctxt, 50*time.Second)
	defer cancel()

	tx, err := database.PgPool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// update items to completed that are without price_change
	if err = arg.completeItemsWithoutPriceChange(ctx, tx); err != nil {
		log.Println("error. failed to complete_grn_items     err =", err)
		return err
	}

	if err := arg.completeGrnLog(ctx, tx); err != nil {
		return err
	}

	if err := arg.completeWithCalcCost(ctx, tx); err != nil {
		return err
	}

	if err := arg.GetItems(ctx); err != nil {
		return err
	}

	// arg.RecvDate has already been populated by Details() via a DB scan of the
	// TIMESTAMPTZ column, so it arrives as Postgres's timestamptz text format
	// (e.g. "2026-06-12 00:00:00+00"), not the "2006-01-02" form used in GetGrn.
	recvDate, err := time.Parse("2006-01-02 15:04:05-07", arg.RecvDate)
	if err != nil {
		return fmt.Errorf("invalid recv_date %q: %w", arg.RecvDate, err)
	}

	for _, itm := range arg.Items {
		if itm.State == "DELETED" {
			continue
		}

		// add to stock transactions
		bal := balances.TxnLog{
			TransDate:   recvDate,
			TxnID:       fmt.Sprintf("%v-%v", arg.GrnNum, itm.AutoID),
			Description: "purchases",
			LocationID:  itm.LocationID,
			ItemCode:    itm.ItemCode,
			QtyIn:       itm.NetQty,
		}

		if err := bal.LogBalTx(ctx, tx); err != nil {
			log.Println("error. failed to log transaction balance     err =", err)
			return err
		}

		fmt.Printf("\t added txn_log  item = %v \t location = %v\n", itm.ItemCode, itm.LocationID)
	}

	return tx.Commit(ctx)
}
