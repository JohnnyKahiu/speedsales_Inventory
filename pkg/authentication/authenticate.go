package authentication

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// User structure of user record
type User struct {
	table              string    `name:"users" type:"table"`
	AutoId             int64     `json:"auto_id" name:"auto_id" type:"field" sql:"BIGSERIAL PRIMARY KEY"`
	FirstName          string    `json:"first_name" name:"first_name" type:"field" sql:"VARCHAR"`
	LastName           string    `json:"last_name" name:"last_name" type:"field" sql:"VARCHAR"`
	OtherName          string    `json:"other_name" name:"other_name" type:"field" sql:"VARCHAR"`
	Telephone          string    `json:"telephone" name:"telephone" type:"field" sql:"VARCHAR NOT NULL DEFAULT ''"`
	Status             string    `json:"status" name:"status" type:"field" sql:"VARCHAR"`
	Username           string    `json:"username" name:"username" type:"field" sql:"VARCHAR NOT NULL UNIQUE"`
	Email              string    `json:"email" name:"email" type:"field" sql:"VARCHAR NOT NULL DEFAULT ''"`
	CompanyID          int64     `json:"company_id" name:"company_id" type:"field" sql:"BIGINT NOT NULL DEFAULT '0'"`
	UserClass          string    `json:"user_class" name:"user_class" type:"field" sql:"VARCHAR NOT NULL DEFAULT 'user'"`
	password           string    `name:"password" type:"field" sql:"VARCHAR NOT NULL DEFAULT ''"`
	RemoteLogin        bool      `json:"remote_login" name:"remote_login" type:"field" sql:"BOOL NOT NULL DEFAULT 'FALSE'"`
	AdoptStockcount    bool      `json:"adopt_stockcount" name:"adopt_stockcount" type:"field" sql:"BOOL NOT NULL DEFAULT 'FALSE'"`
	CompleteStockcount bool      `json:"complete_stockcount" name:"complete_stockcount" type:"field" sql:"BOOL NOT NULL DEFAULT 'FALSE'"`
	StkLocation        string    `json:"stk_location" name:"stk_location" type:"field" sql:"VARCHAR NOT NULL DEFAULT 'shop'"`
	SessionID          string    `json:"session_id" name:"session_id" type:"field" sql:"VARCHAR NOT NULL DEFAULT ''"`
	PostDispatch       bool      `json:"post_dispatch" name:"post_dispatch" type:"field" sql:"BOOL NOT NULL DEFAULT 'FALSE'"`
	ApproveDispatch    bool      `json:"approve_dispatch" name:"approve_dispatch" type:"field" sql:"BOOL NOT NULL DEFAULT 'FALSE'"`
	PostReceive        bool      `json:"post_receive" name:"post_receive" type:"field" sql:"BOOL NOT NULL DEFAULT 'FALSE'"`
	ApproveReceive     bool      `json:"approve_receive" name:"approve_receive" type:"field" sql:"BOOL NOT NULL DEFAULT 'FALSE'"`
	PostOrders         bool      `json:"post_orders" name:"post_orders" type:"field" sql:"BOOL NOT NULL DEFAULT 'FALSE'"`
	ApproveOrders      bool      `json:"approve_orders" name:"approve_orders" type:"field" sql:"BOOL NOT NULL DEFAULT 'FALSE'"`
	PriceChange        bool      `json:"price_change" name:"price_change" type:"field" sql:"BOOL NOT NULL DEFAULT 'FALSE'"`
	GrantPriceChange   bool      `json:"grant_price_change" name:"grant_price_change" type:"field" sql:"BOOL NOT NULL DEFAULT 'FALSE'"`
	CreateStock        bool      `json:"create_stock" name:"create_stock" type:"field" sql:"BOOL NOT NULL DEFAULT 'FALSE'"`
	LinkStock          bool      `json:"link_stock" name:"link_stock" type:"field" sql:"BOOL NOT NULL DEFAULT 'FALSE'"`
	CompleteStockTake  bool      `json:"complete_stock_take" name:"complete_stock_take" type:"field" sql:"BOOL NOT NULL DEFAULT 'FALSE'"`
	Produce            bool      `json:"produce" name:"produce" type:"field" sql:"BOOL NOT NULL DEFAULT 'FALSE'"`
	TillOpened         time.Time `json:"till_opened" name:"till_opened" type:"field" sql:"TIMESTAMP "`
	Till               bool      `json:"till" name:"till" type:"field" sql:"BOOL NOT NULL DEFAULT 'FALSE'"`
	TillNum            string    `json:"till_num" name:"till_num" type:"field" sql:"VARCHAR(50) "`
	Device             string    `json:"device" name:"device" type:"field" sql:"VARCHAR(50)"`
	Token              string    `json:"token" name:"token" type:"field" sql:"VARCHAR(150)"`
	TokenDate          time.Time `json:"token_date" name:"token_date" type:"field" sql:"TIMESTAMP"`
	Reset              bool      `json:"reset" name:"reset" type:"field" sql:"BOOL NOT NULL DEFAULT 'FALSE'"`
	Passcode           string    `json:"passcode"`
	SessionIDs         []string  `name:"session_ids" `
}

func ValidateJWT(tokenStr string) (Users, bool) {
	var user Users
	token, _ := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", "HMAC")
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return mySigningKey, nil
	})

	// check if token is valid and get claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		var err error

		// get today's date to compare with token's expiry
		today := time.Now()

		// get expiry date to compare
		expiry, _ := time.Parse("2006-01-02 15:04", fmt.Sprintf("%v", claims["exp"]))

		// check if token is expired
		if today.After(expiry) {
			return user, false
		}

		username := fmt.Sprintf("%v", claims["username"])

		var sessionIDs []string

		sessionIDs, err = getSessionID(username)
		if err != nil {
			return user, false
		}

		// check if session key exists for user
		for _, sessionID := range sessionIDs {
			if fmt.Sprintf("%v", claims["session"]) != sessionID {
				log.Printf("\tsession id: %v is not same as \n\t claims id: %v\n\n", sessionID, claims["session"])
				continue
			}

			// convert map to json
			jsonStr, _ := json.Marshal(claims["rights"])

			// convert json to struct
			json.Unmarshal(jsonStr, &user)

			return user, true
		}
	}
	return user, false
}

func getSessionID(username string) (string, error) {

	return "", nil
}
