package counts

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/JohnnyKahiu/speedsales_inventory/pkg/count"
	"github.com/JohnnyKahiu/speedsales_inventory/pkg/products"
	"github.com/gorilla/mux"
)

func POST(r *http.Request) map[string]interface{} {
	respMap := make(map[string]interface{})

	vars := mux.Vars(r)

	m := vars["module"]

	switch m {
	case "create", "new":
		b, err := io.ReadAll(r.Body)
		if err != nil {
			respMap["response"] = "error"
			respMap["message"] = "invalid params"
			respMap["trace"] = err
			return respMap
		}
		fmt.Printf("params = %s\n", b)

		clog := count.CountLog{}
		err = json.Unmarshal(b, &clog)
		if err != nil {
			respMap["response"] = "error"
			respMap["message"] = "failed to decode params"
			return respMap
		}

		err = clog.New(r.Context())
		if err != nil {
			respMap["response"] = "error"
			respMap["message"] = "failed to create new count"
			respMap["trace"] = err
			return respMap
		}

		respMap["response"] = "success"
		return respMap

	case "count":
		fmt.Println("counting item")
		b, err := io.ReadAll(r.Body)
		if err != nil {
			respMap["response"] = "error"
			respMap["message"] = "params error"
			respMap["trace"] = err
			return respMap
		}
		fmt.Printf("counting item \n params = %s \n", b)

		ci := count.CountItems{}
		err = json.Unmarshal(b, &ci)
		if err != nil {
			respMap["response"] = "error"
			respMap["message"] = "params error"
			respMap["trace"] = err
			return respMap
		}

		prod, err := products.GetByCode(ci.ItemCode, true, ci.LocationID)
		if err != nil {
			log.Println("error getting products    err =", err)
			respMap["response"] = "error"
			respMap["message"] = "failed to fetch products"
			return respMap
		}

		if prod.ItemCode == "" {
			log.Println("error code is null")
			respMap["response"] = "error"
			respMap["message"] = "invalid product"
			return respMap
		}

		ci.SystemBal = prod.Bal
		ci.Counted = (ci.Cases * float64(prod.UnitsPerPack)) + ci.Pieces

		fmt.Printf("\t counted = %v \n\t cases = %v \n\t items_per_case = %v\n", ci.Counted, ci.Cases, ci.ItemsPerCase)
		if err = ci.Count(r.Context()); err != nil {
			log.Println("error failed to count_items    err =", err)
			respMap["response"] = "error"
			respMap["message"] = "failed to count item"
			respMap["trace"] = err
			return respMap
		}

		return respMap

	case "complete":
		return respMap
	}

	return respMap
}

// GET http get method for stock_Count operations
func GET(w http.ResponseWriter, r *http.Request) map[string]interface{} {
	respMap := make(map[string]interface{})

	vars := mux.Vars(r)

	m := vars["module"]

	switch m {
	case "active":
		c := count.CountLog{}
		vals, err := c.Active(r.Context())
		if err != nil {
			respMap["response"] = "error"
			respMap["message"] = "failed to get active counts"
			respMap["trace"] = err
			return respMap
		}

		respMap["response"] = "success"
		respMap["values"] = vals
		return respMap

	case "count_item_in_bin":
		loc := r.URL.Query().Get("bin")
		locID, _ := strconv.ParseInt(loc, 10, 64)

		count_id := r.URL.Query().Get("count_id")
		CountID, _ := strconv.ParseInt(count_id, 10, 64)

		countLog := count.CountLog{
			CountID: CountID,
			Bin:     locID,
		}

		err := countLog.FetchItems(r.Context())
		if err != nil {
			respMap["response"] = "error"
			respMap["message"] = "failed to fetch count items"
			respMap["trace"] = err
			return respMap
		}
		respMap["response"] = "success"
		respMap["values"] = countLog.Items
		return respMap

	case "bins":
		return respMap

	}

	return respMap
}
