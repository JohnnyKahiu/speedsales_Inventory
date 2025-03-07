package search

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/JohnnyKahiu/speedsales_inventory/pkg/products"
	"github.com/gorilla/mux"
)

func GetRoutes(w http.ResponseWriter, r *http.Request) map[string]interface{} {
	respMap := make(map[string]interface{})

	vars := mux.Vars(r)

	m := vars["module"]

	if m == "code" {
		fmt.Println("searching by code")
		key := r.URL.Query().Get("key")

		// get values from database
		value, err := products.GetByCode(key, false)
		if err != nil {
			respMap["response"] = "error"
			respMap["message"] = "failed to get_by_code"
			respMap["trace"] = err
			return respMap
		}

		respMap["response"] = "success"
		respMap["values"] = value
		return respMap
	}
	if m == "name" {
		key := r.URL.Query().Get("key")

		// get values from database
		values, err := products.SearchDescription(key)
		if err != nil {
			respMap["response"] = "error"
			respMap["message"] = "failed searching"
			respMap["trace"] = err
			return respMap
		}

		respMap["response"] = "success"
		respMap["values"] = values
		return respMap
	}
	if m == "category" {
		key := r.URL.Query().Get("key")

		vals, err := products.SearchByCategory(key)
		if err != nil {
			respMap["response"] = "error"
			respMap["message"] = "failed to fetch categories"
			respMap["trace"] = err
			return respMap
		}

		respMap["response"] = "success"
		respMap["values"] = vals
		return respMap
	}
	if m == "all" {
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
	}
	if m == "all_link_codes" {
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
	}

	respMap["response"] = "error"
	respMap["message"] = "unresolved path"
	return respMap
}
