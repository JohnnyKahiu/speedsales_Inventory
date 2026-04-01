package database

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"
)

// GetTblSql generates the SQL for creating a table from a struct
func GenTablesFieldSQL(tblStruct any) string {
	sqlBody := ``

	val := reflect.ValueOf(tblStruct)
	for i := 0; i < val.Type().NumField(); i++ {
		if val.Type().Field(i).Tag.Get("type") == "field" && i < (val.Type().NumField()-1) {
			fieldName := val.Type().Field(i).Tag.Get("json") + ""
			sqlDef := val.Type().Field(i).Tag.Get("sql") + ""
			if sqlBody == "" {
				sqlBody += fieldName + " " + sqlDef + "\n\t"
			} else {
				sqlBody += ", " + fieldName + " " + sqlDef + "\n\t"
			}
		}

		fmt.Println(sqlBody)
	}

	return sqlBody
}

// CreateFromStruct generates tables from struct
func CreateFromStruct(tblStruct any) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sqlBody := ``
	sqlHead := ``

	// var tblStruct Property

	tblName := ""

	val := reflect.ValueOf(tblStruct)
	for i := 0; i < val.Type().NumField(); i++ {
		if val.Type().Field(i).Tag.Get("type") == "field" && i < (val.Type().NumField()-1) {
			fieldName := val.Type().Field(i).Tag.Get("json") + ""
			sqlDef := val.Type().Field(i).Tag.Get("sql") + ""
			if sqlBody == "" {
				sqlBody += fieldName + " " + sqlDef + "\n\t"
			} else {
				sqlBody += ", " + fieldName + " " + sqlDef + "\n\t"
			}
		} else if val.Type().Field(i).Tag.Get("type") == "table" {
			tblName = val.Type().Field(i).Tag.Get("name")
			fmt.Printf("\n\t %v \n", tblName)
		} else if val.Type().Field(i).Tag.Get("type") == "constraint" {
			fieldName := val.Type().Field(i).Tag.Get("name") + ""
			sqlDef := val.Type().Field(i).Tag.Get("sql") + ""

			sqlBody += ", CONSTRAINT " + fieldName + " " + sqlDef + "\n\t"
		}
	}

	sqlHead = fmt.Sprintf("CREATE TABLE IF NOT EXISTS %v ( ", tblName)
	sql := sqlHead + sqlBody + ");"

	fmt.Println("sql =", sql)
	// run sql transaction
	_, err := PgPool.Exec(ctx, sql)
	if err != nil {
		log.Printf("\nerror %s table\n \t%v", tblName, err.Error())
		return err
	}

	// Add non existing columns
	for i := 0; i < val.Type().NumField(); i++ {
		if val.Type().Field(i).Tag.Get("type") == "field" {
			fieldName := val.Type().Field(i).Tag.Get("json") + ""
			sqlDef := val.Type().Field(i).Tag.Get("sql") + " ;"

			if !strings.Contains(sqlDef, "PRIMARY KEY") && fieldName != "" {
				sqlAlter := fmt.Sprintf("ALTER TABLE IF EXISTS %v ADD IF NOT EXISTS %v %v", tblName, fieldName, sqlDef)
				// fmt.Println("\t", sqlAlter)
				_, err := PgPool.Exec(ctx, sqlAlter)
				if err != nil {
					fmt.Printf("\nerror Altering %v table\n \t%v", tblName, err.Error())
				}
			}
		}

		if val.Type().Field(i).Tag.Get("type") == "constraint" {
			fieldName := val.Type().Field(i).Tag.Get("name") + ""
			sqlDef := val.Type().Field(i).Tag.Get("sql") + " ;"
			sqlConst := fmt.Sprintf("ALTER TABLE IF EXISTS %v ADD CONSTRAINT %v %v", tblName, fieldName, sqlDef)
			fmt.Println(sqlConst)

			_, err := PgPool.Exec(ctx, sqlConst)
			if err != nil {
				fmt.Printf("\nerror Altering %v table\n \t%v", tblName, err.Error())
			}
		}

	}

	return nil
}

func CreateFromXStruct(tblStruct any) error {
	sqlBody := ``
	sqlHead := ``

	// var tblStruct Property

	tblName := ""

	val := reflect.ValueOf(tblStruct)
	for i := 0; i < val.Type().NumField(); i++ {
		if val.Type().Field(i).Tag.Get("type") == "field" && i < (val.Type().NumField()-1) {
			fieldName := val.Type().Field(i).Tag.Get("json") + ""
			sqlDef := val.Type().Field(i).Tag.Get("sql") + ""
			if sqlBody == "" {
				sqlBody += fieldName + " " + sqlDef + "\n\t"
			} else {
				sqlBody += ", " + fieldName + " " + sqlDef + "\n\t"
			}
		} else if val.Type().Field(i).Tag.Get("type") == "table" {
			tblName = val.Type().Field(i).Tag.Get("name")
			fmt.Printf("\n\t%v \t", tblName)
		} else if val.Type().Field(i).Tag.Get("type") == "constraint" {
			fieldName := val.Type().Field(i).Tag.Get("name") + ""
			sqlDef := val.Type().Field(i).Tag.Get("sql") + ""

			sqlBody += ", CONSTRAINT " + fieldName + " " + sqlDef + "\n\t"
		} else if val.Type().Field(i).Tag.Get("type") == "tbl_struct" {
			fmt.Printf("\n\t%v \t", val.Type().Field(i).Type)
			sqlStr := GenTablesFieldSQL(reflect.New(val.Type().Field(i).Type).Elem().Interface())
			if sqlBody == "" {
				sqlBody += sqlStr
			} else {
				sqlBody += ", " + sqlStr
			}
		}
	}
	sqlHead = fmt.Sprintf("CREATE TABLE IF NOT EXISTS %v ( \n\t", tblName)
	sql := sqlHead + sqlBody + ");"
	fmt.Println("sql =", sql)

	_, err := PgPool.Exec(context.Background(), sql)
	if err != nil {
		log.Printf("\nerror creating %v table\n \t%v", tblName)
		return err
	}
	return nil
}

// InsertFromStruct
func InsertFromStruct(tblStruct any) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	val := reflect.ValueOf(tblStruct)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	tblName := ""
	var columns []string
	var values []any
	var placeholders []string

	for i := 0; i < val.Type().NumField(); i++ {
		field := val.Type().Field(i)
		tagType := field.Tag.Get("type")
		switch tagType {
		case "table":
			tblName = field.Tag.Get("name")
		case "field":
			colName := field.Tag.Get("json")
			if colName != "" {
				columns = append(columns, colName)
				values = append(values, val.Field(i).Interface())
				placeholders = append(placeholders, fmt.Sprintf("$%d", len(values)))
			}
		}
	}

	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tblName, strings.Join(columns, ", "), strings.Join(placeholders, ", "))
	fmt.Println(sql)
	_, err := PgPool.Exec(ctx, sql, values...)
	if err != nil {
		log.Printf("\nerror inserting into %v table\n \t%v", tblName, err.Error())
		return err
	}

	return nil
}
