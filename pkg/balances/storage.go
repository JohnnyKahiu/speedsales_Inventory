package balances

import (
	"context"
	"sync"

	"github.com/JohnnyKahiu/speedsales_inventory/database"
)

type BalDB struct {
	ProdBal map[string]map[string]Balance `json:"prod_bal"`
	mx      sync.RWMutex
}

var BalMaster BalDB

func (arg *BalDB) LoadBalMaster() error {
	sql := `SELECT 
				item_code
				, branch
				, stk_location
				, SUM(qty)
			FROM stk_mvmt_live 
			GROUP BY item_code, branch, stk_location
		`
	rows, err := database.PgPool.Query(context.Background(), sql)
	if err != nil {
		return err
	}
	defer rows.Close()

	vals := []Balance{}
	for rows.Next() {
		var r Balance
		err := rows.Scan(&r.ItemCode, &r.Branch, &r.StkLocation, &r.Bal)
		if err != nil {
			return err
		}
		vals = append(vals, r)
	}

	return nil
}
