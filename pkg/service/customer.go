package service

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/logsquaredn/rototiller/pkg/api"
)

var (
	OwnerIDHeader   = "X-Owner-ID"
	errUnauthorized = api.NewErr(
		errors.New(OwnerIDHeader+" header must exist"),
		http.StatusBadRequest,
	)
)

// getOwnerIDFromHeader gets the owner ID from the given http.Header.
func (a *Handler) getOwnerIDFromHeader(header http.Header) (string, error) {
	ownerID := header.Get(OwnerIDHeader)
	if ownerID == "" {
		return "", errUnauthorized
	}
	return ownerID, nil
}

// getOwnerIDFromContext returns the owner ID.
func (a *Handler) getOwnerIDFromContext(ctx *gin.Context) (string, error) {
	ownerID, err := a.getOwnerIDFromHeader(ctx.Request.Header)
	if err != nil {
		return "", err
	}
	return ownerID, nil
}
