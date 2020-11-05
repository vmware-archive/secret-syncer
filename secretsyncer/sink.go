package secretsyncer

import (
	"fmt"

	vaultapi "github.com/hashicorp/vault/api"
)

type VaultSink struct {
	Client VaultClient
}

// TODO these templates depend on the patterns concourse is expecting
func (vs *VaultSink) PipelinePath(pp PipelinePath) string {
	return fmt.Sprintf("/concourse/%s/%s/%s", pp.Team, pp.Pipeline, pp.Secret)
}
func (vs *VaultSink) TeamPath(tp TeamPath) string {
	return fmt.Sprintf("/concourse/%s/%s", tp.Team, tp.Secret)
}
func (vs *VaultSink) SharedPath(sp SharedPath) string {
	return fmt.Sprintf("/concourse/shared/%s", sp.Secret)
}
func (vs *VaultSink) WriteSimple(path string, val SimpleValue) error {
	return vs.write(path, val)
}
func (vs *VaultSink) WriteCompound(path string, val CompoundValue) error {
	return vs.write(path, val)
}
func (vs *VaultSink) write(path string, val interface{}) error {
	var data map[string]interface{}
	switch v := val.(type) {
	case SimpleValue:
		data = map[string]interface{}{"value": v}
	case CompoundValue:
		data = v
	}
	err := vs.Client.Write(path, data)
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

type VaultClient interface {
	Write(string, map[string]interface{}) error
}

type vaultClient struct {
	*vaultapi.Client
}

func (vc vaultClient) Write(path string, data map[string]interface{}) error {
	_, err := vc.Logical().Write(path, data)
	return err
}
