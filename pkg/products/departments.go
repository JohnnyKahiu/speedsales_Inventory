package products

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/JohnnyKahiu/speedsales_inventory/database"
)

// Departments structure holding all products departments
type Departments struct {
	table          string `name:"departments" type:"table"`
	Code           int64  `json:"code" type:"field" sql:"SERIAL UNIQUE"`
	Name           string `json:"name" type:"field" sql:"VARCHAR NOT NULL"`
	SubDeptName    string `json:"sub_dept_name" type:"field" sql:"VARCHAR NOT NULL"`
	MinMargin      string `json:"min_margin" type:"field" sql:"VARCHAR NOT NULL DEFAULT '0.15'"`
	Label          string `json:"label"`
	IsMenuCategory bool   `json:"is_menu_category" type:"field" sql:"bool NOT NULL DEFAULT true"`
	composite      string `name:"departments_pkey" type:"constraint" sql:"PRIMARY KEY(name, sub_dept_name)"`
}

func createDeptTbl() error {
	var tblStruct Departments
	return database.CreateFromStruct(tblStruct)
}

func GetDepartments(onlyMenu bool) ([]Departments, map[string]map[string]int64, error) {
	var deps []Departments

	cond := ""
	if onlyMenu {
		cond = `WHERE is_menu_category = true`
	}
	sql := fmt.Sprintf(`
			SELECT 
				code
				, name
				, sub_dept_name
				, min_margin
				, is_menu_category
			FROM departments 
			%v
			ORDER BY code ASC`, cond)
	fmt.Println("sql =", sql)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rows, err := database.PgPool.Query(ctx, sql)
	if err != nil {
		return deps, nil, err
	}
	defer rows.Close()

	cartegories := make(map[string]map[string]int64)
	for rows.Next() {
		var dep Departments
		err = rows.Scan(&dep.Code, &dep.Name, &dep.SubDeptName, &dep.MinMargin, &dep.IsMenuCategory)
		if err != nil {
			return deps, cartegories, err
		}
		deps = append(deps, dep)

		if _, ok := cartegories[dep.Name]; !ok {
			cartegories[dep.Name] = make(map[string]int64)
		}

		cartegories[dep.Name][dep.SubDeptName] = dep.Code
	}

	fmt.Println("\n\t cartegories = ", cartegories)
	return deps, cartegories, nil
}

// CreateNew adds a new department
// Inserts into departments table
// returns an error if it fails
func (arg *Departments) CreateNew() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sql := `INSERT INTO departments(name, sub_dept_name, min_margin) 
			VALUES($1, $2, $3) 
			RETURNING code`

	rows, err := database.PgPool.Query(ctx, sql, arg.Name, arg.SubDeptName, arg.MinMargin)
	if err != nil {
		log.Println("error. failed to createNew() department    err =", err)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&arg.Code)
		if err != nil {
			return err
		}
	}

	return nil
}

// Update updates a department
// Updates the department in the database
// returns an error if it fails
func (arg *Departments) Update(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	sql := `UPDATE departments
			SET name = $1
				, sub_dept_name = $2
				, min_margin = $3
				, is_menu_category = $4
			WHERE code = $5`

	_, err := database.PgPool.Exec(ctx, sql, arg.Name, arg.SubDeptName, arg.MinMargin, arg.IsMenuCategory, arg.Code)
	if err != nil {
		log.Println("error. failed to update() department    err =", err)
		return err
	}

	return nil
}

// Delete removes a department
// Deletes from departments table in database
// returns an error if it fails
func (arg *Departments) Delete() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sql := `DELETE FROM departments WHERE code = $1`

	_, err := database.PgPool.Query(ctx, sql, arg.Code)
	if err != nil {
		log.Println("error. failed to delete() department    err =", err)
		return err
	}

	return nil
}

// SearchByName searches for departments by name
// Returns a slice of Departments that match the name
// Returns an error if the search fails
func SearchDeptByName(key string) ([]Departments, error) {
	var deps []Departments

	search := "%" + key + "%"

	sql := fmt.Sprintf(`SELECT 
				code
				, name
				, sub_dept_name
				, min_margin
			FROM departments 
			WHERE sub_dept_name ILIKE '%v' 
			ORDER BY code ASC`, search)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rows, err := database.PgPool.Query(ctx, sql)
	if err != nil {
		log.Println("error. failed to searchByName() department    err =", err)
		return deps, err
	}
	defer rows.Close()

	for rows.Next() {
		var dep Departments
		err = rows.Scan(&dep.Code, &dep.Name, &dep.SubDeptName, &dep.MinMargin)
		if err != nil {
			return deps, err
		}
		deps = append(deps, dep)
	}

	return deps, nil
}
