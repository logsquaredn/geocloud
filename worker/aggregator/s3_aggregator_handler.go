package aggregator

import (
	"database/sql"
	"net/http"
	"os"
	"path"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/rs/zerolog/log"
)

var _ http.Handler = (*S3Aggregrator)(nil)

func (h *S3Aggregrator) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Debug().Fields(f{ "runner":runner }).Msgf("%s %s?%s", r.Method, r.URL.Path, r.URL.RawQuery)

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

	input := path.Join(h.prefix, id, "input")
	output, err := h.svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: &h.bucket,
		Prefix: &input,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if len(output.Contents) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	iter := s3manager.DownloadObjectsIterator{}
	for _, content := range output.Contents {
		file, err := os.Open(path.Join("/tmp", id, "input", path.Base(*content.Key)))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer file.Close()
		defer os.RemoveAll(path.Dir(file.Name()))

		iter.Objects = append(iter.Objects, s3manager.BatchDownloadObject{
			Object: &s3.GetObjectInput{
				Bucket: &h.bucket,
				Key: content.Key,
			},
			Writer: file,
		})
	}

	err = h.dwnldr.DownloadWithIterator(aws.BackgroundContext(), &iter)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
