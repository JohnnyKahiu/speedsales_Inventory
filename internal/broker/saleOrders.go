package broker

import (
	"context"
	"fmt"
	"time"

	"github.com/JohnnyKahiu/speedsales_inventory/pkg/balances"
	"github.com/JohnnyKahiu/speedsales_inventory/pkg/products"
)

type SalesOrder struct {
	TransDate    time.Time  `json:"trans_date" `
	CompleteTime time.Time  `json:"complete_time"`
	OrderNum     int64      `json:"order_num" `
	DailyCount   int64      `json:"daily_count" `
	OrderItems   []SaleItem `json:"order_items" `
	Poster       string     `json:"poster" `
	Branch       string     `json:"branch" `
	StkLocation  string     `json:"stk_Location"`
	DispBy       string     `json:"disp_by" `
	DispTime     time.Time  `json:"disp_time" `
	CompanyID    int64      `json:"company_id" `
	TillNum      int64      `json:"till_num" `
	PayTill      int64      `json:"pay_till" `
	Receipt      int64      `json:"receipt" `
	ReceiptNum   int64      `json:"receipt_num" `
	AcNum        string     `json:"ac_num" `
	State        string     `json:"state" `
	Elapsed      float64    `json:"elapsed"`
	Total        float64    `json:"total"`
	ServerID     int64
}

type SaleItem struct {
	table       string    `name:"sales"`
	TransDate   time.Time `json:"trans_date"`
	ReceiptNum  int64     `json:"receipt_num"`
	OrderNum    int64     `json:"order_num"`
	TxnID       int64     `json:"txn_id"`
	HsCode      string    `json:"hs_code"`
	ItemCode    string    `json:"item_code"`
	ItemName    string    `json:"item_name"`
	Quantity    float64   `json:"quantity"`
	Cost        float64   `json:"cost"`
	Price       float64   `json:"price"`
	Discount    float64   `json:"discount"`
	Total       float64   `json:"total"`
	OnOffer     bool      `json:"on_offer"`
	Vat         float64   `json:"vat"`
	VatAlpha    string    `json:"vat_alpha"`
	State       string    `json:"state"`
	ReceiptItem string    `json:"receipt_item"`
}

// ProcessOrder
func (arg *SalesOrder) ProcessOrder(ctx context.Context) error {
	for _, itm := range arg.OrderItems {
		loc := products.Locations{StoreName: arg.Branch, StorageLoc: arg.StkLocation}

		err := loc.GetLocID(ctx)
		if err != nil {
			return err
		}

		prod, _ := products.GetByCode(itm.ItemCode, true)
		if prod.IsCombo && (prod.ComboItems != nil && len(prod.ComboItems) > 0) {
			for _, comboItm := range prod.ComboItems {
				bal := balances.TxnLog{Description: "sales orders"}

				bal.LocationID = loc.AutoID
				bal.ItemCode = comboItm.ItemCode
				bal.QtyOut = itm.Quantity * comboItm.Quantity
				bal.TxnID = fmt.Sprintf("%v-%v", itm.ReceiptItem, comboItm.ItemCode)

				if itm.State == "DELETED" || itm.State == "VOIDED" {
					if err := bal.RemoveBal(ctx); err != nil {
						return err
					}

					if err := bal.SaveBal(ctx); err != nil {
						return err
					}

					continue
				}
				if err := bal.LogBal(ctx); err != nil {
					return err
				}

				if err := bal.SaveBal(ctx); err != nil {
					return err
				}

			}
			continue
		}

		bal := balances.TxnLog{Description: "sales orders"}

		bal.LocationID = loc.AutoID
		bal.ItemCode = itm.ItemCode
		bal.QtyOut = itm.Quantity
		bal.TxnID = itm.ReceiptItem

		if itm.State == "DELETED" || itm.State == "VOIDED" {
			if err := bal.RemoveBal(ctx); err != nil {
				return err
			}

			if err := bal.SaveBal(ctx); err != nil {
				return err
			}

			continue
		}

		if err = bal.LogBal(ctx); err != nil {
			return err
		}
		if err := bal.SaveBal(ctx); err != nil {
			return err
		}
	}
	return nil
}
