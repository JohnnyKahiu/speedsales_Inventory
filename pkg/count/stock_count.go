package count

import (
	"time"

	"github.com/JohnnyKahiu/speedsales_inventory/database"
)

type CountItems struct {
	table        string    `name:"count_items" type:"table"`
	TransDate    time.Time `json:"trans_date" type:"field" sql:"TIMESTAMPTZ NOT NULL DEFAULT now()"`
	CountNum     int64     `json:"count_num" type:"field" sql:"BIGINT NOT NULL"`
	LocationID   int64     `json:"location_id" type:"field" sql:"INT NOT NULL"`
	ItemCode     string    `json:"item_code" type:"field" sql:"VARCHAR(100) NOT NULL"`
	ItemsPerCase float64   `json:"items_per_case" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	Cases        float64   `json:"cases" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	GramsLtrs    float64   `json:"grams_litres" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	Pieces       float64   `json:"pieces" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	Counted      float64   `json:"counted" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	SystemBal    float64   `json:"system_bal" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
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

func (arg *CountItems) Count() error {
	return nil
}
