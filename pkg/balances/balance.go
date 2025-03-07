package balances

import (
	"errors"
)

type Balance struct {
	ItemCode    string  `json:"item_code"`
	Branch      string  `json:"branch"`
	StkLocation string  `json:"stk_location"`
	Bal         float64 `json:"bal"`
}

func (arg *Balance) GetBal() error {
	if arg.ItemCode == "" {
		return errors.New("error. item_code is null")
	}
	if arg.Branch == "" {
		return errors.New("error. branch is null")
	}

	// sql := `SELECT `
	return nil
}
