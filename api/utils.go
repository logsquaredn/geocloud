package api

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	apiKeyQueryParam = "api-key"
	apiKeyHeader     = "X-API-Key"
	apiKeyCookie     = apiKeyHeader
)

var getCustomerID = getAPIKey

func getAPIKey(ctx *gin.Context) string {
	apiKey := ctx.Query(apiKeyQueryParam)
	if apiKey == "" {
		apiKey = ctx.GetHeader(apiKeyHeader)
		if apiKey == "" {
			apiKey, _ = ctx.Cookie(apiKeyCookie)
		}
	}
	return apiKey
}

func isJSON(jsBytes []byte) bool {
	var js map[string]interface{}
	return json.Unmarshal(jsBytes, &js) == nil
}

func isZIP(zipBytes []byte) bool {
	return http.DetectContentType(zipBytes) == "application/zip"
}

