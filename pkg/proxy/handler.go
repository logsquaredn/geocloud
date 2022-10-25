package proxy

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/logsquaredn/rototiller"
	"github.com/logsquaredn/rototiller/pkg/api"
	"github.com/logsquaredn/rototiller/pkg/service"
	files "github.com/swaggo/files"
	swagger "github.com/swaggo/gin-swagger"
)

type Handler struct {
	Key string
}

func NewHandler(ctx context.Context, proxyAddr, key string) (http.Handler, error) {
	var (
		_              = rototiller.LoggerFrom(ctx)
		router         = gin.Default()
		swaggerHandler = swagger.WrapHandler(files.Handler)
		tokenParser    = jwt.NewParser()
		h              = &Handler{key}
	)

	u, err := url.Parse(proxyAddr)
	if err != nil {
		return nil, err
	}

	var (
		reverseProxy = httputil.NewSingleHostReverseProxy(u)
	)

	router.GET("/healthz", func(ctx *gin.Context) {
		ctx.Data(http.StatusOK, "application/text", []byte("ok\n"))
	})
	router.GET("/readyz", func(ctx *gin.Context) {
		ctx.Data(http.StatusOK, "application/text", []byte("ok\n"))
	})

	swagger := router.Group("/swagger/v1")
	{
		swagger.GET("/*any", func(ctx *gin.Context) {
			if ctx.Param("any") == "/" || ctx.Param("any") == "" {
				ctx.Redirect(http.StatusFound, "/swagger/v1/index.html")
			} else {
				swaggerHandler(ctx)
			}
		})
	}

	apiKey := router.Group("/api/v1/api-key")
	{
		apiKey.POST("", h.createApiKey)
	}

	router.NoRoute(func(ctx *gin.Context) {
		rawToken := ctx.GetHeader("Authorization")
		if rawToken == "" {
			ctx.JSON(http.StatusUnauthorized, &api.Error{
				Message: "API key required",
			})
			return
		}
		ctx.Request.Header.Del("Authorization")

		token, err := tokenParser.ParseWithClaims(rawToken, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}

			return []byte(key), nil
		})
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, api.NewErr(err))
			return
		}

		if err = token.Claims.Valid(); err != nil {
			ctx.JSON(http.StatusForbidden, api.NewErr(err))
			return
		}

		if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
			ctx.Request.Header.Set(service.OwnerIDHeader, claims.Subject)
		} else {
			ctx.JSON(http.StatusForbidden, &api.Error{
				Message: "API key invalid",
			})
			return
		}

		reverseProxy.ServeHTTP(ctx.Writer, ctx.Request)
	})

	return router, nil
}
