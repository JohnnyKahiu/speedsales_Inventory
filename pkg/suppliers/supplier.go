package suppliers

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/JohnnyKahiu/speedsales_inventory/database"
)

type Supplier struct {
	table      string `name:"supplier" type:"table"`
	AutoID     int64  `json:"auto_id" type:"field" sql:"BIGSERIAL"`
	SuppName   string `json:"supp_name" type:"field" sql:"VARCHAR NOT NULL"`
	LeadDays   int    `json:"lead_days" type:"field" sql:"INT NOT NULL DEFAULT 1"`
	CreditTerm string `json:"credit_term" type:"field" sql:"VARCHAR NOT NULL"`
	CreditDays int    `json:"credit_days" type:"field" sql:"INT NOT NULL"`
	KraPin     string `json:"kra_pin" type:"field" sql:"VARCHAR(14)"`
	Telephone  string `json:"telephone" type:"field" sql:"VARCHAR(60) NOT NULL DEFAULT ''"`
	Email      string `json:"email" type:"field" sql:"VARCHAR(60) NOT NULL DEFAULT ''"`
}

func GenSupplierTBL() error {
	return database.CreateFromStruct(Supplier{})
}

// New - creates a new supplier
// Inserts supplier details into supplier table
// returns an error if it fails
func (arg *Supplier) New(ctxt context.Context) error {
	fmt.Printf(`
		supp_name: %v
		lead_days: %v
		credit_term: %v
		credit_days: %v
		kra_pin: %v
		telephone: %v
		email: %v
	`, arg.SuppName, arg.LeadDays, arg.CreditTerm, arg.CreditDays, arg.KraPin, arg.Telephone, arg.Email)

	sql := `
		INSERT INTO supplier (supp_name, lead_days, credit_term, credit_days, kra_pin, telephone, email)
		VALUES($1, $2, $3, $4, $5, $6, $7) `

	ctx, cancel := context.WithTimeout(ctxt, 20*time.Second)
	defer cancel()

	_, err := database.PgPool.Exec(ctx, sql, arg.SuppName, arg.LeadDays, arg.CreditTerm, arg.CreditDays, arg.KraPin, arg.Telephone, arg.Email)
	if err != nil {
		log.Println("postgresql error. failed to add supplier     err =", err)
		return err
	}
	fmt.Println("new supplier added successfully")
	return nil
}

// FetchAll - fetches all listed suppliers
// Queries supplier from supplier table
// returns a slice of suppliers and error if it occurs
func (arg *Supplier) FetchAll(ctxt context.Context) ([]Supplier, error) {
	sql := `
		SELECT 
			auto_id
			, supp_name
			, lead_days
			, credit_term
			, credit_days
			, kra_pin
			, telephone
			, email
		FROM supplier
		LIMIT 100 `

	ctx, cancel := context.WithTimeout(ctxt, 20*time.Second)
	defer cancel()

	rows, err := database.PgPool.Query(ctx, sql)
	if err != nil {
		log.Println("postgresql error,  supplier search      err =", err)
		return []Supplier{}, nil
	}
	defer rows.Close()

	vals := []Supplier{}
	for rows.Next() {
		r := Supplier{}
		err = rows.Scan(&r.AutoID, &r.SuppName, &r.LeadDays, &r.CreditTerm, &r.CreditDays, &r.KraPin, &r.Telephone, &r.Email)
		if err != nil {
			log.Println("supplier scan error.    err =", err)
			return vals, err
		}

		vals = append(vals, r)
	}
	return vals, nil
}

func (arg *Supplier) Search(ctxt context.Context) ([]Supplier, error) {
	c := "%" + strings.Replace(arg.SuppName, "'", "|| chr(39) ||", -1) + "%"

	sql := fmt.Sprintf(`
		SELECT 
			auto_id
			, supp_name
			, lead_days
			, credit_term
			, credit_days
			, kra_pin
			, telephone
			, email
		FROM supplier WHERE supp_name ilike '%v'
	`, c)
	fmt.Println(c)

	ctx, cancel := context.WithTimeout(ctxt, 20*time.Second)
	defer cancel()

	rows, err := database.PgPool.Query(ctx, sql)
	if err != nil {
		log.Println("postgresql error,  supplier search      err =", err)
		return []Supplier{}, nil
	}
	defer rows.Close()

	vals := []Supplier{}
	for rows.Next() {
		r := Supplier{}
		err = rows.Scan(&r.AutoID, &r.SuppName, &r.LeadDays, &r.CreditTerm, &r.CreditDays, &r.KraPin, &r.Telephone, &r.Email)
		if err != nil {
			log.Println("supplier scan error.    err =", err)
			return vals, err
		}

		vals = append(vals, r)
	}
	return vals, nil
}
