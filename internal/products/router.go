package products

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
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

	switch m {
	case "new", "create_new":
		b, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println("error. failed to read stock_master params    err =", err)
			respMap["response"] = "error"
			respMap["message"] = "failed to get params"
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

	// /products/link_code
	// creates a new code_translator link to a product master_code
	// expects a code_translator struct
	// master_code
	// link_code
	// pkg_qty
	// discount
	case "link_code":
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

	// /products/add_to_combo
	// adds a product to a combo
	// expects a combo struct
	case "add_to_combo":
		return respMap

	// /products/vats
	// adds a new VAT
	// expects a map[string]float64
	case "vats":
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

		respMap["response"] = "success"
		return respMap

	// /products/department
	// creates a new department
	// expects a department struct
	case "department":
		b, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println("error. failed to read stock_master params    err =", err)
			respMap["response"] = "error"
			respMap["message"] = "error getting params"
			return respMap
		}

		var dept products.Departments
		err = json.Unmarshal([]byte(b), &dept)
		if err != nil {
			log.Println("error failed to unmarshal json    err =", err)
			respMap["response"] = "error"
			respMap["message"] = "unable to marshal json"
			return respMap
		}

		err = dept.CreateNew()
		if err != nil {
			log.Println("error, failed to create New department     err =", err)
			respMap["response"] = "error"
			respMap["message"] = "failed to create new department"
			return respMap
		}

		respMap["response"] = "success"
		return respMap
	}

	return respMap
}

func UpdateRoutes(w http.ResponseWriter, r *http.Request) map[string]interface{} {
	respMap := make(map[string]interface{})

	vars := mux.Vars(r)
	m := vars["module"]

	switch m {
	// /products/update/department -
	// updates details about department
	// it receives a dept struct
	case "department":
		b, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println("error. failed to read stock_master params    err =", err)
			respMap["response"] = "error"
			respMap["message"] = "error getting params"
			return respMap
		}

		var dept products.Departments
		err = json.Unmarshal([]byte(b), &dept)
		if err != nil {
			log.Println("error. failed to marshall params     err =", err)
			respMap["response"] = "error"
			respMap["message"] = "error getting params"
			return respMap
		}

		err = dept.Update()
		if err != nil {
			log.Println("error. failed to update department     err =", err)
			respMap["response"] = "error"
			respMap["message"] = "error updating department"
			return respMap
		}

		respMap["response"] = "success"
		return respMap
	}
	return respMap
}

// DelRoutes handles DELETE requests for products
// returns a map[string]interface{} with the response
func DelRoutes(w http.ResponseWriter, r *http.Request) map[string]interface{} {
	respMap := make(map[string]interface{})
	vars := mux.Vars(r)

	m := vars["module"]

	switch m {
	case "department":
		code, err := strconv.Atoi(vars["code"])
		if err != nil {
			log.Println("error, failed to convert code to int     err =", err)
			respMap["response"] = "error"
			respMap["message"] = "failed to convert code to int"
			return respMap
		}

		dept := products.Departments{Code: int64(code)}
		err = dept.Delete()
		if err != nil {
			log.Println("error, failed to delete department     err =", err)
			respMap["response"] = "error"
			respMap["message"] = "failed to delete department"
			return respMap
		}

		respMap["response"] = "success"
		return respMap
	}

	return respMap
}

func CatalogueGet(w http.ResponseWriter, r *http.Request) map[string]interface{} {
	respMap := make(map[string]interface{})

	vars := mux.Vars(r)
	by := vars["by"]
	switch by {
	case "supplier":

	}
	return respMap
}

func GetGroups(w http.ResponseWriter, r *http.Request) map[string]interface{} {
	respMap := make(map[string]interface{})

	vars := mux.Vars(r)
	key := vars["key"]

	switch key {
	case "departments":
		vals, cartegories, err := products.GetDepartments()
		if err != nil {
			log.Println("error failed getting departments    err =", err)
			respMap["response"] = "error"
			respMap["message"] = "failed to get departments"
			return respMap
		}
		respMap["response"] = "success"
		respMap["values"] = vals
		respMap["categories"] = cartegories
		return respMap

	case "sub_depts":
		return respMap

	case "vats":
		// vals, err := products.ProdMaster	.GetVats()
		// if err != nil {
		// 	log.Println("error failed getting vats    err =", err)
		// 	respMap["response"] = "error"
		// 	respMap["message"] = "failed to get vats"
		// 	return respMap
		// }
		respMap["response"] = "success"
		respMap["values"] = "vals"
		return respMap
	}

	return respMap
}
