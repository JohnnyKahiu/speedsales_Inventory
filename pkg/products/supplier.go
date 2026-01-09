package products

import (
	"context"
	"log"

	"github.com/JohnnyKahiu/speedsales_inventory/database"
)

type Supplier struct {
	table     string  `type:"table" name:"suppliers"`
	AutoID    int64   `json:"auto_id" name:"auto_id" type:"field" sql:"BIGSERIAL UNIQUE NOT NULL"`
	Name      string  `json:"name" name:"name" type:"field" sql:"VARCHAR(100) NOT NULL"`
	Address   string  `json:"address" name:"address" type:"field" sql:"VARCHAR(255) NOT NULL"`
	Telephone string  `json:"telephone" name:"telephone" type:"field" sql:"VARCHAR(20) NOT NULL"`
	Email     string  `json:"email" name:"email" type:"field" sql:"VARCHAR(100) NOT NULL"`
	LeadDays  float32 `json:"lead_days" name:"lead_days" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	TaxPin    string  `json:"tax_pin" name:"tax_pin" type:"field" sql:"VARCHAR(20) NOT NULL"`
}

// createSupplierTbl creates a new supplier table in the database
// returns an error if fails
func createSupplierTbl() error {
	var tblStruct Supplier
	return database.CreateFromStruct(tblStruct)
}

// Create creates a new supplier record in the database
// inserts the record into the database
// returns an error if fails
func (arg *Supplier) Create() error {
	sql := `INSERT INTO suppliers (name, address, telephone, email, lead_days, tax_pin)
			VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := database.PgPool.Exec(context.Background(), sql, arg.Name, arg.Address, arg.Telephone, arg.Email, arg.LeadDays, arg.TaxPin)
	if err != nil {
		log.Println("error. failed to create supplier    err =", err)
		return err
	}
	return nil
}
