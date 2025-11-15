package products

import (
	"context"
	"fmt"
	"log"

	"github.com/JohnnyKahiu/speedsales_inventory/database"
)

// Recipe Holds information about a stock item
type Recipe struct {
	table       string  `name:"recipe" type:"table"`
	ProdCode    string  `json:"prod_code" name:"prod_code" type:"field" sql:"VARCHAR(50) NOT NULL"`
	ItemCode    string  `json:"item_code" name:"item_code" type:"field" sql:"VARCHAR(50) NOT NULL"`
	ItemName    string  `json:"item_name" name:"item_name" type:"field" sql:"VARCHAR(50) NOT NULL"`
	Measurement string  `json:"measurement" name:"measurement" type:"field" sql:"VARCHAR(50) NOT NULL"`
	Amount      float64 `json:"amount" name:"amount" type:"field" sql:"FLOAT NOT NULL DEFAULT '1'"`
	Label       string  `json:"label"`
}

// FetchRecipeItems
func FetchRecipeItems(prodCode string) ([]Recipe, error) {
	return ProdMaster.Recipe[prodCode], nil
}

// AddRecipeItem function adds new items to product's recipe
func AddRecipeItem(arg Recipe) error {
	if ProdMaster.Recipe == nil {
		ProdMaster.Recipe = make(map[string][]Recipe)
	}

	for i, recipeItem := range ProdMaster.Recipe[arg.ProdCode] {
		if recipeItem.ItemCode == arg.ItemCode {
			ProdMaster.Recipe[arg.ProdCode][i].Amount = arg.Amount
			return nil
		}
	}

	ProdMaster.Recipe[arg.ProdCode] = append(ProdMaster.Recipe[arg.ProdCode], arg)

	err := arg.AddToDB()
	if err != nil {
		return err
	}

	err = adjustCostFromRecipe(arg.ProdCode)
	if err != nil {
		return err
	}

	return ProdMaster.Pickle()
}

// AddToDB adds recipe items to database
func (arg Recipe) AddToDB() error {
	sql := `INSERT INTO recipe(prod_code, item_code, item_name, measurement, amount)
			VALUES($1, $2, $3, $4, $5) 
			ON CONFLICT ON CONSTRAINT recipe_item_unique
				DO UPDATE 
				SET 
					amount = EXCLUDED.amount
		`

	_, err := database.PgPool.Exec(context.Background(), sql, arg.ProdCode, arg.ItemCode, arg.ItemName, arg.Measurement, arg.Amount)
	if err != nil {
		log.Println("error.  recipe AddToB()    err =", err)
		return err
	}

	fmt.Println("successfully added recipe items to DB")
	return nil
}

// DBFetch gets recipe items from database
func (arg Recipe) DBFetch() ([]Recipe, error) {
	sql := `SELECT 
				prod_code
				, item_code
				, item_name
				, measurement
				, amount
			FROM recipe WHERE prod_code = $1
		`

	rows, err := database.PgPool.Query(context.Background(), sql, arg.ProdCode)
	if err != nil {
		log.Println("error.  recipe DBFetch()    err =", err)
		return nil, err
	}
	defer rows.Close()

	vals := []Recipe{}
	for rows.Next() {
		var r Recipe
		err := rows.Scan(&r.ProdCode, &r.ItemCode, &r.ItemName, &r.Measurement, &r.Amount)
		if err != nil {
			log.Println("error.  recipe DBFetch()    err =", err)
			return nil, err
		}
		vals = append(vals, r)
	}

	fmt.Println("successfully added recipe items to DB")
	return vals, nil
}

// delRecipe function deletes an existing item from recipe
func delRecipe(prodCode, itemCode string) error {
	var correct []Recipe
	for _, item := range ProdMaster.Recipe[prodCode] {
		if item.ItemCode != itemCode {
			correct = append(correct, item)
		}
	}

	ProdMaster.Recipe[prodCode] = correct
	return nil
}

// delRecipeDB function deletes an existing item from recipe
func delRecipeDB(prodCode, itemCode string) error {
	sql := `DELETE FROM recipe WHERE prod_code = $1 AND item_code = $2`

	_, err := database.PgPool.Exec(context.Background(), sql, prodCode, itemCode)
	if err != nil {
		return err
	}

	return nil
}

// adjustCostFromRecipe function calculates new item cost from recipe items
func adjustCostFromRecipe(itemCode string) error {
	BatchCost := 0.0
	for _, item := range ProdMaster.Recipe[itemCode] {
		itm := ProdMaster.ProductDB[item.ItemCode]

		BatchCost += itm.ItemCost * item.Amount
	}

	newCost := BatchCost / ProdMaster.ProductDB[itemCode].UnitsPerRecipe

	item := ProdMaster.ProductDB[itemCode]
	item.ItemCost = newCost

	ProdMaster.ProductDB[itemCode] = item

	return nil
}

// adjustCostFromRecipe function calculates new item cost from recipe items
func adjustCostFromRecipeDB(itemCode string) error {
	sql := `UPDATE stock_master 
			SET item_cost = (SELECT 
								SUM((stm.item_cost * r.amount) / (SELECT coalesce(units_per_recipe, 1) FROM stock_master WHERE item_code = $1 ))  
							FROM recipe r LEFT JOIN stock_master stm 
								ON stm.item_code = r.item_code 
							WHERE r.prod_code = $1)
			WHERE item_code = $1`

	_, err := database.PgPool.Exec(context.Background(), sql, itemCode)
	if err != nil {
		return err
	}
	fmt.Println("\t adjusting cost from recipe")
	return nil
}
