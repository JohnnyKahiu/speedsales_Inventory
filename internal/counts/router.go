package counts

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/JohnnyKahiu/speedsales_inventory/pkg/count"
	"github.com/gorilla/mux"
)

func POST(w http.ResponseWriter, r *http.Request) map[string]interface{} {
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

	case "bins":
	}

	return respMap
}
