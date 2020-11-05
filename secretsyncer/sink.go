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
func (vs *VaultSink) Clear() error {
	paths, err := vs.Client.List("concourse/")
	if err != nil {
		return err
	}
	for _, path := range paths {
		err = vs.Client.Delete(path)
		if err != nil {
			return err
		}
	}
	return nil
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
	List(string) ([]string, error)
	Delete(string) error
}

type DefaultVaultClient struct {
	*vaultapi.Client
}

func (dvc DefaultVaultClient) Write(path string, data map[string]interface{}) error {
	_, err := dvc.Logical().Write(path, data)
	return err
}
func (dvc DefaultVaultClient) List(path string) ([]string, error) {
	r, err := dvc.Logical().List(path)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return []string{path}, nil
	}
	paths := []string{}
	for _, p := range r.Data["keys"].([]interface{}) {
		fullPath := path + p.(string)
		nestedPaths, err := dvc.List(fullPath)
		if err != nil {
			return nil, err
		}
		paths = append(paths, nestedPaths...)
	}
	return paths, nil
}
func (dvc DefaultVaultClient) Delete(path string) error {
	_, err := dvc.Logical().Delete(path)
	return err
}
