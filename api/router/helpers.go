package router

import "encoding/json"

func isJSON(jsBytes []byte) bool {
	var js map[string]interface{}
	return json.Unmarshal(jsBytes, &js) == nil
}
