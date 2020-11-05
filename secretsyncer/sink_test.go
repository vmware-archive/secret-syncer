package secretsyncer_test

import (
	"errors"

	"github.com/jamieklassen/secret-syncer/secretsyncer"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type VaultSinkSuite struct {
	suite.Suite
	*require.Assertions
}

type MockVaultClient struct {
	secrets map[string]map[string]interface{}
}

func (mvc *MockVaultClient) Write(path string, data map[string]interface{}) error {
	if mvc.secrets == nil {
		mvc.secrets = map[string]map[string]interface{}{}
	}
	mvc.secrets[path] = data
	return nil
}
func (mvc *MockVaultClient) List(path string) ([]string, error) {
	return nil, nil
}
func (mvc *MockVaultClient) Delete(path string) error {
	delete(mvc.secrets, path)
	return nil
}
func (mvc *MockVaultClient) Read(path string) map[string]interface{} {
	return mvc.secrets[path]
}

func (s *VaultSinkSuite) TestWritesSimpleValueWithKey() {
	client := &MockVaultClient{}
	sink := &secretsyncer.VaultSink{client}

	sink.WriteSimple("/path/to/secret", secretsyncer.SimpleValue("cred"))

	s.Equal(
		map[string]interface{}{"value": secretsyncer.SimpleValue("cred")},
		client.Read("/path/to/secret"),
	)
}

func (s *VaultSinkSuite) TestWritesCompoundValueAsIs() {
	client := &MockVaultClient{}
	sink := &secretsyncer.VaultSink{client}

	sink.WriteCompound(
		"/path/to/secret",
		secretsyncer.CompoundValue{
			"username": secretsyncer.SimpleValue("user"),
			"password": secretsyncer.SimpleValue("pass"),
		},
	)

	s.Equal(
		map[string]interface{}{
			"username": secretsyncer.SimpleValue("user"),
			"password": secretsyncer.SimpleValue("pass"),
		},
		client.Read("/path/to/secret"),
	)
}

type ErroringVaultClient struct {
	err error
}

func (evc *ErroringVaultClient) Write(path string, data map[string]interface{}) error {
	return evc.err
}
func (evc *ErroringVaultClient) List(path string) ([]string, error) {
	return nil, evc.err
}
func (evc *ErroringVaultClient) Delete(path string) error {
	return evc.err
}
func (s *VaultSinkSuite) TestSurfacesErrorFromClient() {
	client := &ErroringVaultClient{errors.New("a scary error from vault")}
	sink := &secretsyncer.VaultSink{client}

	err := sink.WriteSimple("/path/to/secret", secretsyncer.SimpleValue("cred"))

	s.EqualError(
		err,
		"writing secret value 'cred' to vault path '/path/to/secret': a scary error from vault",
	)
}
