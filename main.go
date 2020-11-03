package main

import (
	"os"

	"github.com/jamieklassen/secret-syncer/secretsyncer"
)

func main() {
	secretsyncer.FileSyncer(os.Args[1]).Sync()
}

// file -> THIS -> vault
// GCS -> file -> THIS -> k8s -> vault
