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

func buildJobArgs(ctx *gin.Context, taskParams []string) []string {
	jobArgs := make([]string, len(taskParams))
	for i, param := range taskParams {
		jobArgs[i] = ctx.Query(param)
	}
	return jobArgs
}

func isJSON(b []byte) bool {
	var js map[string]interface{}
	return json.Unmarshal(b, &js) == nil
}

func isZIP(b []byte) bool {
	return http.DetectContentType(b) == "application/zip"
}
