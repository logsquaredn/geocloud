package runtime

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/logsquaredn/geocloud"
	"github.com/rs/zerolog/log"
)

type osRuntime struct {
	workdir string
	ds      geocloud.Datastore
	os      geocloud.Objectstore
}

func NewOS(opts *OSRuntimeOpts) (*osRuntime, error) {
	return &osRuntime{
		ds:      opts.Datastore,
		os:      opts.Objectstore,
		workdir: opts.WorkDir,
	}, nil
}

func (o *osRuntime) Send(m geocloud.Message) error {
	k, v := "id", m.GetID()
	log.Info().Str(k, v).Msg("processing message")

	log.Trace().Str(k, v).Msg("getting job from datastore")
	j, err := o.ds.GetJob(m)
	if err != nil {
		return err
	}

	switch j.Status {
	case geocloud.Complete, geocloud.InProgress:
		return nil
	}

	stderr := new(bytes.Buffer)
	defer func() {
		j.EndTime = time.Now()
		jobErr := stderr.Bytes()
		if len(jobErr) > 0 {
			j.Err = fmt.Errorf("%s", jobErr)
			j.Status = geocloud.Error
		} else if err != nil {
			j.Err = err
			j.Status = geocloud.Error
		} else {
			j.Status = geocloud.Complete
		}
		log.Err(j.Err).Str(k, v).Msgf("job finished with status %s", j.Status.Status())
		o.ds.UpdateJob(j)
	}()

	j.Status = geocloud.InProgress
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
	invol, err := o.involume(m)
	if err != nil {
		return err
	}
	defer os.RemoveAll(o.jobdir(m))

	log.Trace().Str(k, v).Msg("getting input")
	input, err := o.os.GetInput(m)
	if err != nil {
		return err
	}

	log.Debug().Str(k, v).Msg("downloading input")
	if err = input.Download(invol.path); err != nil {
		return err
	}

	log.Trace().Str(k, v).Msg("creating output volume")
	outvol, err := o.outvolume(m)
	if err != nil {
		return err
	}

	var filename string
	invol.Walk(func(_ string, f geocloud.File, e error) error {
		if e != nil {
			return e
		}
		filename = f.Name()
		return fmt.Errorf("found") // we only expect 1 input, so use the first one we find
	})

	if filename == "" {
		return fmt.Errorf("no input found")
	}

	ex, err := os.Executable()
	if err != nil {
		return err
	}

	name, err := exec.LookPath(t.Type.Name())
	if err != nil {
		path := filepath.Join(
			filepath.Dir(ex),
			t.Type.Name(),
			"task",
		)
		fi, err := os.Stat(path)
		if err != nil {
			return err
		} else if fi.IsDir() {
			return fmt.Errorf("unable to find executable for task type %s", t.Type.Name())
		}

		name = path
	}

	args := append(
		[]string{
			filepath.Join(invol.path, filename),
			outvol.path,
		},
		j.Args...,
	)
	task := exec.Command(name, args...)
	task.Stdin = os.Stdin
	task.Stdout = os.Stdout
	task.Stderr = stderr

	log.Info().Str(k, v).Msgf("running task %s", task.Path)
	if err := task.Run(); err != nil {
		return err
	}

	log.Debug().Str(k, v).Msg("uploading output")
	if err = o.os.PutOutput(m, outvol); err != nil {
		return err
	}

	return nil
}

func volume(path string) (*dirVolume, error) {
	return &dirVolume{path: path}, os.MkdirAll(path, 0755)
}

func (o *osRuntime) involume(m geocloud.Message) (*dirVolume, error) {
	return volume(filepath.Join(o.jobdir(m), "input"))
}

func (o *osRuntime) outvolume(m geocloud.Message) (*dirVolume, error) {
	return volume(filepath.Join(o.jobdir(m), "output"))
}

func (o *osRuntime) jobdir(m geocloud.Message) string {
	return filepath.Join(o.jobsdir(), m.GetID())
}

func (o *osRuntime) jobsdir() string {
	return filepath.Join(o.workdir, "jobs")
}
