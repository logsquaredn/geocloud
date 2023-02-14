package proxy

import (
	"fmt"
	"net/http"
	"net/mail"
	"net/smtp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/logsquaredn/rototiller"
	"github.com/logsquaredn/rototiller/pb"
)

// @Summary      Create an API key
// @Description  <b><u>Create an API key</u></b>
// @Tags         API-Key
// @Accept       application/json
// @Produce      application/json
// @Param        request  body      rototiller.Claims  true  "user info"
// @Success      200      {object}  rototiller.Auth
// @Success      201
// @Failure      400  {object}  rototiller.Error
// @Failure      500  {object}  rototiller.Error
// @Router       /api/v1/api-key [post].
func (h *Handler) createApiKey(ctx *gin.Context) {
	var (
		logr   = rototiller.NewLogger()
		claims = &rototiller.Claims{}
	)

	if err := ctx.ShouldBindJSON(claims); err != nil {
		logr.Error(err, "failed bind request body")
		ctx.JSON(http.StatusBadRequest, pb.NewErr(err))
		return
	}

	if _, err := mail.ParseAddress(claims.GetEmail()); err != nil {
		logr.Error(err, "failed to parse email address")
		ctx.JSON(http.StatusBadRequest, pb.NewErr(fmt.Errorf("failed to parse email address")))
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
		ctx.JSON(http.StatusInternalServerError, pb.NewErr(err))
		return
	}

	if h.SMTPURL != nil {
		err = h.sendEmail(claims.Email, apiKey)
		if err != nil {
			logr.Error(err, "failed to send email containing api-key")
			ctx.JSON(http.StatusInternalServerError, pb.NewErr(fmt.Errorf("failed to send email containing API key")))
			return
		}

		ctx.Status(http.StatusCreated)
	} else {
		ctx.JSON(http.StatusOK, &pb.Auth{
			ApiKey: apiKey,
		})
	}
}

func (h *Handler) sendEmail(email string, apiKey string) error {
	var (
		to      = []string{email}
		message = []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: Rototiller API Key\r\n%s", h.From, to, apiKey))
	)

	return smtp.SendMail(h.SMTPURL.Host, h.SMTPAuth, h.From, to, message)
}
