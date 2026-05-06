package count

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/JohnnyKahiu/speedsales_inventory/database"
	"github.com/JohnnyKahiu/speedsales_inventory/pkg/balances"
	"github.com/jackc/pgx/v5"
)

type CountItems struct {
	table        string       `name:"count_items" type:"table"`
	AutoID       int64        `json:"auto_id" type:"field" sql:"BIGSERIAL NOT NULL"`
	TransDate    time.Time    `json:"trans_date" type:"field" sql:"TIMESTAMPTZ NOT NULL DEFAULT now()"`
	CountNum     int64        `json:"count_num" type:"field" sql:"BIGINT NOT NULL"`
	LocationID   int64        `json:"location_id" type:"field" sql:"INT NOT NULL"`
	ItemCode     string       `json:"item_code" type:"field" sql:"VARCHAR(100) NOT NULL"`
	ItemsPerCase float64      `json:"items_per_case" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	Cases        float64      `json:"cases" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	GramsLtrs    float64      `json:"grams_litres" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	Pieces       float64      `json:"pieces" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	Counted      float64      `json:"counted" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	SystemBal    float64      `json:"system_bal" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	pkey         string       `name:"count_items_pk" type:"constraint" sql:"PRIMARY KEY (auto_id)"`
	DeptName     string       `json:"dept_name"`
	ItemName     string       `json:"item_name"`
	AutoIDs      []int64      `json:"auto_ids"`
	LogIDs       []int64      `json:"log_ids"`
	CountTrail   []countTrail `json:"count_trail"`
}

type countTrail struct {
	TransDate    time.Time `json:"trans_date"`
	Poster       string    `json:"poster"`
	LocationID   int64     `json:"location_id"`
	ItemCode     string    `json:"item_code"`
	Cases        float64   `json:"cases"`
	UnitsPerPack float64   `json:"units_per_pack"`
	Pieces       float64   `json:"pieces"`
}

func genCountItemsTbl() error {
	return database.CreateFromStruct(CountItems{})
}

// GenCountTbls creates all count tables
func GenCountTbls() error {
	err := genCountLogTbl()
	if err != nil {
		return err
	}

	return genCountItemsTbl()
}

// Count - records items counted
// Receives a context and updates count_items with count details and system_bal
// Returns an error if it occurs
func (arg *CountItems) Count(ctxt context.Context) error {
	fmt.Printf("\t pieces = %v \n\t counted = %v \n\t balance = %v \n\t auto_id = %v \n counted = %v \n\n", arg.Pieces, arg.Counted, arg.SystemBal, arg.AutoID, arg.Counted)
	sql := `
		UPDATE count_items
		SET 
			pieces = $1
			, cases = $2
			, items_per_case = $3
			, counted = $4
			, system_bal = $5
			, trans_date = now()
		WHERE auto_id = $6`

	ctx, cancel := context.WithTimeout(ctxt, 15*time.Second)
	defer cancel()

	_, err := database.PgPool.Exec(ctx, sql, arg.Pieces, arg.Cases, arg.ItemsPerCase, arg.Counted, arg.SystemBal, arg.AutoID)
	if err != nil {
		log.Println("error, failed to insert into count_items.    err =", err)
		return err
	}

	return nil
}

func (arg *CountItems) ArchiveTxn(ctx context.Context, tx pgx.Tx) error {
	sql := `
		INSERT INTO txn_archives(location_id, item_code, txn_trail, txn_trace)
		SELECT 
			ci.location_id
			, ci.item_code
			, jsonb_agg(to_jsonb(l)) AS txn_trail
			, CONCAT(ci.count_num, '-', ci.auto_id) as txn_trace
		FROM txn_log l 
		INNER JOIN count_items ci 
			ON ci.item_code = l.item_code 
			AND l.location_id = ci.location_id
		WHERE ci.auto_id = ANY($1) 
			AND l.trans_date <= ci.trans_date 
		GROUP BY ci.location_id, ci.item_code, ci.count_num, ci.auto_id;  `

	_, err := tx.Exec(ctx, sql, arg.AutoIDs)
	if err != nil {
		log.Println("postgresql error.  failed to insert itno archives     err =", err)
		return err
	}

	return nil
}

func (arg *CountItems) WriteNewBal(ctx context.Context, tx pgx.Tx) error {
	fmt.Println("writing balance")
	sql := `
		INSERT INTO txn_log(trans_date, description, txn_id, location_id, item_code, qty_in)
		SELECT 
			trans_date
			, 'adopted stock count' as description
			, coalesce(ci.count_num) as txn_id
			, ci.location_id
			, ci.item_code
			, ci.counted as qty_in
		FROM count_items ci 
		WHERE auto_id = ANY($1)
		RETURNING item_code, qty_in, location_id
	`

	locID := int64(0)
	Qty := float64(0)
	itemCode := ""

	err := tx.QueryRow(ctx, sql, arg.AutoIDs).Scan(&itemCode, &Qty, &locID)
	if err != nil {
		log.Println("postgresql error     failed to add new balance   err =", err)
		return err
	}

	return nil
}

func (arg *CountItems) GetById(ctx context.Context, tx pgx.Tx) error {
	sql := `
		SELECT 
			trans_date
			, count_num
			, location_id
			, item_code
			, counted
			, system_bal
		FROM count_items WHERE auto_id = $1
	`

	rows, err := database.PgPool.Query(ctx, sql, arg.AutoID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&arg.TransDate, &arg.CountNum, &arg.LocationID, &arg.ItemCode, &arg.Counted, &arg.SystemBal); err != nil {
			return err
		}
	}
	return nil
}

// AdoptItem = adopts stock balance from count
// archives
func (arg *CountItems) AdoptItem(ctxt context.Context) error {
	ctx, cancel := context.WithTimeout(ctxt, 30*time.Second)
	defer cancel()

	tx, err := database.PgPool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	fmt.Println("\t coun_num =", arg.CountNum)
	fmt.Println("\t auto_ids =", arg.AutoIDs)

	for _, id := range arg.AutoIDs {
		arg.AutoID = id

		err = arg.GetById(ctx, tx)
		if err != nil {
			return err
		}

		b := balances.Balance{
			LocationID: fmt.Sprintf("%v", arg.LocationID),
			ItemCode:   arg.ItemCode,
		}
		b.GetBal()

		txn := balances.TxnLog{
			TransDate:   arg.TransDate,
			TxnID:       fmt.Sprintf("%v", arg.CountNum),
			ItemCode:    arg.ItemCode,
			Description: "adopted stock count",
			LocationID:  arg.LocationID,
			QtyIn:       arg.Counted,
			QtyOut:      0,
		}
		err = txn.LogBalTx(ctx, tx)
		if err != nil {
			return err
		}

		err = txn.SaveBal(ctx)
		if err != nil {
			return err
		}

		err = txn.ArchivePrev(ctx, tx)
		if err != nil {
			return err
		}

		err = txn.PurgePrev(ctx, tx)
		if err != nil {
			return err
		}
	}

	err = tx.Commit(ctx)
	return err
}
