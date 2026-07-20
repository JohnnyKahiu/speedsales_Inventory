package ledger

import (
	"context"
	"log"
	"time"

	"github.com/JohnnyKahiu/speedsales_inventory/database"
)

type Trail struct {
	TransDate   time.Time `json:"trans_date"`
	ItemCode    string    `json:"item_code"`
	Description string    `json:"description"`
	OpenBal     float64   `json:"open_bal"`
	In          float64   `json:"in"`
	Out         float64   `json:"out"`
	Balance     float64   `json:"balance"`
	LocationIDs []int
	Start       string
	ItemName    string
	End         string
}

func (arg *Trail) FetchTrail(ctx context.Context) ([]Trail, error) {
	// Compute opening balance from all transactions before the start date.
	var openingBal float64
	err := database.PgPool.QueryRow(ctx,
		`SELECT coalesce(SUM(qty_in - qty_out), 0)
		 FROM txn_log l
		 INNER JOIN stock_master m ON m.item_code = l.item_code
		 WHERE l.item_code = $1 AND l.trans_date::date < $2`,
		arg.ItemCode, arg.Start,
	).Scan(&openingBal)
	if err != nil {
		return []Trail{}, err
	}

	sql := `SELECT
				l.trans_date
				, l.description
				, l.item_code
				, qty_in
				, qty_out
				, m.item_name
			FROM txn_log l INNER JOIN stock_master m ON m.item_code = l.item_code
			WHERE l.item_code = $1
				AND l.trans_date::date >= $2 AND l.trans_date::date <= $3
			ORDER BY l.trans_date`

	rows, err := database.PgPool.Query(ctx, sql, arg.ItemCode, arg.Start, arg.End)
	if err != nil {
		return []Trail{}, err
	}
	defer rows.Close()

	vals := []Trail{}
	cTotal := openingBal
	for rows.Next() {
		r := Trail{}

		err = rows.Scan(&r.TransDate, &r.Description, &r.ItemCode, &r.In, &r.Out, &arg.ItemName)
		if err != nil {
			log.Println("scan error.   err =", err)
			return []Trail{}, err
		}

		if r.Description == "adopted stock count" {
			r.OpenBal = r.In
			cTotal = r.OpenBal

			r.In = 0
		}

		cTotal += (r.In - r.Out)
		r.Balance = cTotal

		vals = append(vals, r)
	}

	return vals, nil
}
