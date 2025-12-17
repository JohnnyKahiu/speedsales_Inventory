package count

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/JohnnyKahiu/speedsales_inventory/database"
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

// AdoptItem = adopts stock balance from count
func (arg *CountItems) AdoptItem(ctxt context.Context) error {
	// Archive current balance
	// add new item
	// sql := `UPDATE count_items
	// 		SET
	// 			`

	return nil
}
