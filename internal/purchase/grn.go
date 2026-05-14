package purchase

// errorResp is a convenience builder for error response maps.
func errorResp(m map[string]interface{}, msg string) map[string]interface{} {
	m["response"] = "error"
	m["message"] = msg
	return m
}
