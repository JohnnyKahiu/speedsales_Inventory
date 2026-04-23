package api

import (
	"encoding/json"
	"net/http"

	"github.com/JohnnyKahiu/speedsales_inventory/pkg/authentication"
	"github.com/JohnnyKahiu/speedsales_inventory/pkg/branches"
	"github.com/gorilla/mux"
)

func AppGet(w http.ResponseWriter, r *http.Request) {
	userStr := r.Header.Get("user_details")
	if userStr == "" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"response": "error", "message": "unauthorized user"}`))
	}

	details := authentication.User{}
	err := json.Unmarshal([]byte(userStr), &details)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"response": "error", "message": "unauthorized user"}`))
		return
	}

	vars := mux.Vars(r)

	m := vars["module"]

	switch m {
	case "branches":
		branchReq := branches.Branch{}

		vals, err := branchReq.FetchAll()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"response": "error", "message": "error fetching branch details"}`))
			return
		}

		respMap := make(map[string]interface{})
		respMap["response"] = "success"
		respMap["values"] = vals

		jstr, _ := json.Marshal(respMap)

		w.WriteHeader(http.StatusOK)
		w.Write(jstr)
	}
}
