package aggregator

import (
	"database/sql"
	"net/http"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/logsquaredn/geocloud/shared/das"
	"github.com/rs/zerolog/log"
)

type s3AggregatorHandler struct {
	das *das.Das
	svc *s3.S3
}

var _ http.Handler = (*s3AggregatorHandler)(nil)

func (h *s3AggregatorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Debug().Fields(f{ "runner":runner }).Msgf("%s %s?%s", r.Method, r.URL.Path, r.URL.RawQuery)

	id := r.URL.Query().Get("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
	}

	_, err := h.das.GetJobTypeByJobId(id)
	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
