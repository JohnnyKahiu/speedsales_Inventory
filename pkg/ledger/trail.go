package ledger

import (
	"context"
	"fmt"
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
	fmt.Println("location_ids =", arg.LocationIDs)
	if len(arg.LocationIDs) <= 0 || arg.LocationIDs == nil {
		arg.LocationIDs = []int{}
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
				AND location_id = ANY($2)
				AND l.trans_date::date >= $3 AND l.trans_date::date <= $4
			ORDER BY l.trans_date`

	fmt.Println(sql, arg.ItemCode, arg.Start, arg.End)

	rows, err := database.PgPool.Query(ctx, sql, arg.ItemCode, arg.LocationIDs, arg.Start, arg.End)
	if err != nil {
		return []Trail{}, err
	}
	defer rows.Close()

	vals := []Trail{}
	cTotal := float64(0)
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
		}

		cTotal += (r.In - r.Out)
		r.Balance = cTotal

		vals = append(vals, r)
	}

	return vals, nil
}
