package products

import (
	"context"
	"fmt"
	"time"

	"github.com/JohnnyKahiu/speedsales_inventory/database"
)

type Vats struct {
	table string  `name:"vats" type:"table"`
	Code  string  `json:"code" type:"field" sql:"VARCHAR NOT NULL"`
	Name  string  `json:"name" type:"field" sql:"VARCHAR NOT NULL"`
	Rate  float64 `json:"rate" type:"field" sql:"FLOAT NOT NULL DEFAULT '0.0'"`
	pkey  string  `name:"vats_pkey" type:"constraint" sql:"PRIMARY KEY (code)"`
}

// create vats table
func CreateVatsTable() error {
	var v Vats
	return database.CreateFromStruct(v)
}

// CreateVat creates a new vat
func (v *Vats) CreateVat(ctx context.Context) error {
	sql := `INSERT INTO vats(code, name, rate) 
			VALUES($1, $2, $3)
			ON CONFLICT DO NOTHING`

	_, err := database.PgPool.Query(ctx, sql, v.Code, v.Name, v.Rate)
	if err != nil {
		return err
	}

	return nil
}

// UpdateVat updates a vat
// Updates Vats table with new rates and code
// returns an error if fails
func (v *Vats) UpdateVat(ctx context.Context) error {
	sql := `UPDATE vats SET rate = $1 WHERE code = $2`

	_, err := database.PgPool.Query(ctx, sql, v.Rate, v.Code)
	if err != nil {
		return err
	}
	return nil
}

// CreateVatsDefaults creates default vats
// Creates a list of default vats for a defined country
// returns an error if fails
func CreateVatsDefaults() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	vatTbl := []Vats{
		{Code: "A", Name: "VAT", Rate: 16},
		{Code: "B", Name: "VAT", Rate: 0},
		{Code: "C", Name: "VAT", Rate: 0},
		{Code: "D", Name: "VAT", Rate: 0},
		{Code: "E", Name: "VAT", Rate: 0},
	}

	fmt.Println("Creating defaults  ")
	// create default vats
	for _, v := range vatTbl {
		fmt.Printf("\t code = %s \n", v.Code)
		err := v.CreateVat(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}
