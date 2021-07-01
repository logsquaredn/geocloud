package aggregator

import (
	"database/sql"
	"net/http"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/rs/zerolog/log"
)

type s3AggregatorHandler struct {
	db  *sql.DB
	svc *s3.S3
}

var _ http.Handler = (*s3AggregatorHandler)(nil)

func (h *s3AggregatorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Debug().Fields(f{ "runner":runner }).Msgf("%s %s?%s", r.Method, r.URL.Path, r.URL.RawQuery)

	id := r.URL.Query().Get("id")
	if id == "" {
		w.WriteHeader(400)
	}
}
