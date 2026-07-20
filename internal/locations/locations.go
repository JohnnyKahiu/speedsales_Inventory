package locations

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/JohnnyKahiu/speedsales_inventory/pkg/authentication"
	"github.com/JohnnyKahiu/speedsales_inventory/pkg/products"
	"github.com/gorilla/mux"
)

func POST(w http.ResponseWriter, r *http.Request) map[string]interface{} {
	respMap := make(map[string]interface{})

	vars := mux.Vars(r)

	m := vars["module"]

	switch m {
	case "new_branch":
		b, err := io.ReadAll(r.Body)
		if err != nil {
			respMap["response"] = "error"
			respMap["message"] = "invalid params"
			return respMap
		}

		loc := products.Locations{}
		if err = json.Unmarshal(b, &loc); err != nil {
			log.Println("failed to unmarshal location json    err =", err)
			respMap["response"] = "error"
			respMap["message"] = "json error"
			return respMap
		}

		if loc.StoreName == "" || loc.StorageLoc == "" {
			respMap["response"] = "error"
			respMap["message"] = "branch and location name are required"
			return respMap
		}

		if err = loc.GenNew(r.Context()); err != nil {
			log.Println("failed to create new stock location    err =", err)
			respMap["response"] = "error"
			respMap["message"] = "failed to create location"
			return respMap
		}

		respMap["response"] = "success"
		respMap["message"] = "Location created"
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

func GET(w http.ResponseWriter, r *http.Request) map[string]interface{} {
	respMap := make(map[string]interface{})

	vars := mux.Vars(r)
	m := vars["module"]

	switch m {
	case "list":
		details := authentication.User{}
		if err := json.Unmarshal([]byte(r.Header.Get("user_details")), &details); err != nil {
			respMap["response"] = "error"
			respMap["message"] = "user error"
			return respMap
		}

		loc := products.Locations{StoreName: details.ResolveBranch(r.URL.Query().Get("branch"))}
		vals, err := loc.Fetch()
		if err != nil {
			log.Println("failed to fetch locations    err =", err)
			respMap["response"] = "error"
			respMap["message"] = "failed to fetch locations"
			return respMap
		}

		respMap["response"] = "success"
		respMap["values"] = vals
		return respMap

	case "all":
		loc := products.Locations{}
		vals, err := loc.FetchAll()
		if err != nil {
			log.Println("failed to fetch all locations    err =", err)
			respMap["response"] = "error"
			respMap["message"] = "failed to fetch locations"
			return respMap
		}

		respMap["response"] = "success"
		respMap["values"] = vals
		return respMap
	}

	return respMap
}
