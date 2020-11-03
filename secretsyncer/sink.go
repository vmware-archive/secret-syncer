package secretsyncer

import (
	"fmt"

	vaultapi "github.com/hashicorp/vault/api"
)

type VaultSink struct{}

func (vs VaultSink) PipelinePath(pp PipelinePath) string {
	return fmt.Sprintf("/concourse/%s/%s/%s", pp.Team, pp.Pipeline, pp.Secret)
}
func (vs VaultSink) WriteSimple(path string, val SimpleValue) error {
	// TODO how might this fail? invalid env vars?
	client, _ := vaultapi.NewClient(nil)
	_, err := client.Logical().Write(
		path,
		map[string]interface{}{"value": val},
	)
	return err
}
