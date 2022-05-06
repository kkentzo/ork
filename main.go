package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	cfg, err := ParseOrkfile("Orkfile.yml")
	if err != nil {
		log.Fatalf("failed to parse Orkfile: %v", err)
	}

	registry := NewTaskRegistry(cfg)

	task := cfg.Global.Default
	if len(os.Args) == 2 {
		task = os.Args[1]
	}
	if t, ok := registry[task]; ok {
		if err := t.Execute(); err != nil {
			fmt.Printf("[%s] error: %v\n", t.Name, err)
		}
	} else {
		fmt.Printf("[%s] task not found in Orkfile\n", task)
	}
}
