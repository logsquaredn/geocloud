package shared

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func SetupLogging() {
	output := zerolog.ConsoleWriter{
		NoColor: true,
		Out: os.Stdout,
		TimeFormat: time.RFC3339Nano,
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
