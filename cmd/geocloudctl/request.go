package main

import (
	"encoding/json"
	"os"

	"github.com/logsquaredn/geocloud"
	"github.com/spf13/cobra"
)

func getRequest(cmd *cobra.Command) (geocloud.Request, error) {
	var (
		req geocloud.Request
		i   = cmd.Flag("input").Value.String()
		io  = cmd.Flag("input-of").Value.String()
		oo  = cmd.Flag("output-of").Value.String()
		dec = json.NewDecoder(os.Stdin)
	)
	switch {
	case i == fromStdin:
		storage := &geocloud.Storage{}
		if err := dec.Decode(storage); err != nil {
			return nil, nil
		}
		req = geocloud.NewJobWithInput(storage.ID, jobQuery)
	case i != "":
		req = geocloud.NewJobWithInput(i, jobQuery)
	case io == fromStdin:
		job := &geocloud.Job{}
		if err := dec.Decode(job); err != nil {
			return nil, err
		}
		req = geocloud.NewJobWithInputOfJob(job.ID, jobQuery)
	case io != "":
		req = geocloud.NewJobWithInputOfJob(io, jobQuery)
	case oo == fromStdin:
		job := &geocloud.Job{}
		if err := dec.Decode(job); err != nil {
			return nil, err
		}
		req = geocloud.NewJobWithOutputOfJob(job.ID, jobQuery)
	case oo != "":
		req = geocloud.NewJobWithOutputOfJob(oo, jobQuery)
	default:
		f, contentType, err := getInput(cmd)
		if err != nil {
			return nil, err
		}
		req = geocloud.NewJobFromInput(f, contentType, jobQuery)
	}

	return req, nil
}
