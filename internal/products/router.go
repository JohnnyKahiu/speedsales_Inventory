package products

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/JohnnyKahiu/speedsales_inventory/pkg/products"
	"github.com/gorilla/mux"
)

func PostRoutes(w http.ResponseWriter, r *http.Request) map[string]interface{} {
	respMap := make(map[string]interface{})

	vars := mux.Vars(r)
	m := vars["module"]

	if m == "new" || m == "create_new" {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println("error. failed to read stock_master params    err =", err)
			respMap["response"] = "error"
			respMap["message"] = "error getting params"
			return respMap
		}
		fmt.Println("params = ", string(b))

		var stmas products.StockMaster
		json.Unmarshal(b, &stmas)
		fmt.Println("\t stk_mas = ", string(stmas.ItemCode))
		fmt.Println("\t is_active = ", stmas.IsActive)

		err = stmas.CreateNew()
		if err != nil {
			respMap["response"] = "error"
			respMap["message"] = "creating item failed"
			respMap["trace"] = err
			return respMap
		}

		return respMap
	}
	if m == "link_code" {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println("error. failed to read stock_master params    err =", err)
			respMap["response"] = "error"
			respMap["message"] = "error getting params"
			return respMap
		}

		var codeTrans products.CodeTranslator
		err = json.Unmarshal(b, &codeTrans)
		if err != nil {
			respMap["response"] = "error"
			respMap["message"] = "failed decoding params"
			respMap["trace"] = err
			return respMap
		}

	}
	if m == "add_to_combo" {
		return respMap
	}
	if m == "vats" {

	}

	return respMap
}
