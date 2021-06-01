package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/logsquaredn/geocloud"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	var cmd GeocloudCmd

	cmd.Version = func() {
		_, err := fmt.Println(geocloud.Version)
		if err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}

	logSetup()
	parser := flags.NewParser(&cmd, flags.HelpFlag)
	parser.NamespaceDelimiter = "-"
	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			fmt.Println(err)
		} else {
			fmt.Fprintf(os.Stderr, "error: %s\n", err)
			os.Exit(1)
		}
	}
	os.Exit(0)
}

func logSetup() {
	output := zerolog.ConsoleWriter{
		NoColor: true,
		Out: os.Stdout,
		TimeFormat: time.RFC3339,
	}

	output.FormatTimestamp = func(i interface{}) string {
		return fmt.Sprintf("time=\"%s\"", i)
	}

	output.FormatLevel = func(i interface{}) string {
		return fmt.Sprintf("level=%s", i)
	}

	output.FormatMessage = func(i interface{}) string {
		if strings.Contains(fmt.Sprintf("%s", i), " ") {
			return fmt.Sprintf("msg=\"%s\"", i)
		}

		return fmt.Sprintf("msg=%s", i)
	}

	output.FormatFieldName = func(i interface{}) string {
		return fmt.Sprintf("%s=", i)
	}

	output.FormatFieldValue = func(i interface{}) string {
		if strings.Contains(fmt.Sprintf("%s", i), " ") {
			return fmt.Sprintf("\"%s\"", i)
		}

		return fmt.Sprintf("%s", i)
	}

	log.Logger = zerolog.New(output).With().Timestamp().Logger()
}

func (cmd *VersionCmd) Execute(args []string) error {
	_, err := fmt.Println(geocloud.Version)
	return err
}
