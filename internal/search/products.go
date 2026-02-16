package search

import (
	"github.com/JohnnyKahiu/speedsales_inventory/pkg/products"
)

type Search struct {
	ItemCode   string                 `json:"item_code"`
	ItemName   string                 `json:"item_name"`
	CategoryID int                    `json:"category_id"`
	Department string                 `json:"department"`
	Values     []products.StockMaster `json:"values"`
	Value      products.StockMaster   `json:"value"`
	Response   map[string]interface{} `json:"response"`
}

func (arg *Search) SearchProduct() {
	respMap := make(map[string]interface{})
	if arg.ItemCode != "" {
		value, err := products.GetByCode(arg.ItemCode, false)
		if err != nil {
			respMap["response"] = "error"
			respMap["message"] = "failed to get_by_code"
			respMap["trace"] = err
			arg.Response = respMap

			return
		}
		respMap["response"] = "success"
		respMap["values"] = value
		arg.Value = value
		arg.Response = respMap

		return
	}

	if arg.ItemName != "" {
		values, err := products.SearchDescription(arg.ItemName)
		if err != nil {
			respMap["response"] = "error"
			respMap["message"] = "failed searching"
			respMap["trace"] = err
			arg.Response = respMap

			return
		}

		respMap["response"] = "success"
		respMap["values"] = values
		arg.Values = values
		arg.Response = respMap

		return
	}

	if arg.CategoryID != 0 {
		return
	}
}
