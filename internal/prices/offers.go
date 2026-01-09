package prices

import (
	"time"

	"github.com/JohnnyKahiu/speedsales_inventory/database"
)

type Offers struct {
	ItemCode      string    `json:"item_code" type:"field" sql:"VARCHAR NOT NULL"`
	OfferType     string    `json:"offer_type" type:"field" sql:"VARCHAR"`
	StartDate     time.Time `json:"" type:"field" sql:"TIMESTAMPTZ NOT NULL DEFAULT '1999-01-12'"`
	EndDate       time.Time `json:"" type:"field" sql:"TIMESTAMPTZ NOT NULL DEFAULT '1999-01-12'"`
	StartTime     time.Time `json:"" type:"field" sql:"TIMESTAMPTZ NOT NULL DEFAULT '1999-01-12'"`
	EndTime       time.Time `json:"" type:"field" sql:"TIMESTAMPTZ NOT NULL DEFAULT '1999-01-12'"`
	DiscQuantity  float64   `json:"" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	OfferPrice    float64   `json:"" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	DiscPerc      float64   `json:"" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	QtyDiscThresh float32   `json:"" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
}

// gen table
func GenOfferTbl() error {
	var tblStruct Offers
	return database.CreateFromStruct(tblStruct)
}
