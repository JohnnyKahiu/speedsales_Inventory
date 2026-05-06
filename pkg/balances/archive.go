package balances

import (
	"time"

	"github.com/JohnnyKahiu/speedsales_inventory/database"
)

type TxnArchive struct {
	table       string    `name:"txn_archives" type:"table"`
	AutoID      int64     `json:"auto_id" type:"field" sql:"BIGSERIAL NOT NULL UNIQUE"`
	TransDate   time.Time `json:"trans_date" type:"field" sql:"TIMESTAMPTZ NOT NULL DEFAULT NOW()"`
	Description string    `json:"description" type:"field" sql:"VARCHAR NOT NULL DEFAULT ''"`
	TxnID       string    `json:"txn_id" type:"field" sql:"VARCHAR NOT NULL"`
	LocationID  int64     `json:"location_id" type:"field" sql:"BIGINT NOT NULL DEFAULT '0'"`
	ItemCode    string    `json:"item_code" type:"field" sql:"VARCHAR NOT NULL"`
	QtyIn       float64   `json:"qty_in" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	QtyOut      float64   `json:"qty_out" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
}

func GenArchiveTbl() error {
	t := TxnArchive{}
	return database.CreateFromStruct(t)
}

func (arg *TxnArchive) New() error {

	return nil
}
