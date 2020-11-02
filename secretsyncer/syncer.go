package secretsyncer

import (
	vaultapi "github.com/hashicorp/vault/api"
)

type Syncer struct {
	Source Source
	Sink   Sink
}

func FileSyncer(secretsFile string) Syncer {
	return Syncer{SecretsFile: secretsFile}
}

func (s Syncer) Sync() {
	client, _ := vaultapi.NewClient(nil)
	client.Logical().Write(
		"/concourse/team_name/pipeline_name/pipeline_scoped",
		map[string]interface{}{"value": "credential"},
	)
}

type Credential struct {
	Location interface{}
	Value    interface{}
}

type PipelinePath struct {
	Team     string
	Pipeline string
	Secret   string
}

type SimpleValue string
