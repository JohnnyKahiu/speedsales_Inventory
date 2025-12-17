package balances

import (
	"time"

	"github.com/JohnnyKahiu/speedsales_inventory/database"
)

type TxnArchive struct {
	table      string    `name:"txn_archives" type:"table"`
	AutoID     int64     `json:"auto_id" type:"field" sql:"BIGSERIAL NOT NULL UNIQUE"`
	ArchDate   time.Time `json:"arch_date" type:"field" sql:"TIMESTAMPTZ NOT NULL DEFAULT NOW()"`
	ItemCode   string    `json:"item_code" type:"field" sql:"VARCHAT NOT NULL"`
	LocationID int64     `json:"location_id" type:"field" sql:"BIGINT NOT NULL"`
	TxnTrail   TxnLog    `json:"txn_trail" type:"field" sql:"JSONB NOT NULL"`
}

func GenArchiveTbl() error {
	t := TxnArchive{}
	return database.CreateFromStruct(t)
}

func (arg *TxnArchive) New() error {

	return nil
}
