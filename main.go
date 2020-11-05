package main

import (
	"fmt"
	"os"

	"github.com/jamieklassen/secret-syncer/secretsyncer"
)

func main() {
	err := secretsyncer.FileSyncer(os.Args[1]).Sync()
	if err != nil {
		fmt.Println(err)
	}
}

// file -> THIS -> vault
// GCS -> file -> THIS -> k8s -> vault
