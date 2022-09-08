package worker

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/frantjc/go-js"
	"github.com/logsquaredn/rototiller"
	"github.com/logsquaredn/rototiller/pkg/store/blob/bucket"
	"github.com/logsquaredn/rototiller/pkg/store/data/postgres"
	"github.com/logsquaredn/rototiller/pkg/volume"
	"google.golang.org/protobuf/types/known/timestamppb"
	"mellium.im/sysexit"
)

var (
	HyphenToUnderscoreReplacer = strings.NewReplacer("-", "_")
)

type Worker struct {
	*postgres.Datastore
	*bucket.Blobstore
	WorkingDir string
}

const (
	EnvVarInputFile = "ROTOTILLER_INPUT_FILE"
	EnvVarOutputDir = "ROTOTILLER_OUTPUT_DIR"
)

func New(ctx context.Context, workingDir string, datastore *postgres.Datastore, blobstore *bucket.Blobstore) (*Worker, error) {
	return &Worker{
		Datastore:  datastore,
		Blobstore:  blobstore,
		WorkingDir: workingDir,
	}, nil
}

func (w *Worker) DoJob(ctx context.Context, id string) error {
	logr := rototiller.LoggerFrom(ctx)

	j, err := w.Datastore.GetJob(id)
	if err != nil {
		return err
	}

	switch j.Status {
	case rototiller.JobStatusComplete.String(), rototiller.JobStatusInProgress.String():
		return nil
	}

	stderr := new(bytes.Buffer)
	defer func() {
		j.EndTime = timestamppb.New(time.Now())
		jobErr := stderr.Bytes()
		switch {
		case len(jobErr) > 0:
			j.Error = string(jobErr)
			j.Status = rototiller.JobStatusError.String()
		case err != nil:
			j.Error = err.Error()
			j.Status = rototiller.JobStatusError.String()
		default:
			j.Status = rototiller.JobStatusComplete.String()
		}

		if updatedJob, err := w.Datastore.UpdateJob(j); err != nil {
			j = updatedJob
		} else {
			logr.Error(err, "updating job", "id", j.GetId())
		}
	}()

	inputStorage, err := w.Datastore.GetStorage(j.GetInputId())
	if err != nil {
		return err
	}

	switch inputStorage.Status {
	case rototiller.StorageStatusFinal.String(), rototiller.StorageStatusUnusable.String():
		return fmt.Errorf("input storage status '%s'", inputStorage.Status)
	}

	defer func() {
		_, _ = w.Datastore.UpdateStorage(inputStorage)
	}()

	j.Status = rototiller.JobStatusInProgress.String()
	j, err = w.Datastore.UpdateJob(j)
	if err != nil {
		return err
	}

	t, err := w.Datastore.GetTaskByJobID(id)
	if err != nil {
		return err
	}

	invol, err := w.inputVolume(id)
	if err != nil {
		return err
	}
	defer os.RemoveAll(w.jobDir(id))

	input, err := w.Blobstore.GetObject(ctx, j.GetInputId())
	if err != nil {
		return err
	}

	if err = input.Download(w.inputVolumePath(j.GetId())); err != nil {
		return err
	}

	outvol, err := w.outputVolume(id)
	if err != nil {
		return err
	}

	var (
		filename string
		_        = invol.Walk(func(_ string, f volume.File, e error) error {
			if e != nil {
				return e
			}
			filename = f.GetName()
			// we only expect 1 input, so use the first one we find and end the Walk
			return fmt.Errorf("found")
		})
	)

	if filename == "" {
		return fmt.Errorf("no input found")
	}

	task := exec.Command(t.Type) //nolint:gosec // t.Type is guaranteed to refer to a Task binary
	// start with current env minus configuration that might contain secrets
	// e.g. ROTOTILLER_POSTGRES_PASSWORD
	task.Env = js.Filter(os.Environ(), func(e string, _ int, _ []string) bool {
		return !(strings.HasPrefix(e, "ROTOTILLER_") || strings.HasPrefix(e, "AWS_") || strings.Contains(e, "PASSWORD") || strings.Contains(e, "USERNAME") || strings.Contains(e, "SECRET"))
	})
	// add input file path and output dir path
	task.Env = append(task.Env,
		EnvVarInputFile+"="+filepath.Join(w.inputVolumePath(j.GetId()), filename),
		EnvVarOutputDir+"="+w.outputVolumePath(j.GetId()),
	)
	// add arbitrary args defined by the task entry in the datastore
	// e.g. task.type = 'reproject'
	//		=> task.params = ['target-projection'],
	//      => ROTOTILLER_TARGET_PROJECTION=${?target-projection}
	task.Env = append(task.Env, js.Map(j.Args, func(a string, i int, _ []string) string {
		return "ROTOTILLER_" + strings.ToUpper(HyphenToUnderscoreReplacer.Replace(t.Params[i])) + "=" + a
	})...)
	task.Stdin = os.Stdin
	task.Stdout = os.Stdout
	task.Stderr = stderr

	if err := task.Run(); err != nil {
		return err
	}

	switch task.ProcessState.ExitCode() {
	case int(sysexit.Ok):
		inputStorage.Status = rototiller.StorageStatusTransformable.String()
	case int(sysexit.ErrData), int(sysexit.ErrNoInput):
		inputStorage.Status = rototiller.StorageStatusUnusable.String()
		err = fmt.Errorf("unusable input")
	case int(sysexit.ErrCantCreat):
		err = fmt.Errorf("can't create output file")
	case int(sysexit.ErrConfig):
		err = fmt.Errorf("configuration error")
	default:
		inputStorage.Status = rototiller.StorageStatusUnknown.String()
		err = fmt.Errorf("unknown error")
	}
	if err != nil {
		return err
	}

	ost, err := w.Datastore.CreateStorage(&rototiller.Storage{
		CustomerId: j.GetCustomerId(),
		Status:     js.Ternary(t.GetKind() == rototiller.TaskKindLookup.String(), rototiller.StorageStatusFinal.String(), rototiller.StorageStatusTransformable.String()),
	})
	if err != nil {
		return err
	}
	j.OutputId = ost.GetId()

	return w.Blobstore.PutObject(ctx, j.GetOutputId(), outvol)
}

func (w *Worker) inputVolumePath(id string) string {
	return filepath.Join(w.jobDir(id), "input")
}

func (w *Worker) outputVolumePath(id string) string {
	return filepath.Join(w.jobDir(id), "output")
}

func (w *Worker) inputVolume(id string) (volume.Volume, error) {
	return volume.NewDir(w.inputVolumePath(id))
}

func (w *Worker) outputVolume(id string) (volume.Volume, error) {
	return volume.NewDir(w.outputVolumePath(id))
}

func (w *Worker) jobDir(id string) string {
	return filepath.Join(w.jobsDir(), id)
}

func (w *Worker) jobsDir() string {
	return filepath.Join(w.WorkingDir, "jobs")
}
