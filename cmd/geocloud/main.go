package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/logsquaredn/geocloud/cmd"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	o := zerolog.ConsoleWriter{
		Out: os.Stdout,
		TimeFormat: time.RFC3339Nano,
	}

	o.FormatTimestamp = func(i interface{}) string {
		return fmt.Sprintf("time=\"%s\"", i)
	}

	o.FormatLevel = func(i interface{}) string {
		return fmt.Sprintf("level=%s", i)
	}

	o.FormatMessage = func(i interface{}) string {
		s := fmt.Sprintf("%s", i)
		if strings.Contains(s, " ") {
			s = fmt.Sprintf("msg=\"%s\"", s)
		}
		return s
	}

	o.FormatFieldName = func(i interface{}) string {
		return fmt.Sprintf("%s=", i)
	}

	o.FormatFieldValue = func(i interface{}) string {
		s := fmt.Sprintf("%s", i)
		if strings.Contains(s, " ") {
			s = fmt.Sprintf("\"%s\"", s)
		}
		return s
	}

	log.Logger = zerolog.New(o).With().Timestamp().Logger()
}

var p *flags.Parser

func init() {
	p = flags.NewParser(cmd.GeocloudCmd, flags.HelpFlag)
	p.NamespaceDelimiter = "-"
}

func main() {
	if _, err := p.Parse(); err == flags.ErrHelp {
		fmt.Println(err)
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}
