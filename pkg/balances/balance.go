package balances

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/JohnnyKahiu/speedsales_inventory/database"
)

type Balance struct {
	ItemCode   string  `json:"item_code"`
	LocationID string  `json:"location_id"`
	QtyIn      float64 `json:"qty_in"`
	QtyOut     float64 `json:"qty_out"`
	Bal        float64 `json:"bal"`
}

type TxnLog struct {
	table       string    `name:"txn_log" type:"table"`
	TransDate   time.Time `json:"trans_date" type:"field" name:"trans_date" sql:"TIMESTAMP NOT NULL DEFAULT NOW()"`
	Description string    `json:"description" type:"field" sql:"VARCHAR NOT NULL DEFAULT ''"`
	TxnID       string    `json:"txn_id" type:"field" sql:"VARCHAR NOT NULL"`
	LocationID  int64     `json:"location_id" type:"field" sql:"BIGINT NOT NULL DEFAULT '0'"`
	ItemCode    string    `json:"item_code" type:"field" sql:"VARCHAR NOT NULL"`
	QtyIn       float64   `json:"qty_in" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	QtyOut      float64   `json:"qty_out" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	Qty         float64   `json:"quantity"`
	Bal         float64   `json:"bal"`
	Rcpt        string    `json:"receipt_item"`
	Constraint  string    `name:"pk_txn_log" type:"constraint" sql:"PRIMARY KEY(description, txn_id)"`
	ForeignLoc  string    `name:"fk_location_id" type:"constraint" sql:"FOREIGN KEY (location_id) REFERENCES stock_locations(auto_id)"`
}

func GenBalTable() error {
	var tblStruct TxnLog
	return database.CreateFromStruct(tblStruct)
}

// GetBal() fetches balances
// balance
// returns an error if it fails
func (arg *Balance) GetBal() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if arg.ItemCode == "" {
		return errors.New("error. item_code is null")
	}
	if arg.LocationID == "" {
		return errors.New("error. stock location is null")
	}

	sql := `SELECT
				SUM(qty_in - qty_out) balance
			FROM txn_log 
			WHERE location_id = $1 AND item_code = $2`

	row, err := database.PgPool.Query(ctx, sql, arg.LocationID, arg.ItemCode)
	if err != nil {
		return err
	}
	defer row.Close()

	for row.Next() {
		row.Scan(&arg.Bal)
	}

	return nil
}

func (arg *TxnLog) LogBal(ctx context.Context) error {
	sql := `INSERT INTO txn_log(description, txn_id, location_id, item_code, qty_in, qty_out)
			VALUES($1, $2, $3, $4, $5, $6) 
			ON CONFLICT ON CONSTRAINT pk_txn_log 
			DO UPDATE SET qty_in = excluded.qty_in, qty_out = excluded.qty_out`

	c, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	_, err := database.PgPool.Exec(c, sql, arg.Description, arg.TxnID, arg.LocationID, arg.ItemCode, arg.QtyIn, arg.QtyOut)
	if err != nil {
		log.Println("sql error. failed to insert into txn_log    err =", err)
		return err
	}

	log.Println("transaction logged successfully")

	return nil
}

func (arg *TxnLog) RemoveBal(ctx context.Context) error {
	sql := `UPDATE txn_log
			SET
				qty_in = 0
				, qty_out = 0
			WHERE description = $1 AND txn_id = $2`

	c, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	_, err := database.PgPool.Exec(c, sql, arg.Description, arg.TxnID, arg.LocationID, arg.ItemCode, arg.QtyIn, arg.QtyOut)
	if err != nil {
		log.Println("sql error. failed to insert into txn_log    err =", err)
		return err
	}

	log.Println("transaction logged successfully")

	return nil
}
