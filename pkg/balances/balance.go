package balances

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/JohnnyKahiu/speedsales_inventory/database"
	"github.com/JohnnyKahiu/speedsales_inventory/pkg/products"
	"github.com/jackc/pgx/v5"
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

// LogBal
// returns an error if it fails
func (arg *TxnLog) LogBal(ctx context.Context) error {
	c, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	tx, err := database.PgPool.BeginTx(c, pgx.TxOptions{})
	if err != nil {
		log.Println("logBal error   err =", err)
		return err
	}
	defer tx.Rollback(c)

	sql := `INSERT INTO txn_log(description, txn_id, location_id, item_code, qty_in, qty_out)
			VALUES($1, $2, $3, $4, $5, $6) 
			ON CONFLICT ON CONSTRAINT pk_txn_log 
			DO UPDATE 
			SET 
				qty_in = excluded.qty_in
				, qty_out = excluded.qty_out`

	_, err = tx.Exec(c, sql, arg.Description, arg.TxnID, arg.LocationID, arg.ItemCode, arg.QtyIn, arg.QtyOut)
	if err != nil {
		log.Println("sql error. failed to insert into txn_log    err =", err)
		return err
	}

	sql = ` SELECT 
				SUM(qty_in - qty_out) as bal 
			FROM txn_log 
			WHERE location_id = $1 AND item_code = $2`
	if err = tx.QueryRow(c, sql, arg.LocationID, arg.ItemCode).Scan(&arg.Bal); err != nil {
		log.Println("sql error, failed to fetch balance     err =", err)
		return err
	}

	log.Println("transaction logged successfully")
	return tx.Commit(c)
}

// RemoveBal
// returns an error if it fails
func (arg *TxnLog) RemoveBal(ctx context.Context) error {
	c, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	tx, err := database.PgPool.BeginTx(c, pgx.TxOptions{})
	if err != nil {
		log.Println("logBal error   err =", err)
		return err
	}
	defer tx.Rollback(c)

	sql := `UPDATE txn_log
			SET
				qty_in = 0
				, qty_out = 0
			WHERE description = $1 AND txn_id = $2`

	_, err = tx.Exec(c, sql, arg.Description, arg.TxnID, arg.LocationID, arg.ItemCode, arg.QtyIn, arg.QtyOut)
	if err != nil {
		log.Println("sql error. failed to insert into txn_log    err =", err)
		return err
	}

	sql = ` SELECT 
				SUM(qty_in - qty_out) as bal 
			FROM txn_log 
			WHERE location_id = $1 AND item_code = $2`
	if err = tx.QueryRow(c, sql, arg.LocationID, arg.ItemCode).Scan(&arg.Bal); err != nil {
		log.Println("sql error, failed to fetch balance     err =", err)
		return err
	}

	log.Println("transaction logged successfully")
	return tx.Commit(c)
}

// SaveBal saves the balance to cache
// returns an error if it fails
func (arg *TxnLog) SaveBal(ctx context.Context) error {
	// cache the  balance
	product := products.ProdMaster.ProductDB[arg.ItemCode]
	LocBal := product.Balance[arg.LocationID]
	LocBal = arg.Bal

	if product.Balance == nil {
		product.Balance = make(map[int64]float64)
	}

	product.Balance[arg.LocationID] = LocBal
	product.Bal = arg.Bal

	if products.ProdMaster.ProductDB == nil {
		products.ProdMaster.ProductDB = make(map[string]products.StockMaster)
	}

	products.ProdMaster.ProductDB[arg.ItemCode] = product
	return products.ProdMaster.Pickle()
}
