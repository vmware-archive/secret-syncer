package main

import (
	"fmt"
	"os"

	"github.com/jamieklassen/secret-syncer/secretsyncer"
)

func main() {
	// TODO validate args
	syncer, err := secretsyncer.FileSyncer(os.Args[1])
	if err != nil {
		fmt.Printf("constructing syncer: %s\n", err)
		os.Exit(1)
	}
	err = syncer.Sync()
	if err != nil {
		fmt.Printf("syncing: %s\n", err)
		os.Exit(1)
	}
}

// file -> THIS -> vault
// GCS -> file -> THIS -> k8s -> vault
