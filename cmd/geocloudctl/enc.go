package main

import (
	"encoding/json"
	"os"
)

func write(i any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent(" ", "  ")
	return enc.Encode(i)
}
