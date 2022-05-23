package main

import (
	"io"
	"os"

	"github.com/spf13/viper"
)

func getInput() ([]byte, error) {
	f, err := os.Open(viper.GetString("file"))
	if err != nil {
		return nil, err
	}

	return io.ReadAll(f)
}
