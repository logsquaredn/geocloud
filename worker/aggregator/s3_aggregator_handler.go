package aggregator

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/logsquaredn/geocloud"
	"github.com/rs/zerolog/log"
)

var _ http.Handler = (*S3Aggregrator)(nil)

type message struct {
	id string
}

func (m *message) ID() string {
	return m.id
}

var _ geocloud.Message = (*message)(nil)

func (h *S3Aggregrator) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Debug().Fields(f{ "runner": runner }).Msgf("%s %s?%s", r.Method, r.URL.Path, r.URL.RawQuery)

	id := r.URL.Query().Get("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("no id specified"))
		return
	}

	err := h.Aggregate(context.Background(), &message{ id })
	if err == sql.ErrNoRows {
		w.Write([]byte(err.Error()))
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
}
