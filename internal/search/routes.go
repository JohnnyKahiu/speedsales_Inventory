package search

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/JohnnyKahiu/speedsales_inventory/pkg/authentication"
	"github.com/JohnnyKahiu/speedsales_inventory/pkg/products"
	"github.com/gorilla/mux"
)

// GetRoutes handles GET requests for products
// returns a map[string]interface{} with the response
func GetRoutes(w http.ResponseWriter, r *http.Request) map[string]interface{} {
	respMap := make(map[string]interface{})

	userStr := r.Header.Get("user_details")
	if userStr == "" {
		respMap["response"] = "error"
		respMap["message"] = "user details not found"
		return respMap
	}

	details := authentication.User{}
	err := json.Unmarshal([]byte(userStr), &details)
	if err != nil {
		return respMap
	}

	vars := mux.Vars(r)

	m := vars["module"]

	switch m {
	case "code":
		fmt.Println("searching by code")
		key := r.URL.Query().Get("key")

		loc := r.Header.Get("location_id")
		locID := int64(0)
		if loc != "" {
			locID, err = strconv.ParseInt(loc, 10, 64)
			if err != nil {
				locID = 0
			}
		}

		// get values from database
		value, err := products.GetByCode(key, false, locID)
		if err != nil {
			respMap["response"] = "error"
			respMap["message"] = "failed to get_by_code"
			respMap["trace"] = err
			return respMap
		}

		respMap["response"] = "success"
		if value.ItemCode == "" {
			respMap["values"] = ""
			return respMap
		}
		respMap["values"] = value
		return respMap

	case "name":
		key := r.URL.Query().Get("key")

		loc := r.Header.Get("location_id")
		locID := int64(0)
		if loc != "" {
			locID, err = strconv.ParseInt(loc, 10, 64)
			if err != nil {
				locID = 0
			}
		}

		// get values from database
		values, err := products.SearchDescription(key, locID)
		if err != nil {
			respMap["response"] = "error"
			respMap["message"] = "failed searching"
			respMap["trace"] = err
			return respMap
		}

		respMap["response"] = "success"
		respMap["values"] = values
		return respMap

	case "department":
		keys := r.URL.Query().Get("key")

		fmt.Println("searching dept name   keyword =", keys)

		vals, err := products.SearchDeptByName(keys)
		if err != nil {
			respMap["response"] = "error"
			respMap["message"] = "failed to fetch categories"
			respMap["trace"] = err
			return respMap
		}

		fmt.Println("values =", vals)

		respMap["response"] = "success"
		respMap["values"] = vals
		return respMap

	case "category":
		keys := r.URL.Query().Get("key")

		loc := r.Header.Get("location_id")
		locID := int64(0)
		if loc != "" {
			locID, err = strconv.ParseInt(loc, 10, 64)
			if err != nil {
				locID = 0
			}
		}

		// code := r.URL.Query().Get("code")
		fmt.Println("searching dept items for code ", keys)

		values := []products.StockMaster{}

		if keys == "" {
			values, err := products.All(100)
			if err != nil {
				respMap["response"] = "error"
				respMap["message"] = "failed to fetch categories"
				respMap["trace"] = err
				return respMap
			}

			respMap["response"] = "success"
			respMap["values"] = values
			return respMap
		}

		for _, key := range strings.Split(keys, ",") {
			vals, err := products.SearchByCategory(key, locID)
			if err != nil {
				respMap["response"] = "error"
				respMap["message"] = "failed to fetch categories"
				respMap["trace"] = err
				return respMap
			}

			values = append(values, vals...)
		}

		respMap["response"] = "success"
		respMap["values"] = values
		return respMap

	case "inventory_items":
		return respMap

	case "all":
		l := r.URL.Query().Get("limit")
		if l == "" {
			l = "100"
		}

		limit, _ := strconv.ParseInt(l, 10, 0)
		vals, err := products.All(int(limit))
		if err != nil {
			respMap["response"] = "error"
			respMap["message"] = "failed to fetch categories"
			respMap["trace"] = err
			return respMap
		}

		respMap["response"] = "success"
		respMap["values"] = vals
		return respMap

	case "all_link_codes":
		key := r.URL.Query().Get("key")

		c := products.CodeTranslator{}

		c.MasterCode = key
		vals, err := c.GetAllLinks()
		if err != nil {
			respMap["response"] = "error"
			respMap["values"] = "failed to get all links"
			respMap["trace"] = err
			return respMap
		}

		respMap["response"] = "success"
		respMap["values"] = vals
		return respMap

	case "vats":
		vals := products.ProdMaster.Vats

		respMap["response"] = "success"
		respMap["values"] = vals
		return respMap

	case "bin-details":
		fmt.Println("fetching bin-details")
		loc_id := r.URL.Query().Get("id")
		loc := products.Locations{}
		loc.AutoID, _ = strconv.ParseInt(loc_id, 10, 64)

		err := loc.Details(r.Context())
		if err != nil {
			respMap["response"] = "error"
			respMap["message"] = "failed to fetch details"
			respMap["trace"] = err
			return respMap
		}

		respMap["response"] = "success"
		respMap["vals"] = loc
		return respMap
	}

	respMap["response"] = "error"
	respMap["message"] = "unresolved path"
	return respMap
}
