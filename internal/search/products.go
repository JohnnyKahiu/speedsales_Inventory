package search

import (
	"github.com/JohnnyKahiu/speedsales_inventory/pkg/products"
)

type Search struct {
	ItemCode   string
	ItemName   string
	CategoryID int
	Department string
	Values     map[string]interface{} `json:"values"`
}

func (arg *Search) SearchProduct(key string) {
	respMap := make(map[string]interface{})
	if arg.ItemCode != "" {
		value, err := products.GetByCode(key, false)
		if err != nil {
			respMap["response"] = "error"
			respMap["message"] = "failed to get_by_code"
			respMap["trace"] = err
			arg.Values = respMap
			return
		}
		respMap["response"] = "success"
		respMap["values"] = value
		arg.Values = respMap
	}

	if arg.ItemName != "" {
		values, err := products.SearchDescription(key)
		if err != nil {
			respMap["response"] = "error"
			respMap["message"] = "failed searching"
			respMap["trace"] = err
			arg.Values = respMap
			return
		}
		respMap["response"] = "success"
		respMap["values"] = values
		arg.Values = respMap
	}

	if arg.CategoryID != 0 {

	}
}
