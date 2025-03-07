package products

import "log"

func GenProductsTables() {
	// err := ge

	err := genLocationsTbl()
	if err != nil {
		log.Println("error.  failed to generate stock_locations table    err =", err)
	}

}
