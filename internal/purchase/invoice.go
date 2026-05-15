package purchase

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/JohnnyKahiu/speedsales_inventory/pkg/authentication"
	"github.com/JohnnyKahiu/speedsales_inventory/pkg/products"
	"github.com/JohnnyKahiu/speedsales_inventory/pkg/purchases"
	"github.com/gorilla/mux"
)

func POST(w http.ResponseWriter, r *http.Request) map[string]interface{} {
	respMap := make(map[string]interface{})

	vars := mux.Vars(r)
	m := vars["module"]

	details := authentication.User{}
	err := json.Unmarshal([]byte(r.Header.Get("user_details")), &details)
	if err != nil {
		respMap["response"] = "error"
		respMap["message"] = "authentication error"
	}

	switch m {
	case "new":
		b, err := io.ReadAll(r.Body)
		if err != nil {
			respMap["response"] = "error"
			respMap["message"] = "params error"
			return respMap
		}
		fmt.Printf("new grn params = %s\n", b)

		grn := purchases.GrnLog{RegisteredBy: details.Username}
		err = json.Unmarshal(b, &grn)
		if err != nil {
			respMap["response"] = "error"
			respMap["message"] = "failed to unmarshal grn"
			respMap["trace"] = err
		}
		fmt.Printf("new decoded grn params = %v\n", grn)

		err = grn.GetGrn(r.Context())
		if err != nil {
			log.Println("failed to generate GRN     err =", err)
			respMap["response"] = "error"
			respMap["message"] = "failed to generate GRN Num"
			return respMap
		}
		respMap["response"] = "success"
		return respMap

	case "add-item":
		item := purchases.GrnItem{}

		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			respMap["response"] = "error"
			respMap["message"] = "params error"
			respMap["trace"] = err
			return respMap
		}

		loc := products.Locations{
			StoreName:  details.Branch,
			StorageLoc: details.StkLocation,
			ItemCode:   item.ItemCode,
		}

		err := loc.GetSaleLoc(r.Context())
		if err != nil {
			respMap["response"] = "error"
			respMap["message"] = "params error"
			respMap["trace"] = err
			return respMap
		}

		item.LocationID = loc.AutoID
		err = item.AddItem(r.Context())
		if err != nil {
			log.Println("error. failed to add grn_item     err =", err)
			respMap["response"] = "error"
			respMap["message"] = "failed to add item"
			respMap["trace"] = err
			return respMap
		}

		respMap["response"] = "success"
		respMap["auto_id"] = item.AutoID
		respMap["values"] = item
		return respMap

	case "complete":
		grnLog := purchases.GrnLog{}

		b, err := io.ReadAll(r.Body)
		if err != nil {
			respMap["response"] = "error"
			respMap["message"] = "params error"
			respMap["trace"] = err
			return respMap
		}

		json.Unmarshal(b, &grnLog)
		if err := grnLog.Details(r.Context()); err != nil {
			log.Println("failed to get details    err =", err)
			respMap["response"] = "error"
			respMap["message"] = "failed to get details"
			respMap["trace"] = err
			return respMap
		}

		grnLog.Poster = details.Username

		err = grnLog.Complete(r.Context())
		if err != nil {
			log.Println("failed to complete posting     err =", err)
			respMap["response"] = "error"
			respMap["message"] = "failed to complete grn"
			respMap["trace"] = err
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
	case "pending":
		key := r.URL.Query().Get("grn_num")

		grnNum, _ := strconv.ParseInt(key, 10, 64)
		grnLog := purchases.GrnLog{GrnNum: grnNum}

		values, err := grnLog.Pending(r.Context())
		if err != nil {
			respMap["response"] = "error"
			respMap["message"] = "failed to get pending grns"
			respMap["traace"] = err
			return respMap
		}

		respMap["response"] = "success"
		respMap["values"] = values
		return respMap

	case "restore":
		key := r.URL.Query().Get("grn_num")

		grnNum, _ := strconv.ParseInt(key, 10, 64)
		grnLog := purchases.GrnLog{GrnNum: grnNum}

		err := grnLog.GetItems(r.Context())
		if err != nil {
			respMap["response"] = "error"
			respMap["message"] = "failed to get grn items"
			respMap["traace"] = err
			return respMap
		}

		respMap["response"] = "success"
		respMap["values"] = grnLog.Items
		return respMap

	case "grn_list":
		start := r.URL.Query().Get("start")
		end := r.URL.Query().Get("start")

		query := purchases.QueryLog{
			Start: start,
			End:   end,
			State: []string{"POSTED"},
		}

		t := time.Now()
		fmt.Printf("start_date = %v\n end_date = %v\n", start, end)
		if query.Start == "" || query.Start == "null" {
			lw := t.AddDate(0, 0, -7)
			query.Start = fmt.Sprintf("%d-%02d-%02d", lw.Year(), lw.Month(), lw.Day())
		}

		if query.End == "" || query.End == "null" {
			query.End = fmt.Sprintf("%d-%02d-%02d", t.Year(), t.Month(), t.Day())
		}

		fmt.Println("query params =", query)

		list, err := query.FetchGrnList(r.Context())
		if err != nil {
			respMap["response"] = "error"
			respMap["message"] = "failed to query grn list"
			return respMap
		}

		respMap["response"] = "success"
		respMap["values"] = list
		return respMap

	case "pc_list":
		query := purchases.QueryLog{
			Start: "-1",
			End:   "-1",
			State: []string{"PRICE CHANGE"},
		}

		t := time.Now()
		if query.Start == "" || query.Start == "null" {
			lw := t.AddDate(0, 0, -7)
			query.Start = fmt.Sprintf("%d-%02d-%02d", lw.Year(), lw.Month(), lw.Day())
		}

		if query.End == "" || query.End == "null" {
			query.End = fmt.Sprintf("%d-%02d-%02d", t.Year(), t.Month(), t.Day())
		}

		fmt.Println("query params =", query)

		list, err := query.FetchGrnList(r.Context())
		if err != nil {
			respMap["response"] = "error"
			respMap["message"] = "failed to query grn list"
			return respMap
		}

		respMap["response"] = "success"
		respMap["values"] = list
		return respMap

	case "receipt":
		grnNum := r.URL.Query().Get("grn_num")
		grnLog := purchases.GrnLog{}

		grnLog.GrnNum, _ = strconv.ParseInt(grnNum, 10, 64)
		fmt.Printf("\t receipt for grn %v\n", grnLog.GrnNum)

		err := grnLog.GetReceipt(r.Context())
		if err != nil {
			log.Println("error. failed to get grn receipt    err =", err)
			respMap["response"] = "error"
			respMap["message"] = "failed to get receipt"
			respMap["trace"] = err
			return respMap
		}

		fmt.Printf("\n\t grn =\n\t %v\n", grnLog.Items)
		respMap["response"] = "success"
		respMap["values"] = grnLog
		return respMap

	}

	return respMap
}

func DELETE(w http.ResponseWriter, r *http.Request) map[string]interface{} {
	respMap := make(map[string]interface{})

	vars := mux.Vars(r)
	m := vars["module"]

	switch m {
	case "grn_item":
		key := r.URL.Query().Get("auto_id")
		autoID, err := strconv.ParseInt(key, 10, 64)
		if err != nil {
			respMap["response"] = "error"
			respMap["message"] = "parse error"
			return respMap
		}

		grnItem := purchases.GrnItem{AutoID: autoID}

		if err := grnItem.Delete(r.Context()); err != nil {
			respMap["response"] = "error"
			respMap["message"] = "failed to delete grn"
			respMap["trace"] = err
			return respMap
		}

		respMap["response"] = "success"
		return respMap
	}

	return respMap
}
