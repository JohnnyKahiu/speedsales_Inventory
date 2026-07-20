package product

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/JohnnyKahiu/speedsales_inventory/pkg/authentication"
	"github.com/JohnnyKahiu/speedsales_inventory/pkg/balances"
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
	//POST: /products/new  or POST:/products/create_new
	case "new", "create_new":
		b, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println("error. failed to read stock_master params    err =", err)
			respMap["response"] = "error"
			respMap["message"] = "failed to get params"
			return respMap
		}

		var stmas products.StockMaster
		err = json.Unmarshal(b, &stmas)
		if err != nil {
			log.Println("json unmarshalling error    err =", err)
		}

		err = stmas.CreateNew()
		if err != nil {
			respMap["response"] = "error"
			respMap["message"] = "creating item failed"
			respMap["trace"] = err
			return respMap
		}

		return respMap

	// POST:/products/link_code
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

	// POST: /products/new  or POST:/products/create_new/products/add_to_combo
	// adds a product to a combo
	// expects a combo struct
	case "add_to_combo":
		return respMap

	case "recipe":
		b, err := io.ReadAll(r.Body)
		if err != nil {
			respMap["response"] = "error"
			respMap["message"] = "bad request"
			return respMap
		}

		var recipe products.Recipe
		if err = json.Unmarshal(b, &recipe); err != nil || recipe.ProdCode == "" || recipe.ItemCode == "" {
			respMap["response"] = "error"
			respMap["message"] = "prod_code, item_code are required"
			return respMap
		}

		if err = products.AddRecipeItem(recipe); err != nil {
			log.Println("error adding recipe item    err =", err)
			respMap["response"] = "error"
			respMap["message"] = "failed to save recipe item"
			return respMap
		}

		respMap["response"] = "success"
		return respMap

	// POST: /products/new  or POST:/products/create_new/products/vats
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

		v := make(map[string]float64)
		err = json.Unmarshal([]byte(b), &v)
		if err != nil {
			log.Println("error unmarshaling json    err =", err)
			respMap["response"] = "error"
			respMap["message"] = "unable to marshal json"
			respMap["trace"] = err
			return respMap
		}

		// add a new_VAT
		err = products.ProdMaster.NewVat(v)
		if err != nil {
			log.Println("error adding NewVat    err =", err)
			respMap["response"] = "error"
			respMap["message"] = "unable to create new_vat"
			respMap["trace"] = err
			return respMap
		}

		respMap["response"] = "success"
		return respMap

	// POST: /products/new  or POST:/products/create_new/products/department
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

	case "produce":
		userStr := r.Header.Get("user_details")
		user := authentication.User{}
		json.Unmarshal([]byte(userStr), &user)

		b, err := io.ReadAll(r.Body)
		if err != nil {
			respMap["response"] = "error"
			respMap["message"] = "bad request"
			return respMap
		}

		type ProductionReq struct {
			ProdCode     string  `json:"prod_code"`
			Quantity     float64 `json:"quantity"`
			Damages      float64 `json:"damages"`
			DamageReason string  `json:"damage_reason"`
			LocationID   int64   `json:"location_id"`
		}

		var req ProductionReq
		if err = json.Unmarshal(b, &req); err != nil || req.ProdCode == "" || req.Quantity <= 0 {
			respMap["response"] = "error"
			respMap["message"] = "prod_code and quantity > 0 are required"
			return respMap
		}

		recipe, err := products.FetchRecipeItems(req.ProdCode)
		if err != nil {
			respMap["response"] = "error"
			respMap["message"] = "failed to fetch recipe"
			return respMap
		}

		txnID := fmt.Sprintf("PROD-%s-%d", req.ProdCode, time.Now().UnixNano())
		ctx := context.Background()

		// credit produced item
		prodTxn := balances.TxnLog{
			Description: "PRODUCTION",
			TxnID:       txnID,
			LocationID:  req.LocationID,
			ItemCode:    req.ProdCode,
			QtyIn:       req.Quantity,
			TransDate:   time.Now(),
		}
		if err = prodTxn.LogBal(ctx); err != nil {
			log.Println("produce: failed to log produced item    err =", err)
			respMap["response"] = "error"
			respMap["message"] = "failed to record production"
			return respMap
		}

		// debit each ingredient
		for _, ing := range recipe {
			ingTxn := balances.TxnLog{
				Description: "PRODUCTION-INGREDIENT",
				TxnID:       fmt.Sprintf("%s-%s", txnID, ing.ItemCode),
				LocationID:  req.LocationID,
				ItemCode:    ing.ItemCode,
				QtyOut:      ing.Amount * req.Quantity,
				TransDate:   time.Now(),
			}
			if err = ingTxn.LogBal(ctx); err != nil {
				log.Println("produce: failed to log ingredient    err =", err)
			}
		}

		// debit damages if any
		if req.Damages > 0 {
			reason := req.DamageReason
			if reason == "" {
				reason = "DAMAGE"
			}
			dmgTxn := balances.TxnLog{
				Description: fmt.Sprintf("DAMAGE-%s", reason),
				TxnID:       fmt.Sprintf("%s-DMG", txnID),
				LocationID:  req.LocationID,
				ItemCode:    req.ProdCode,
				QtyOut:      req.Damages,
				TransDate:   time.Now(),
			}
			if err = dmgTxn.LogBal(ctx); err != nil {
				log.Println("produce: failed to log damage    err =", err)
			}
		}

		respMap["response"] = "success"
		respMap["txn_id"] = txnID
		return respMap
	}

	return respMap
}

// UpdateRoutes handles PUT requests
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

		err = dept.Update(r.Context())
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

	case "recipe":
		itemCode := vars["code"]
		prodCode := r.URL.Query().Get("prod_code")
		if itemCode == "" || prodCode == "" {
			respMap["response"] = "error"
			respMap["message"] = "item_code (path) and prod_code (query) are required"
			return respMap
		}

		if err := products.RemoveRecipeItem(prodCode, itemCode); err != nil {
			log.Println("error removing recipe item    err =", err)
			respMap["response"] = "error"
			respMap["message"] = "failed to remove recipe item"
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

	det_str := r.Header.Get("user_details")
	details := authentication.User{}
	json.Unmarshal([]byte(det_str), &details)

	switch key {
	case "departments":
		isMenu := true

		onlyMenu := r.URL.Query().Get("only_menu")
		if onlyMenu == "false" {
			isMenu = false
		}

		vals, cartegories, err := products.GetDepartments(isMenu)
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

	case "bins":
		stkLoc := products.Locations{
			StoreName: details.ResolveBranch(r.URL.Query().Get("branch")),
		}

		vals, err := stkLoc.Fetch()
		if err != nil {
			respMap["response"] = "error"
			respMap["message"] = "failed to get stk_locateions"
			respMap["trace"] = err
			return respMap
		}

		respMap["response"] = "success"
		respMap["values"] = vals
		return respMap

	case "bins_multi":
		key := r.URL.Query().Get("ids")

		autoIDs := []int{}
		err := json.Unmarshal([]byte(key), &autoIDs)
		if err != nil {
			log.Println("error failed to unmarshal ids    err =", err)
			respMap["response"] = "error"
			respMap["message"] = "failed to marshal json"
			respMap["trace"] = err
			return respMap
		}

		stkLoc := products.Locations{
			IDS: autoIDs,
		}

		vals, err := stkLoc.FetchMultiIDs(r.Context())
		if err != nil {
			respMap["response"] = "error"
			respMap["message"] = "failed to fetch items"
			respMap["trace"] = err
			return respMap
		}

		respMap["response"] = "success"
		respMap["values"] = vals
		return respMap
	}

	return respMap
}
