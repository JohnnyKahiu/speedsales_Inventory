package supplier

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/JohnnyKahiu/speedsales_inventory/pkg/authentication"
	"github.com/JohnnyKahiu/speedsales_inventory/pkg/suppliers"
	"github.com/gorilla/mux"
)

func POST(w http.ResponseWriter, r *http.Request) map[string]interface{} {
	respMap := make(map[string]interface{})

	details := authentication.User{}
	udetails := r.Header.Get("user_details")

	err := json.Unmarshal([]byte(udetails), &details)
	if err != nil {
		respMap["response"] = "error"
		respMap["message"] = "user error"
		return respMap
	}

	vars := mux.Vars(r)

	m := vars["module"]

	switch m {

	case "new":
		fmt.Println("new suppliers")
		b, err := io.ReadAll(r.Body)
		if err != nil {
			respMap["response"] = "error"
			respMap["message"] = "params error"
			return respMap
		}
		fmt.Printf("\n\tvalues = %s \n", b)

		supp := suppliers.Supplier{}
		err = json.Unmarshal(b, &supp)
		if err != nil {
			log.Println("error jsonify     err trace", err)
			respMap["response"] = "error"
			respMap["message"] = "jsonify params error"
			respMap["trace"] = err
			return respMap
		}

		err = supp.New(r.Context())
		if err != nil {
			respMap["response"] = "error"
			respMap["message"] = "failed to create new supplier"
			respMap["trace"] = err
			return respMap
		}

		respMap["response"] = "success"
		return respMap
	}

	return respMap
}

func GET(w http.ResponseWriter, r *http.Request) map[string]interface{} {
	respMap := make(map[string]interface{})

	details := authentication.User{}
	udetails := r.Header.Get("user_details")

	err := json.Unmarshal([]byte(udetails), &details)
	if err != nil {
		respMap["response"] = "error"
		respMap["message"] = "user error"
		return respMap
	}

	vars := mux.Vars(r)

	m := vars["module"]

	switch m {

	case "name":
		key := r.URL.Query().Get("key")

		supp := suppliers.Supplier{
			SuppName: key,
		}

		values, err := supp.Search(r.Context())
		if err != nil {
			respMap["response"] = "error"
			respMap["message"] = "failed to create new supplier"
			respMap["trace"] = err
			return respMap
		}

		respMap["response"] = "success"
		respMap["values"] = values
		return respMap

	case "all":
		fmt.Println("fetching all suppliers")
		supp := suppliers.Supplier{}
		values, err := supp.FetchAll(r.Context())
		if err != nil {
			respMap["response"] = "error"
			respMap["message"] = "failed to fetch all suppliers"
			respMap["trace"] = err
			return respMap
		}

		respMap["response"] = "success"
		respMap["values"] = values
		return respMap
	}

	return respMap
}
