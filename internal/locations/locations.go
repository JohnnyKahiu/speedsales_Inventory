package locations

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/JohnnyKahiu/speedsales_inventory/pkg/products"
	"github.com/gorilla/mux"
)

func POST(w http.ResponseWriter, r *http.Request) map[string]interface{} {
	respMap := make(map[string]interface{})

	vars := mux.Vars(r)

	m := vars["module"]

	switch m {
	case "new_branch":
		return respMap

	case "add_to_stock_list":
		b, err := io.ReadAll(r.Body)
		if err != nil {
			respMap["response"] = "error"
			respMap["message"] = "invalid params"
			respMap["trace"] = err
			return respMap
		}
		fmt.Printf("request params = %s\n\n", b)

		loc := products.Locations{}
		err = json.Unmarshal(b, &loc)
		if err != nil {
			log.Println("failed to unmarshal json    err =", err)
			respMap["response"] = "error"
			respMap["message"] = "json error"
			respMap["trace"] = err
			return respMap
		}
		fmt.Printf("request params = %s\n\n", b)

		if loc.StockList == nil {
			loc.StockList = []string{}
		}

		err = loc.Details(r.Context())
		if err != nil {
			log.Println("failed to fetch details     err =", err)
			respMap["response"] = "error"
			respMap["message"] = "failed to fetch location details"
			return respMap
		}
		fmt.Printf("location details = %v\n", loc)

		// check if item exists in stock_list
		if len(loc.StockList) > 0 {
			fmt.Println("exists items in stock_list")
			for _, item := range loc.StockList {
				if item == loc.ItemCode && item != "" {
					respMap["response"] = "success"
					return respMap
				}
			}
		}

		// add item to stock_list
		loc.StockList = append(loc.StockList, loc.ItemCode)

		// update changes to database
		if err := loc.AddToStockList(r.Context()); err != nil {
			log.Println("failed to add item to location's stock list     err =", err)
			respMap["response"] = "error"
			respMap["message"] = "failed to add items"
			respMap["trace"] = err
			return respMap
		}

		respMap["response"] = "success"
		return respMap
	}

	return respMap
}
