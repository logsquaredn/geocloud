package proxy

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/mail"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/logsquaredn/rototiller"
	"github.com/logsquaredn/rototiller/pkg/api"
	"github.com/logsquaredn/rototiller/pkg/service"
	files "github.com/swaggo/files"
	swagger "github.com/swaggo/gin-swagger"
)

func NewHandler(ctx context.Context, proxyAddr, key string) (http.Handler, error) {
	var (
		_              = rototiller.LoggerFrom(ctx)
		router         = gin.Default()
		swaggerHandler = swagger.WrapHandler(files.Handler)
		tokenParser    = jwt.NewParser()
	)

	u, err := url.Parse(proxyAddr)
	if err != nil {
		return nil, err
	}

	var (
		reverseProxy   = httputil.NewSingleHostReverseProxy(u)
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
			if ctx.Param("any") == "" {
				ctx.Redirect(http.StatusFound, "/swagger/v1/index.html")
			} else {
				swaggerHandler(ctx)
			}
		})
	}

	token := router.Group("/api/v1/token")
	{
		// @Summary      Create a token
		// @Description  <b><u>Create a token</u></b>
		// @Tags         Token
		// @Accept       application/json
		// @Produce      application/json
		// @Success      201                {object}  rototiller.Auth
		// @Failure      400                {object}  rototiller.Error
		// @Failure      500                {object}  rototiller.Error
		// @Router       /api/v1/token [post].
		token.POST("", func(ctx *gin.Context) {
			claims := &api.Claims{}
			if err := ctx.ShouldBindJSON(claims); err != nil {
				ctx.JSON(http.StatusBadRequest, api.NewErr(err))
				return
			}

			if _, err = mail.ParseAddress(claims.GetEmail()); err != nil {
				ctx.JSON(http.StatusBadRequest, api.NewErr(err))
				return
			}

			var (
				now   = time.Now()
				exp   = now.Add(time.Hour * 24 * 7 * 4)
				token = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
					Subject:   claims.GetEmail(),
					NotBefore: jwt.NewNumericDate(now),
					IssuedAt:  jwt.NewNumericDate(now),
					ExpiresAt: jwt.NewNumericDate(exp),
					ID:        uuid.NewString(),
				})
			)

			apiKey, err := token.SignedString([]byte(key))
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, api.NewErr(err))
				return
			}

			ctx.JSON(http.StatusCreated, &api.Auth{
				ApiKey: apiKey,
			})
		})
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
			ctx.Header(service.OwnerIDHeader, claims.Subject)
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
