package products

type Department struct {
	Code        int64  `json:"dept_code"`
	Name        string `json:"dept_name"`
	SubDeptName string `json:"sub_dept_name"`
}
