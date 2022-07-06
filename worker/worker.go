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
	"github.com/logsquaredn/rototiller"
	"github.com/logsquaredn/rototiller/datastore"
	"github.com/logsquaredn/rototiller/internal/conf"
	"github.com/logsquaredn/rototiller/objectstore"
	"github.com/rs/zerolog/log"
	"mellium.im/sysexit"
)

type Worker struct {
	ds      *datastore.Postgres
	os      *objectstore.S3
	workdir string
}

var (
	envVarInputFile = fmt.Sprintf("%sINPUT_FILE", conf.EnvPrefix)
	envVarOutputDir = fmt.Sprintf("%sOUTPUT_DIR", conf.EnvPrefix)
)

func New(opts *Opts) (*Worker, error) {
	return &Worker{
		ds:      opts.Datastore,
		os:      opts.Objectstore,
		workdir: opts.WorkDir,
	}, nil
}

func (o *Worker) Send(m rototiller.Message) error {
	k, v := "id", m.GetID()
	log.Info().Str(k, v).Msg("processing message")

	log.Trace().Str(k, v).Msg("getting job from datastore")
	j, err := o.ds.GetJob(m)
	if err != nil {
		return err
	}

	switch j.Status {
	case rototiller.JobStatusComplete, rototiller.JobStatusInProgress:
		return nil
	}

	stderr := new(bytes.Buffer)
	defer func() {
		j.EndTime = time.Now()
		jobErr := stderr.Bytes()
		switch {
		case len(jobErr) > 0:
			j.Error = string(jobErr)
			j.Status = rototiller.JobStatusError
		case err != nil:
			j.Error = err.Error()
			j.Status = rototiller.JobStatusError
		default:
			j.Status = rototiller.JobStatusComplete
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

	inputStorage, err := o.ds.GetStorage(rototiller.Msg(j.InputID))
	if err != nil {
		return err
	}

	switch inputStorage.Status {
	case rototiller.StorageStatusFinal, rototiller.StorageStatusUnusable:
		return fmt.Errorf("input storage status '%s'", inputStorage.Status)
	}

	defer func() {
		_, err = o.ds.UpdateStorage(inputStorage)
		if err != nil {
			log.Err(err).Msg("updating input storage")
		}
	}()

	j.Status = rototiller.JobStatusInProgress
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
	input, err := o.os.GetObject(rototiller.Msg(j.InputID))
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
		_        = invol.Walk(func(_ string, f rototiller.File, e error) error {
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

	task := exec.Command(t.Type.String()) //nolint:gosec
	// start with current env minus configuration that might contain secrets
	// e.g. ROTOTILLER_POSTGRES_PASSWORD
	task.Env = js.Filter(os.Environ(), func(e string, _ int, _ []string) bool {
		return !(strings.HasPrefix(e, conf.EnvPrefix) || strings.HasPrefix(e, "AWS_"))
	})
	// add input file path and output dir path
	task.Env = append(task.Env,
		fmt.Sprintf(
			"%s=%s",
			envVarInputFile,
			filepath.Join(o.inputVolumePath(j), filename),
		),
		fmt.Sprintf(
			"%s=%s",
			envVarOutputDir,
			o.outputVolumePath(j),
		),
	)
	// add arbitrary args defined by the task entry in the datastore
	// e.g. task.type = 'reproject'
	//		=> task.params = ['target-projection'],
	//      => ROTOTILLER_TARGET_PROJECTION=${?target-projection}
	task.Env = append(task.Env, js.Map(j.Args, func(a string, i int, _ []string) string {
		return fmt.Sprintf(
			"%s%s=%s",
			conf.EnvPrefix,
			strings.ToUpper(
				conf.HyphenToUnderscoreReplacer.Replace(t.Params[i]),
			),
			a,
		)
	})...)
	task.Stdin = os.Stdin
	task.Stdout = os.Stdout
	task.Stderr = stderr

	log.Info().Str(k, v).Msgf("running task %s", task.Path)
	if err := task.Run(); err != nil {
		return err
	}

	switch task.ProcessState.ExitCode() {
	case int(sysexit.Ok):
		inputStorage.Status = rototiller.StorageStatusTransformable
	case int(sysexit.ErrData), int(sysexit.ErrNoInput):
		inputStorage.Status = rototiller.StorageStatusUnusable
		err = fmt.Errorf("unusable input")
		return err
	case int(sysexit.ErrCantCreat):
		err = fmt.Errorf("can't create output file")
		return err
	case int(sysexit.ErrConfig):
		err = fmt.Errorf("configuration error")
		return err
	default:
		inputStorage.Status = rototiller.StorageStatusUnknown
		err = fmt.Errorf("unknown error")
		return err
	}

	log.Trace().Str(k, v).Msg("creating output storage")
	ost, err := o.ds.CreateStorage(&rototiller.Storage{
		CustomerID: j.CustomerID,
		Status:     js.Ternary(t.Kind == rototiller.TaskKindLookup, rototiller.StorageStatusFinal, rototiller.StorageStatusTransformable),
	})
	if err != nil {
		return err
	}
	j.OutputID = ost.ID

	log.Debug().Str(k, v).Msg("uploading output")
	return o.os.PutObject(rototiller.Msg(j.OutputID), outvol)
}

func (o *Worker) inputVolumePath(m rototiller.Message) string {
	return filepath.Join(o.jobDir(m), "input")
}

func (o *Worker) outputVolumePath(m rototiller.Message) string {
	return filepath.Join(o.jobDir(m), "output")
}

func (o *Worker) inputVolume(m rototiller.Message) (rototiller.Volume, error) {
	return rototiller.NewDirVolume(o.inputVolumePath(m))
}

func (o *Worker) outputVolume(m rototiller.Message) (rototiller.Volume, error) {
	return rototiller.NewDirVolume(o.outputVolumePath(m))
}

func (o *Worker) jobDir(m rototiller.Message) string {
	return filepath.Join(o.jobsDir(), m.GetID())
}

func (o *Worker) jobsDir() string {
	return filepath.Join(o.workdir, "jobs")
}
