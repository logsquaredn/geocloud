package worker

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/frantjc/go-js"
	"github.com/logsquaredn/geocloud"
	"github.com/logsquaredn/geocloud/datastore"
	"github.com/logsquaredn/geocloud/objectstore"
	"github.com/rs/zerolog/log"
)

type Worker struct {
	ds      *datastore.Postgres
	os      *objectstore.S3
	workdir string
}

func New(opts *Opts) (*Worker, error) {
	return &Worker{
		ds:      opts.Datastore,
		os:      opts.Objectstore,
		workdir: opts.WorkDir,
	}, nil
}

func (o *Worker) Send(m geocloud.Message) error {
	k, v := "id", m.GetID()
	log.Info().Str(k, v).Msg("processing message")

	log.Trace().Str(k, v).Msg("getting job from datastore")
	j, err := o.ds.GetJob(m)
	if err != nil {
		return err
	}

	switch j.Status {
	case geocloud.JobStatusComplete, geocloud.JobStatusInProgress:
		return nil
	}

	stderr := new(bytes.Buffer)
	defer func() {
		j.EndTime = time.Now()
		jobErr := stderr.Bytes()
		switch {
		case len(jobErr) > 0:
			j.Error = string(jobErr)
			j.Status = geocloud.JobStatusError
		case err != nil:
			j.Error = err.Error()
			j.Status = geocloud.JobStatusError
		default:
			j.Status = geocloud.JobStatusComplete
		}
		log.Err(
			js.Ternary(
				j.Error != "",
				fmt.Errorf(j.Error),
				nil,
			),
		).Str(k, v).Msgf("job finished with status '%s'", j.Status.Status())
		if _, err = o.ds.UpdateJob(j); err != nil {
			log.Err(err).Msgf("updating finished job '%s'", j.ID)
		}
	}()

	go func() {
		ist, err := o.ds.GetStorage(geocloud.NewMessage(j.InputID))
		if err == nil {
			_, err = o.ds.UpdateStorage(ist)
			if err != nil {
				log.Err(err).Msg("updating input storage")
			}
		} else {
			log.Err(err).Msg("getting input storage")
		}
	}()

	j.Status = geocloud.JobStatusInProgress
	log.Trace().Str(k, v).Msgf("setting job to %s", j.Status.Status())
	j, err = o.ds.UpdateJob(j)
	if err != nil {
		return err
	}

	log.Trace().Str(k, v).Msg("getting task for job from datastore")
	t, err := o.ds.GetTaskByJobID(m)
	if err != nil {
		return err
	}

	log.Trace().Str(k, v).Msg("creating input volume")
	invol, err := o.inputVolume(m)
	if err != nil {
		return err
	}
	defer os.RemoveAll(o.jobDir(m))

	log.Trace().Str(k, v).Msg("getting input")
	input, err := o.os.GetObject(geocloud.NewMessage(j.InputID))
	if err != nil {
		return err
	}

	log.Debug().Str(k, v).Msg("downloading input")
	if err = input.Download(o.inputVolumePath(j)); err != nil {
		return err
	}

	log.Trace().Str(k, v).Msg("creating output volume")
	outvol, err := o.outputVolume(m)
	if err != nil {
		return err
	}

	var (
		filename string
		_        = invol.Walk(func(_ string, f geocloud.File, e error) error {
			if e != nil {
				return e
			}
			filename = f.Name()
			// we only expect 1 input, so use the first one we find and end the Walk
			return fmt.Errorf("found")
		})
	)

	if filename == "" {
		return fmt.Errorf("no input found")
	}

	args := append(
		[]string{
			filepath.Join(o.inputVolumePath(j), filename),
			o.outputVolumePath(j),
		},
		j.Args...,
	)
	// TODO pass args as env
	task := exec.Command(t.Type.Name(), args...) //nolint:gosec
	// don't let tasks see potentially sensitive environment variables
	task.Env = js.Filter(os.Environ(), func(e string, _ int, _ []string) bool {
		return !(strings.HasPrefix(e, "GEOCLOUD_") || strings.HasPrefix(e, "AWS_"))
	})
	task.Stdin = os.Stdin
	task.Stdout = os.Stdout
	task.Stderr = stderr

	log.Info().Str(k, v).Msgf("running task %s", task.Path)
	if err := task.Run(); err != nil {
		return err
	}

	log.Trace().Str(k, v).Msg("creating output storage")
	ost, err := o.ds.CreateStorage(&geocloud.Storage{
		CustomerID: j.CustomerID,
	})
	if err != nil {
		return err
	}
	j.OutputID = ost.ID

	log.Debug().Str(k, v).Msg("uploading output")
	return o.os.PutObject(geocloud.NewMessage(j.OutputID), outvol)
}

func (o *Worker) inputVolumePath(m geocloud.Message) string {
	return filepath.Join(o.jobDir(m), "input")
}

func (o *Worker) outputVolumePath(m geocloud.Message) string {
	return filepath.Join(o.jobDir(m), "output")
}

func (o *Worker) inputVolume(m geocloud.Message) (geocloud.Volume, error) {
	return geocloud.NewDirVolume(o.inputVolumePath(m))
}

func (o *Worker) outputVolume(m geocloud.Message) (geocloud.Volume, error) {
	return geocloud.NewDirVolume(o.outputVolumePath(m))
}

func (o *Worker) jobDir(m geocloud.Message) string {
	return filepath.Join(o.jobsDir(), m.GetID())
}

func (o *Worker) jobsDir() string {
	return filepath.Join(o.workdir, "jobs")
}
