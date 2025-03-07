package products

// CodeTranslator structure holds translation of master and linked codes
type CodeTranslator struct {
	table       string  `name:"code_translator" type:"table"`
	AutoID      int64   `json:"auto_id" name:"auto_id" type:"field" sql:"BIGSERIAL PRIMARY KEY"`
	MasterCode  string  `json:"master_code" name:"master_code" type:"field" sql:"VARCHAR(30) NOT NULL"`
	LinkCode    string  `json:"link_code" name:"link_code" type:"field" sql:"VARCHAR(30) NOT NULL"`
	PkgQty      float64 `json:"pkg_qty" name:"pkg_qty" type:"field" sql:"INT NOT NULL DEFAULT '1'"`
	Discount    float64 `json:"discount" name:"discount" type:"field" sql:"FLOAT NOT NULL DEFAULT '0'"`
	constraint  int     `name:"code_translator_pkey" type:"constraint" sql:"PRIMARY KEY(link_code, master_code)"`
	constraint2 string  `name:"code_translator_link_code_key" type:"constraint" sql:"UNIQUE(link_code)"`
}

// GetAllLinks fetches a list of all codes linked to a master
// searches all codes and returns all where master code in arg method is same as in productsMaster
// returns a slice of CodeTranslator and an error if exists
func (arg CodeTranslator) GetAllLinks() ([]CodeTranslator, error) {
	results := []CodeTranslator{}
	codes := ProdMaster.Codes
	for _, code := range codes {
		if code.MasterCode == arg.MasterCode {
			results = append(results, code)
		}
	}
	return results, nil
}
