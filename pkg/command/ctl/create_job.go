package command

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/logsquaredn/rototiller/pkg/client"
	"github.com/spf13/cobra"
)

func NewCreateJobCommand() *cobra.Command {
	var (
		addr, apiKey, contentType string
		input, inputOf, outputOf  string
		file                      string
		query                     = map[string]string{}
		rpc                       bool
		createJobCmd              = &cobra.Command{
			Use:     "job",
			Aliases: []string{"j"},
			Args:    cobra.ExactArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				var (
					req  client.Request
					opts = []client.ClientOpt{}
				)
				if rpc {
					opts = append(opts, client.WithRPC)
				}

				c, err := client.New(addr, apiKey, opts...)
				if err != nil {
					cmd.PrintErrln(err)
					os.Exit(1)
				}

				if contentType == "" {
					switch file {
					case "", stdin:
					default:
						ext := filepath.Ext(file)
						if ext == "geojson" {
							ext = "json"
						}
						contentType = "application/" + ext
					}
				}

				switch {
				case input != "":
					req = client.NewJobWithInput(input, query)
				case inputOf != "":
					req = client.NewJobWithInputOfJob(inputOf, query)
				case outputOf != "":
					req = client.NewJobWithOutputOfJob(outputOf, query)
				default:
					var (
						r = cmd.InOrStdin()
					)
					switch file {
					case stdin:
					case "":
						cmd.PrintErrln("no input given")
						os.Exit(1)
					default:
						if r, err = os.Open(file); err != nil {
							cmd.PrintErrln(err)
							os.Exit(1)
						}
					}

					if contentType == "" {
						b, err := io.ReadAll(r)
						if err != nil {
							cmd.PrintErrln(err)
							os.Exit(1)
						}

						contentType = http.DetectContentType(b)
						r = bytes.NewReader(b)
					}

					req = client.NewJobFromInput(r, contentType, query)
				}

				j, err := c.CreateJob(args[0], req)
				if err != nil {
					cmd.PrintErrln(err)
					os.Exit(1)
				}

				encoder := json.NewEncoder(cmd.OutOrStdout())
				encoder.SetIndent("", "  ")

				if err = encoder.Encode(j); err != nil {
					cmd.PrintErrln(err)
					os.Exit(1)
				}
			},
		}
	)

	createJobCmd.Flags().BoolVar(&rpc, "rpc", false, "use RPC")
	createJobCmd.Flags().StringVar(&addr, "addr", "", "rototiller address")
	createJobCmd.Flags().StringVar(&apiKey, "api-key", "", "rototiller API key")
	createJobCmd.Flags().StringVarP(&file, "file", "f", "", "path to input file")
	_ = createJobCmd.MarkFlagFilename("file", "json", "yaml", "yml", "zip")
	createJobCmd.Flags().StringVar(&input, "input", "", "storage ID to use")
	createJobCmd.Flags().StringVar(&inputOf, "input-of", "", "job ID to use the input of")
	createJobCmd.Flags().StringVar(&outputOf, "output-of", "", "job ID to use the output of")
	createJobCmd.Flags().StringVar(&contentType, "content-type", "", "content-type")
	createJobCmd.Flags().StringToStringVarP(&query, "query", "q", map[string]string{}, "query params")

	return createJobCmd
}
