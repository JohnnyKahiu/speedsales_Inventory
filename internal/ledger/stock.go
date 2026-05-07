package ledger

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/JohnnyKahiu/speedsales_inventory/pkg/authentication"
	"github.com/JohnnyKahiu/speedsales_inventory/pkg/ledger"
	"github.com/JohnnyKahiu/speedsales_inventory/pkg/products"
	"github.com/gorilla/mux"
)

func GET(w http.ResponseWriter, r *http.Request) map[string]interface{} {
	details := authentication.User{}
	udetails := r.Header.Get("user_details")
	respMap := make(map[string]interface{})

	err := json.Unmarshal([]byte(udetails), &details)
	if err != nil {
		respMap["response"] = "error"
		respMap["message"] = "user error"
		return respMap
	}

	vars := mux.Vars(r)

	t := time.Now()

	start := r.URL.Query().Get("start")
	if start == "" {
		start = fmt.Sprintf("%d-%02d-%02d", t.Year(), t.Month(), t.Day())
	}

	end := r.URL.Query().Get("end")
	if end == "" {
		end = fmt.Sprintf("%d-%02d-%02d", t.Year(), t.Month(), t.Day())
	}

	itemCode := vars["code"]

	loc := products.Locations{StoreName: details.Branch}
	loc.GetAllLocInBranch(r.Context())

	trail := ledger.Trail{
		ItemCode:    itemCode,
		Start:       start,
		End:         end,
		LocationIDs: loc.IDS,
	}
	vals, err := trail.FetchTrail(r.Context())
	if err != nil {
		log.Println("failed to fetch trail    err =", err)
		respMap["response"] = "error"
		respMap["message"] = "failed to fetch ledger trail"
		respMap["trace"] = err
		return respMap
	}

	fmt.Println("values =", vals)

	respMap["response"] = "success"
	respMap["item_name"] = trail.ItemName
	respMap["values"] = vals
	return respMap
}
