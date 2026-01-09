package balances

import (
	"context"
	"errors"
	"time"

	"github.com/JohnnyKahiu/speedsales_inventory/database"
)

type Balance struct {
	ItemCode   string  `json:"item_code"`
	LocationID int64   `json:"location_id"`
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
}

func genBalTable() error {
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
	if arg.LocationID == 0 {
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
	Bal         float64   `json:"bal"`
	defer row.Close()

	for row.Next() {
		row.Scan(&arg.Bal)
	}

	return nil
}
