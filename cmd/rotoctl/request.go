package main

import (
	"encoding/json"
	"os"

	"github.com/logsquaredn/rototiller"
	"github.com/spf13/cobra"
)

func getRequest(cmd *cobra.Command) (rototiller.Request, error) {
	var (
		req rototiller.Request
		i   = cmd.Flag("input").Value.String()
		io  = cmd.Flag("input-of").Value.String()
		oo  = cmd.Flag("output-of").Value.String()
		dec = json.NewDecoder(os.Stdin)
	)
	switch {
	case i == fromStdin:
		storage := &rototiller.Storage{}
		if err := dec.Decode(storage); err != nil {
			return nil, nil
		}
		req = rototiller.NewJobWithInput(storage.ID, jobQuery)
	case i != "":
		req = rototiller.NewJobWithInput(i, jobQuery)
	case io == fromStdin:
		job := &rototiller.Job{}
		if err := dec.Decode(job); err != nil {
			return nil, err
		}
		req = rototiller.NewJobWithInputOfJob(job.ID, jobQuery)
	case io != "":
		req = rototiller.NewJobWithInputOfJob(io, jobQuery)
	case oo == fromStdin:
		job := &rototiller.Job{}
		if err := dec.Decode(job); err != nil {
			return nil, err
		}
		req = rototiller.NewJobWithOutputOfJob(job.ID, jobQuery)
	case oo != "":
		req = rototiller.NewJobWithOutputOfJob(oo, jobQuery)
	default:
		f, contentType, err := getInput(cmd)
		if err != nil {
			return nil, err
		}
		req = rototiller.NewJobFromInput(f, contentType, jobQuery)
	}

	return req, nil
}
