package command

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"runtime"
	"strings"

	"github.com/logsquaredn/rototiller"
	"github.com/logsquaredn/rototiller/internal/static"
	"github.com/spf13/cobra"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func New() *cobra.Command {
	var (
		verbosity int
		port      int64
		proxyAddr string
		cmd       = &cobra.Command{
			Use:     "rotoui",
			Version: rototiller.GetSemver(),
			PersistentPreRun: func(cmd *cobra.Command, args []string) {
				cmd.SetContext(rototiller.WithLogger(cmd.Context(), rototiller.NewLogger().V(2-verbosity)))
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

				u, err := url.Parse(proxyAddr)
				if err != nil {
					return err
				}

				var (
					apiReverseProxy = httputil.NewSingleHostReverseProxy(u)
					uiFileServer    = http.FileServer(http.FS(static.FS))
					srv             = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/swagger/") {
							apiReverseProxy.ServeHTTP(w, r)
						} else {
							uiFileServer.ServeHTTP(w, r)
						}
					})
				)

				logr.Info("serving on " + addr)
				return http.Serve(l, h2c.NewHandler(srv, &http2.Server{})) //nolint:gosec,nolintlint // lint in GitHub Actions doesn't like this
			},
		}
	)

	cmd.PersistentFlags().CountVarP(&verbosity, "verbose", "V", "verbose")
	cmd.Flags().StringVar(&proxyAddr, "proxy-addr", os.Getenv("ROTOTILLER_PROXY_ADDR"), "proxy address")
	cmd.Flags().Int64VarP(&port, "port", "p", 8080, "listen port")
	cmd.SetVersionTemplate("{{ .Name }}{{ .Version }} " + runtime.Version() + "\n")

	return cmd
}
