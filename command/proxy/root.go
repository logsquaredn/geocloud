package command

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime"

	"github.com/logsquaredn/rototiller"
	"github.com/logsquaredn/rototiller/proxy"
	"github.com/spf13/cobra"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func NewRoot() *cobra.Command {
	var (
		verbosity                          int
		port                               int64
		proxyAddr, smtpAddr, smtpFrom, key string
		rootCmd                            = &cobra.Command{
			Use:     "rotoproxy",
			Version: rototiller.GetSemver(),
			PersistentPreRun: func(cmd *cobra.Command, args []string) {
				cmd.SetContext(rototiller.WithLogger(cmd.Context(), rototiller.NewLogger().V(verbosity)))
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				var (
					ctx  = cmd.Context()
					logr = rototiller.LoggerFrom(ctx)
					addr = fmt.Sprintf(":%d", port)
				)

				l, err := net.Listen("tcp", addr)
				if err != nil {
					return err
				}

				srv, err := proxy.NewHandler(ctx, proxyAddr, smtpAddr, smtpFrom, key)
				if err != nil {
					return err
				}

				logr.Info("serving on " + addr)
				return http.Serve(l, h2c.NewHandler(srv, &http2.Server{})) //nolint:gosec,nolintlint // lint in GitHub Actions doesn't like this
			},
		}
	)

	rootCmd.PersistentFlags().CountVarP(&verbosity, "verbose", "V", "verbose")
	rootCmd.Flags().StringVar(&proxyAddr, "proxy-addr", os.Getenv("ROTOTILLER_PROXY_ADDR"), "proxy address")
	rootCmd.Flags().StringVar(&smtpAddr, "smtp-addr", os.Getenv("ROTOTILLER_SMTP_ADDR"), "smtp address")
	rootCmd.Flags().StringVar(&smtpFrom, "smtp-from", os.Getenv("ROTOTILLER_SMTP_FROM"), "smtp from")
	// TODO use viper
	// setting secrets as the default can easily leak the secret
	// if the command exits with an error
	// rootCmd.Flags().StringVar(&smtpUsername, "smtp-username", os.Getenv("ROTOTILLER_SMTP_USERNAME"), "smtp username")
	// rootCmd.Flags().StringVar(&smtpPassword, "smtp-password", os.Getenv("ROTOTILLER_SMTP_PASSWORD"), "smtp password")
	rootCmd.Flags().StringVar(&key, "key", "", "key")
	rootCmd.Flags().Int64VarP(&port, "port", "p", 8080, "listen port")
	rootCmd.SetVersionTemplate("{{ .Name }}{{ .Version }} " + runtime.Version() + "\n")

	return rootCmd
}