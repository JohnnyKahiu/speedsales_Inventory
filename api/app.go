package api

import (
	"encoding/json"
	"io"
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

func AppPost(w http.ResponseWriter, r *http.Request) {
	EnableCors(&w)

	vars := mux.Vars(r)
	m := vars["module"]

	respMap := make(map[string]interface{})

	switch m {
	case "branches":
		b, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"response": "error", "message": "invalid params"}`))
			return
		}

		br := branches.Branch{}
		if err = json.Unmarshal(b, &br); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"response": "error", "message": "json error"}`))
			return
		}

		if br.BranchName == "" {
			respMap["response"] = "error"
			respMap["message"] = "branch name is required"
			jstr, _ := json.Marshal(respMap)
			w.Write(jstr)
			return
		}

		if err = br.New(); err != nil {
			respMap["response"] = "error"
			respMap["message"] = "failed to create branch"
			jstr, _ := json.Marshal(respMap)
			w.Write(jstr)
			return
		}

		respMap["response"] = "success"
		respMap["message"] = "Branch created"
		respMap["auto_id"] = br.AutoID

	default:
		respMap["response"] = "error"
		respMap["message"] = "unknown module"
	}

	jstr, _ := json.Marshal(respMap)
	w.Write(jstr)
}
