package main

import (
	"fmt"
	"os"
)

func main() {
	logger, err := NewLogger()
	if err != nil {
		fmt.Println("failed to initialize logger")
		os.Exit(1)
	}

	// read Orkfile contents
	contents, err := Read(DEFAULT_ORKFILE)
	if err != nil {
		logger.Fatalf("failed to find Orkfile: %v", err)
	}
	orkfile := New(logger)
	if err := orkfile.Parse(contents); err != nil {
		logger.Fatalf("failed to parse Orkfile: %v", err)
	}

	if err := orkfile.Execute(os.Args[1]); err != nil {
		logger.Fatal(err.Error())
	}
}
