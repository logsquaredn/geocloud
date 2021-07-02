package aggregator

import (
	"database/sql"
	"net/http"
	"path"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/logsquaredn/geocloud/shared/das"
	"github.com/rs/zerolog/log"
)

type s3Handler struct {
	das    *das.Das
	svc    *s3.S3
	bucket string
	prefix string
}

var _ http.Handler = (*s3Handler)(nil)

func (h *s3Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	prefix := path.Join(h.prefix, id)
	_, err = h.svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: &h.bucket,
		Prefix: &prefix,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = h.svc.GetObject(&s3.GetObjectInput{
		Bucket: &h.bucket,

	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
