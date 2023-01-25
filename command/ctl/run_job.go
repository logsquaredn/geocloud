package command

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/logsquaredn/rototiller/client"
	"github.com/spf13/cobra"
)

func NewRunJob() *cobra.Command {
	var (
		addr, apiKey, contentType string
		input, inputOf, outputOf  string
		file                      string
		query                     = map[string]string{}
		cmd                       = &cobra.Command{
			Use:     "job",
			Aliases: []string{"j"},
			Args:    cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				var (
					req  client.Request
					opts = []client.ClientOpt{}
				)
				c, err := client.New(addr, apiKey, opts...)
				if err != nil {
					return err
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
					r := cmd.InOrStdin()
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

					switch contentType {
					case "application/json", "application/zip":
					default:
						contentType = "application/json"
					}

					req = client.NewJobFromInput(r, contentType, query)
				}

				j, err := c.RunJob(args[0], req)
				if err != nil {
					return err
				}

				encoder := json.NewEncoder(cmd.OutOrStdout())
				encoder.SetIndent("", "  ")

				return encoder.Encode(j)
			},
		}
	)

	cmd.Flags().StringVar(&addr, "addr", "", "rototiller address")
	cmd.Flags().StringVar(&apiKey, "api-key", "", "rototiller API key")
	cmd.Flags().StringVarP(&file, "file", "f", "", "path to input file")
	_ = cmd.MarkFlagFilename("file", "json", "yaml", "yml", "zip")
	cmd.Flags().StringVar(&input, "input", "", "storage ID to use")
	cmd.Flags().StringVar(&inputOf, "input-of", "", "job ID to use the input of")
	cmd.Flags().StringVar(&outputOf, "output-of", "", "job ID to use the output of")
	cmd.Flags().StringVar(&contentType, "content-type", "", "content type")
	cmd.Flags().StringToStringVarP(&query, "query", "q", map[string]string{}, "query params")

	return cmd
}
