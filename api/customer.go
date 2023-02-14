package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var NamespaceHeader = "X-Namespace"

// getNamespaceFromHeader gets the namespace from the given http.Header.
func (a *Handler) getNamespaceFromHeader(header http.Header) (string, error) {
	return header.Get(NamespaceHeader), nil
}

// getNamespaceFromContext returns the namespace.
func (a *Handler) getNamespaceFromContext(ctx *gin.Context) (string, error) {
	namespace, err := a.getNamespaceFromHeader(ctx.Request.Header)
	if err != nil {
		return "", err
	}

	return namespace, nil
}
