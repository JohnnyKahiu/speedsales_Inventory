package count

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/JohnnyKahiu/speedsales_inventory/database"
	"github.com/jackc/pgx/v5"
)

type CountLog struct {
	table     string    `name:"count_log" type:"table"`
	TransDate time.Time `json:"trans_date" type:"field" sql:"TIMESTAMPTZ NOT NULL DEFAULT now()"`
	CountID   int64     `json:"count_id" type:"field" sql:"BIGINT NOT NULL DEFAULT '0'"`
	Branch    string    `json:"branch" type:"field" sql:"VARCHAR NOT NULL"`
	CountType string    `json:"count_type" type:"field" sql:"VARCHAR NOT NULL"`
	Bins      []int64   `json:"bins" type:"field" sql:"INT[]"`
	Suppliers []string  `json:"suppliers" type:"field" sql:"VARCHAR[]"`
	CountEnd  time.Time `json:"count_end" type:"field" sql:"TIMESTAMPTZ"`
	Total     int       `json:"total" type:"field" sql:"INT NOT NULL DEFAULT '0'"`
	Variance  int       `json:"variance" type:"field" sql:"INT NOT NULL DEFAULT '0'"`
	Count     int64     `json:"count"`
}

func genCountLogTbl() error {
	return database.CreateFromStruct(CountLog{})
}

// GenNewID fetches a new id from count_log table
func (arg *CountLog) GenNewID(ctxt context.Context, tx pgx.Tx) error {
	sql := `
			SELECT 
				COUNT(*) 
			FROM count_log 
			WHERE trans_date::date = now()::date`

	err := tx.QueryRow(ctxt, sql).Scan(&arg.Count)
	if err != nil {
		log.Println("error failed to fetch count_log count     err =", err)
		return err
	}

	t := time.Now()
	countID := fmt.Sprintf("%d%02d%02d0%d", t.Year(), t.Month(), t.Day(), arg.Count)
	arg.CountID, _ = strconv.ParseInt(countID, 10, 64)

	return nil
}

// New creates a new count_log
// generates new count_id and inserts into count_log table
// returns an error if it fails
func (arg *CountLog) New(ctxt context.Context) error {
	ctx, cancel := context.WithTimeout(ctxt, 15*time.Second)
	defer cancel()

	tx, err := database.PgPool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	err = arg.GenNewID(ctx, tx)
	if err != nil {
		return err
	}

	sql := `INSERT INTO count_log(branch, count_id, count_type, bins, suppliers)
			VALUES($1, $2, $3, $4, $5)`
	_, err = tx.Exec(ctx, sql, arg.Branch, arg.CountID, arg.CountType, arg.Bins, arg.Suppliers)
	if err != nil {
		log.Println("error. failed to create new count_log     err =", err)
		return err
	}
	return tx.Commit(ctx)
}

// Active  fetches all active count_logs
// Queries CountLog from the database
// returns a list of count_log and an error if found
func (arg *CountLog) Active(ctxt context.Context) ([]CountLog, error) {
	sql := `
		SELECT 
			branch, trans_date, count_id, count_type, bins, suppliers, total, variance
		FROM count_log 
		WHERE count_end IS NULL`

	ctx, cancel := context.WithTimeout(ctxt, 15*time.Second)
	defer cancel()

	rows, err := database.PgPool.Query(ctx, sql)
	if err != nil {
		log.Println("pg error.   failed to query active_counts    err =", err)
		return []CountLog{}, err
	}
	defer rows.Close()

	vals := []CountLog{}
	for rows.Next() {
		r := CountLog{}
		err = rows.Scan(&r.Branch, &r.TransDate, &r.CountID, &r.CountType, &r.Bins, &r.Suppliers, &r.Total, &r.Variance)
		if err != nil {
			return vals, err
		}

		vals = append(vals, r)
	}

	return vals, nil
}

// FetchBins fetches all bin details
func (arg *CountLog) FetchBins(ctxt context.Context) ([]CountLog, error) {
	return nil, nil
}
