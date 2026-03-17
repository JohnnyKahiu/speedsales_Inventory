package products

import (
	"log"
)

func GenProductsTables() error {
	err := createCodeTbls()
	if err != nil {
		log.Println("error.  failed to generate code_translator table    err =", err)
	}

	err = CreateVatsTable()
	if err != nil {
		log.Println("error, failed to generate vats table    err =", err)
	}

	err = CreateVatsDefaults()
	if err != nil {
		log.Println("error, failed to generate default vats    err =", err)
	}

	err = createSupplierTbl()
	if err != nil {
		log.Println("error, failed to generate supplier table    err =", err)
	}

	err = genLocationsTbl()
	if err != nil {
		log.Println("error.  failed to generate stock_locations table    err =", err)
	}

	err = createStockMasterTbl()
	if err != nil {
		log.Println("error, failed to generate stock_master table    err =", err)
	}

	err = createDeptTbl()
	if err != nil {
		log.Println("error, failed to generate department table    err =", err)
	}

	err = genDescriptionTbl()
	if err != nil {
		log.Println("error, failed to generate description table    err =", err)
	}

	return err
}
