package purchases

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/JohnnyKahiu/speedsales_inventory/database"
)

type QueryLog struct {
	GrnNum   int64    `json:"grn_num"`
	Start    string   `json:"start"`
	End      string   `json:"end"`
	Poster   string   `json:"poster"`
	Supplier string   `json:"supplier"`
	State    []string `json:"state"`
}

// FetchGrnList - gets all grn logs within a given range
// Queries from grn_log
// returns a slice of grn_log and an error
func (arg *QueryLog) FetchGrnList(ctxt context.Context) ([]GrnLog, error) {
	posterCon := ""
	if arg.Poster != "" {
		posterCon = fmt.Sprintf("AND poster = '%v'", arg.Poster)
	}

	suppCon := ""
	if arg.Supplier != "" {
		suppName := "%" + strings.ReplaceAll(arg.Supplier, "'", "|| chr(39) ||") + "%"
		suppCon = fmt.Sprintf("AND supp_name ILIKE '%v'", suppName)
	}

	timeCond := "AND trans_date::date >= $2 AND trans_date::date <= $3"
	if arg.Start == "-1" || arg.End == "-1" {
		timeCond = "AND $2 = $2 AND $3 = $3"
	}

	sql := fmt.Sprintf(`
		SELECT 
			trans_date, grn_num, supp_name, inv_num, supp_pin
			, inv_type, tims_cuin, total_exc, total_amount_inc, total_vat, discount
			, vatable, exempt, vehicle_num, driver_name, registered_by, poster
			, inv_date, recv_date, state
		FROM grn_log 
		WHERE state = ANY($1) 
			%v
			%v 
			%v
	`, timeCond, posterCon, suppCon)

	ctx, cancel := context.WithTimeout(ctxt, 20*time.Second)
	defer cancel()

	rows, err := database.PgPool.Query(ctx, sql, arg.State, arg.Start, arg.End)
	if err != nil {
		log.Println("postgresql error.   failed to query grn_log details     err =", err)
		return []GrnLog{}, err
	}
	defer rows.Close()

	vals := []GrnLog{}
	for rows.Next() {
		r := GrnLog{}
		if err := rows.Scan(&r.TransDate, &r.GrnNum, &r.SuppName, &r.InvNum, &r.SuppPin,
			&r.InvType, &r.TimsCUIN, &r.TotalExc, &r.TotalAmountInc, &r.TotalVat, &r.Discount,
			&r.Vatable, &r.Exempt, &r.VehicleNum, &r.DriverName, &r.RegisteredBy, &r.Poster,
			&r.InvDate, &r.RecvDate, &r.State); err != nil {
			log.Println("scan error.    err =", err)
			return vals, err
		}

		vals = append(vals, r)
	}

	return vals, nil
}
