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
	table     string       `name:"count_log" type:"table"`
	TransDate time.Time    `json:"trans_date" type:"field" sql:"TIMESTAMPTZ NOT NULL DEFAULT now()"`
	CountID   int64        `json:"count_id" type:"field" sql:"BIGINT NOT NULL DEFAULT '0'"`
	Branch    string       `json:"branch" type:"field" sql:"VARCHAR NOT NULL"`
	CountType string       `json:"count_type" type:"field" sql:"VARCHAR NOT NULL"`
	Bins      []int64      `json:"bins" type:"field" sql:"INT[]"`
	Suppliers []string     `json:"suppliers" type:"field" sql:"VARCHAR[]"`
	CountEnd  time.Time    `json:"count_end" type:"field" sql:"TIMESTAMPTZ"`
	Total     int          `json:"total" type:"field" sql:"INT NOT NULL DEFAULT '0'"`
	Variance  int          `json:"variance" type:"field" sql:"INT NOT NULL DEFAULT '0'"`
	Items     []CountItems `json:"count_items"`
	Count     int64        `json:"count"`
	Bin       int64        `json"bin"`
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

	// generate new count_id
	err = arg.GenNewID(ctx, tx)
	if err != nil {
		return err
	}

	// Record into database
	err = arg.RecordCount(ctx, tx)
	if err != nil {
		return err
	}

	// add items to count_items
	err = arg.AddItems(ctx, tx)
	if err != nil {
		log.Println("error. failed to add items to count_log     err =", err)
		return err
	}
	return tx.Commit(ctx)
}

// RecordCount - inserts data into count_log table
// returns an error if it fails
func (arg *CountLog) RecordCount(ctx context.Context, tx pgx.Tx) error {
	sql := `INSERT INTO count_log(branch, count_id, count_type, bins, suppliers)
			VALUES($1, $2, $3, $4, $5)`

	_, err := tx.Exec(ctx, sql, arg.Branch, arg.CountID, arg.CountType, arg.Bins, arg.Suppliers)
	if err != nil {
		log.Println("error. failed to create new count_log     err =", err)
		return err
	}

	return nil
}

// AddItems adds count_log items into count Items
// returns an error if it fails
func (arg *CountLog) AddItems(ctx context.Context, tx pgx.Tx) error {
	sql := `INSERT INTO count_items(count_num, location_id, item_code, items_per_case)
			SELECT 
				cl.count_id, l.auto_id, m.item_code, m.units_per_pack::float
			FROM stock_locations l 
				INNER JOIN count_log cl ON l.auto_id = ANY (cl.bins)
				LEFT JOIN stock_master m ON m.item_code = ANY (l.stock_list)
			WHERE cl.count_id = $1
			`

	_, err := tx.Exec(ctx, sql, arg.CountID)
	if err != nil {
		log.Println("sql error. failed to add count_items    err =", err)
		return err
	}

	return nil
}

// Active  fetches all active count_logs
// Queries CountLog from the database
// returns a list of count_log and an error if found
func (arg *CountLog) Active(ctxt context.Context) ([]CountLog, error) {
	sql := `
		SELECT 
			branch, trans_date, count_id, count_type, 
			bins, suppliers, total, variance
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

// FetchItems - queries count items
// database queries of count_items table by count_id
// populates the data in count_log
// returns an error if it fails
func (arg *CountLog) FetchItems(ctxt context.Context) error {
	binCon := `$2 = $2`
	if arg.Bin != 0 {
		binCon = `ci.location_id = $2`
	}

	sql := `
		SELECT 
			ci.auto_id, cl.count_id, ci.location_id, 
			m.item_code, m.item_name, m.units_per_pack::float, m.dept_name, 
			ci.counted, ci.system_bal
		FROM count_items ci
			INNER JOIN count_log cl ON ci.count_num = cl.count_id
			LEFT JOIN stock_master m ON m.item_code = ci.item_code
		WHERE cl.count_id = $1 AND ` + binCon + `
		ORDER BY m.item_code ASC`
	fmt.Printf("%s \n vals(%v, %v)", sql, arg.CountID, arg.Bin)

	ctx, cancel := context.WithTimeout(ctxt, 15*time.Second)
	defer cancel()

	rows, err := database.PgPool.Query(ctx, sql, arg.CountID, arg.Bin)
	if err != nil {
		return err
	}
	defer rows.Close()

	arg.Items = []CountItems{}
	for rows.Next() {
		r := CountItems{}
		rows.Scan(&r.AutoID, &r.CountNum, &r.LocationID,
			&r.ItemCode, &r.ItemName, &r.ItemsPerCase, &r.DeptName,
			&r.Counted, &r.SystemBal)

		arg.Items = append(arg.Items, r)
	}

	return nil
}

// UpdateEnd - updates end_time and variances
// Updates count_log with now as count_end and
// count of total items with varying balance and count
// returns an error if it fails
func (arg *CountLog) UpdateEnd(ctxt context.Context) error {
	sql := `UPDATE count_log 
			SET 
				count_end = now() 
				, variance = (SELECT count(*) FROM count_items WHERE counted != system_bal AND count_num = $1)
				, total = (SELECT count(*) FROM count_items WHERE count_num = $1)
			WHERE count_id = $1`

	ctx, cancel := context.WithTimeout(ctxt, 15*time.Second)
	defer cancel()

	_, err := database.PgPool.Exec(ctx, sql, arg.CountID)
	if err != nil {
		log.Println("postgresql error  failed to update count_log       err =", err)
		return err
	}

	return nil
}

// Complete  sets count_log as complete
// Updates count_end in count_log
// returns an error if fails
func (arg *CountLog) Complete(ctxt context.Context) error {
	// update variance
	err := arg.UpdateEnd(ctxt)
	if err != nil {
		log.Println("error failed to complete count_log     err =", err)
		return err
	}

	// add items to variance list
	return nil
}

// FetchCompeted := Queries all completed stock counts between given time
// receives a context,  start = upper time limit; end = latest time limit in string
// returns a slice of countLogs and an error if it exists
func (arg *CountLog) FetchCompeted(ctxt context.Context, start, end string) ([]CountLog, error) {
	sql := `
		SELECT 
			count_id 
			, count_end
			, trans_date
			, branch
			, count_type
			, total
			, variance
		FROM count_log 
		WHERE count_end IS NOT NULL 
			AND count_end::date >= $1 
			AND count_end::date <= $2`

	ctx, cancel := context.WithTimeout(ctxt, 15*time.Second)
	defer cancel()

	rows, err := database.PgPool.Query(ctx, sql, start, end)
	if err != nil {
		log.Println("postgresql error   failed to query logs     err =", err)
		return []CountLog{}, err
	}
	defer rows.Close()

	vals := []CountLog{}
	for rows.Next() {
		r := CountLog{}
		err = rows.Scan(&r.CountID, &r.CountEnd, &r.TransDate, &r.Branch, &r.CountType, &r.Total, &r.Variance)
		if err != nil {
			log.Println("failed to scan count_ids     err = ", err)
			return vals, err
		}

		vals = append(vals, r)
	}
	return vals, nil
}
