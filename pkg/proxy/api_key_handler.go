package proxy

import (
	"net/http"
	"net/mail"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/logsquaredn/rototiller"
	"github.com/logsquaredn/rototiller/pkg/api"
)

// @Summary      Create an API key
// @Description  <b><u>Create an API key</u></b>
// @Tags         API-Key
// @Accept       application/json
// @Produce      application/json
// @Param        request  body      rototiller.Claims  true  "user info"
// @Success      201      {object}  rototiller.Auth
// @Failure      400      {object}  rototiller.Error
// @Failure      500      {object}  rototiller.Error
// @Router       /api/v1/api-key [post].
func (h *Handler) createApiKey(ctx *gin.Context) {
	claims := &rototiller.Claims{}
	if err := ctx.ShouldBindJSON(claims); err != nil {
		ctx.JSON(http.StatusBadRequest, api.NewErr(err))
		return
	}

	if _, err := mail.ParseAddress(claims.GetEmail()); err != nil {
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

	apiKey, err := token.SignedString([]byte(h.Key))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, api.NewErr(err))
		return
	}

	ctx.JSON(http.StatusCreated, &api.Auth{
		ApiKey: apiKey,
	})
}
