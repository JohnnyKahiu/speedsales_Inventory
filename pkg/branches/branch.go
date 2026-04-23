package branches

import (
	"context"
	"log"
	"time"

	"github.com/JohnnyKahiu/speedsales_inventory/database"
)

type Branch struct {
	table      string `name:"branches" type:"table"`
	AutoID     int64  `json:"auto_id" type:"field" sql:"BIGSERIAL NOT NULL"`
	BranchName string `json:"branch_name" type:"field" sql:"VARCHAR(50) NOT NULL"`
	Telephone  string `json:"telephone" type:"field" sql:"VARCHAR(50) NOT NULL DEFAULT ''"`
	Email      string `json:"email" type:"field" sql:"VARCHAR NOT NULL DEFAULT ''"`
	Coordinate string `json:"coordinates" type:"field" sql:"VARCHAR NOT NULL DEFAULT '{0.213, -3.234}'"`
	constraint string `name:"pk_branch" type:"constraint" sql:"PRIMARY KEY(auto_id)"`
}

func GenBranchTbl() error {
	return database.CreateFromStruct(Branch{})
}

// New creates a new branch
// Inserts a new branch record
// populates corresponding branch auto_id and
// returns an error if it fails
func (arg *Branch) New() error {
	sql := `INSERT INTO branches(branch_name, telephone, email, coordinates)
			VALUES ($1, $2, $3, $4)
			RETURNING auto_id`

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := database.PgPool.QueryRow(ctx, sql, arg.BranchName, arg.Telephone, arg.Email, arg.Coordinate).Scan(&arg.AutoID); err != nil {
		log.Println("database error. failed to add branch    err =", err)
		return err
	}

	return nil
}

// Fetch queries all Branches from database
// reuturns a slice of branches and an error if it occurs
func (arg *Branch) FetchAll() ([]Branch, error) {
	sql := `SELECT 
				auto_id, branch_name, telephone, email, coordinates
			FROM branches`

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	rows, err := database.PgPool.Query(ctx, sql)
	if err != nil {
		return []Branch{}, err
	}
	defer rows.Close()

	vals := []Branch{}
	for rows.Next() {
		r := Branch{}
		if err := rows.Scan(&r.AutoID, &r.BranchName, &r.Telephone, &r.Email, &r.Coordinate); err != nil {
			return vals, err
		}

		vals = append(vals, r)
	}

	return vals, nil
}
