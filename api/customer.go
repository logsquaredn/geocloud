package api

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/logsquaredn/geocloud"
	"github.com/rs/zerolog/log"
)

func (a *API) middleware(ctx *gin.Context) {
	apiKey := getCustomerID(ctx)
	if _, err := a.ds.GetCustomer(geocloud.NewMessage(apiKey)); err != nil {
		if err == sql.ErrNoRows {
			log.Err(err).Msgf("query parameter '%s', header '%s' or cookie '%s' must be a valid API Key", apiKeyQueryParam, apiKeyHeader, apiKeyCookie)
			ctx.AbortWithStatusJSON(http.StatusForbidden, &geocloud.ErrorResponse{Error: fmt.Sprintf("header '%s', header '%s' cookie '%s' must be a valid API Key", apiKeyQueryParam, apiKeyHeader, apiKeyCookie)})
			return
		}

		log.Err(err).Msgf("failed to get user information")
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, &geocloud.ErrorResponse{Error: "failed to get user information"})
		return
	}

	ctx.Next()
}
