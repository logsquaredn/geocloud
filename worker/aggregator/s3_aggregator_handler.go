package aggregator

import (
	"database/sql"
	"net/http"

	"github.com/rs/zerolog/log"
)

var _ http.Handler = (*S3Aggregrator)(nil)

func (h *S3Aggregrator) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Debug().Fields(f{ "runner": runner }).Msgf("%s %s?%s", r.Method, r.URL.Path, r.URL.RawQuery)

	id := r.URL.Query().Get("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
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
