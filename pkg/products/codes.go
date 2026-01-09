package products

import (
	"context"
	"errors"
	"log"

	"github.com/JohnnyKahiu/speedsales_inventory/database"
)

// CodeTranslator structure holds translation of master and linked codes
type CodeTranslator struct {
	table       string  `name:"code_translator" type:"table"`
	AutoID      int64   `json:"auto_id" name:"auto_id" type:"field" sql:"BIGSERIAL UNIQUE NOT NULL"`
	MasterCode  string  `json:"master_code" name:"master_code" type:"field" sql:"VARCHAR(30) NOT NULL"`
	LinkCode    string  `json:"link_code" name:"link_code" type:"field" sql:"VARCHAR(30) NOT NULL"`
	PkgQty      float64 `json:"pkg_qty" name:"pkg_qty" type:"field" sql:"INT NOT NULL DEFAULT '1'"`
	Discount    float64 `json:"discount" name:"discount" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	constraint  int     `name:"code_translator_pkey" type:"constraint" sql:"PRIMARY KEY(link_code, master_code)"`
	constraint2 string  `name:"code_translator_link_code_key" type:"constraint" sql:"UNIQUE(link_code)"`
}

// createCodeTbls - generates a new code_translator table from
// returns an error if it fails
func createCodeTbls() error {
	var tblStruct CodeTranslator
	return database.CreateFromStruct(tblStruct)
}

// GetAllLinks fetches a list of all codes linked to a master
// searches all codes and returns all where master code in arg method is same as in productsMaster
// returns a slice of CodeTranslator and an error if exists
func (arg CodeTranslator) GetAllLinks() ([]CodeTranslator, error) {
	results := []CodeTranslator{}
	codes := ProdMaster.Codes
	for _, code := range codes {
		if code.MasterCode == arg.MasterCode {
			results = append(results, code)
		}
	}
	return results, nil
}

// New() Creates a new link to Codes database
// Adds the code_translation to database
// Adds the code_translation to file cache
// returns an error if it fails
func (arg *CodeTranslator) New(ctx context.Context) error {
	if arg.LinkCode == "" {
		log.Println("error CodeTranslator.New()    null link_code")
		return errors.New("null params link_code is null")
	}
	if arg.MasterCode == "" {
		log.Println("error CodeTranslator.New()    null master_code")
		return errors.New("null params master_code is null")
	}

	// Cache links
	err := ProdMaster.UpdateLinks(*arg)
	if err != nil {
		log.Println("error CodeTranslator.New()    ProdMaster linking fail   err =", err)
		return err
	}

	err = arg.AddDB(ctx)
	if err != nil {
		log.Println("error CodeTranslator.New()    DB txn fail    err =", err)
		return err
	}

	return nil
}

// CodeTranslator.AddDB adds a new code translation to products relational database
// Inserts record to database
// returns an error if any occurs
func (arg *CodeTranslator) AddDB(ctx context.Context) error {
	sql := `INSERT INTO code_translator(link_code, master_code, pkg_qty, discount) VALUES($1, $2, $3, $4) 
			ON CONFLICT ON CONSTRAINT code_translator_link_code_key DO 
			UPDATE SET master_code = $2, pkg_qty = $3, discount = $4 `

	_, err := database.PgPool.Exec(ctx, sql, arg.LinkCode, arg.MasterCode, arg.PkgQty, arg.Discount)
	if err != nil {
		log.Println("failed adding link_code to code_translator    err =", err)
		return err
	}
	return nil
}
