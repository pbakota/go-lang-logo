package main

import (
	"io"
	"os"

	"rs.lab/go-logo/logo"
)

func main() {
	text, err := io.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}

	r := logo.NewRuntime()
	r.Trace = true
	err = r.Run(string(text))

	if err != nil {
		panic(err)
	}
}
