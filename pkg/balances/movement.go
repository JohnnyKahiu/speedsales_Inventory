package balances

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
)

type Movement struct {
	table       string  `name:"stk_mvmt" type:"table"`
	AutoID      int64   `json:"auto_id" name:"auto_id" type:"field" sql:"BIGSERIAL UNIQUE"`
	TransDate   string  `json:"trans_date" name:"trans_date" type:"field" sql:"TIMESTAMPTZ NOT NULL DEFAULT NOW()"`
	RecID       string  `json:"rec_id" name:"rec_id" type:"field" sql:"VARCHAR NOT NULL"`
	ItemCode    string  `json:"item_code" name:"item_code" type:"field" sql:"VARCHAR NOT NULL"`
	TxnTrace    string  `json:"txn_trace" name:"txn_trace" type:"field" sql:"VARCHAR NOT NULL"`
	CompanyID   int64   `json:"company_id" name:"company_id" type:"field" sql:"BIGINT NOT NULL DEFAULT '0'"`
	Branch      string  `json:"branch" name:"branch" type:"field" sql:"VARCHAR NOT NULL"`
	StkLocation string  `json:"stk_location" name:"stk_location" type:"field" sql:"VARCHAR NOT NULL"`
	Description string  `json:"description" name:"description" type:"field" sql:"VARCHAR NOT NULL"`
	Qty         float64 `json:"qty" name:"qty" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	Bal         float64 `json:"bal" name:"bal" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	SerialNum   string  `json:"serial_num" name:"serial_num" type:"field" sql:"VARCHAR"`
	constraint  string  `name:"" type:"field" sql:"PRIMARY KEY (item_code, rec_id, description, stk_location, branch)"`
}

// RecordMvmtTX updates a stock transaction in context
// Inserts stock movement transaction to database
// Returns an error if it fails
func (arg *Movement) RecordMvmtTX(ctx context.Context, tx pgx.Tx) error {
	if arg.Qty == 0 {
		return fmt.Errorf("zero quantity to record")
	}

	sql := `
		INSERT INTO stk_mvmt_live(rec_id, item_code, txn_trace, company_id, branch, stk_location, description, qty)
		VALUES($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING auto_id`

	var autoID int64
	err := tx.QueryRow(ctx, sql, arg.RecID, arg.ItemCode, arg.TxnTrace, arg.CompanyID, arg.Branch, arg.StkLocation, arg.Description, arg.Qty).Scan(&autoID)
	if err != nil {
		fmt.Printf("\t\tTrace = %v \n\t\tRec_ID = %v\n", arg.TxnTrace, arg.RecID)
		log.Println("error inserting stk_mvmt    err =", err)
		return err

	}

	return nil
}
