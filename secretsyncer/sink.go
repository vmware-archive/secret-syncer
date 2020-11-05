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
	client, err := vaultapi.NewClient(nil)
	if err != nil {
		return fmt.Errorf("creating vault client: %s", err)
	}
	_, err = client.Logical().Write(
		path,
		map[string]interface{}{"value": val},
	)
	if err != nil {
		return fmt.Errorf(
			"writing secret value '%s' to vault path '%s': %s",
			val,
			path,
			err,
		)
	}
	return nil
}
