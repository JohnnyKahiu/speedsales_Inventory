package products

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/JohnnyKahiu/speedsales_inventory/pkg/products"
	"github.com/gorilla/mux"
)

// PostRoutes handles POST requests for products
// returns a map[string]interface{} with the response
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
		err = json.Unmarshal(b, &stmas)
		if err != nil {
			log.Println("json unmarshalling error    err =", err)
		}
		fmt.Println("\t stk_mas = ", string(stmas.ItemCode))
		fmt.Println("\t is_active = ", stmas.IsActive)
		fmt.Println("\t description = ", stmas.Description)

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

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err = codeTrans.New(ctx)
		if err != nil {
			respMap["response"] = "error"
			respMap["message"] = "failed"
			respMap["trace"] = err
			return respMap
		}

		respMap["response"] = "success"
		return respMap
	}
	if m == "add_to_combo" {
		return respMap
	}
	if m == "vats" {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println("error. failed to read stock_master params    err =", err)
			respMap["response"] = "error"
			respMap["message"] = "error getting params"
			return respMap
		}
		fmt.Println("params = ", string(b))

		v := make(map[string]float64)
		err = json.Unmarshal([]byte(b), &v)
		if err != nil {
			respMap["response"] = "error"
			respMap["message"] = "unable to marshal json"
			respMap["trace"] = err
			return respMap
		}

		// add a new_VAT
		err = products.ProdMaster.NewVat(v)
		if err != nil {
			respMap["response"] = "error"
			respMap["message"] = "unable to create new_vat"
			respMap["trace"] = err
			return respMap
		}

	}

	return respMap
}
