package main_test

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/jamieklassen/secret-syncer/secretsyncer"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestSuite(t *testing.T) {
	suite.Run(t, &SecretSyncerIntegrationSuite{
		Assertions: require.New(t),
	})
}

type SecretSyncerIntegrationSuite struct {
	suite.Suite
	*require.Assertions
	vaultClient *vaultapi.Client
}

func (s *SecretSyncerIntegrationSuite) SetupSuite() {
	rand.Seed(time.Now().UnixNano())
	ensureDefaultEnvVar("VAULT_TOKEN", "myroot")
	ensureDefaultEnvVar("VAULT_ADDR", "http://127.0.0.1:8200")
	var err error
	s.vaultClient, err = vaultapi.NewClient(nil)
	if err != nil {
		panic(fmt.Errorf("creating vault client: %s", err))
	}
	s.vaultClient.Sys().Mount(
		"concourse",
		&vaultapi.MountInput{
			Type:    "kv",
			Options: map[string]string{"version": "1"},
		},
	)
}

func ensureDefaultEnvVar(envVarName, defaultValue string) {
	if os.Getenv(envVarName) == "" {
		os.Setenv(envVarName, defaultValue)
	}
}

func (s *SecretSyncerIntegrationSuite) TestWritesSimplePipelineSecretInEmptyDevVaultUsingTokenAuth() {
	secretName := randomSecretName()
	secretsFile := writeFixture(
		fmt.Sprintf(
			`team_name/pipeline_name/%s: credential`,
			secretName,
		),
	)
	defer os.Remove(secretsFile)

	syncer, _ := secretsyncer.FileSyncer(secretsFile)
	syncer.Sync()

	s.HasSecret(
		fmt.Sprintf("/concourse/team_name/pipeline_name/%s", secretName),
		map[string]interface{}{"value": "credential"},
	)
}

func (s *SecretSyncerIntegrationSuite) TestVaultSinkRecursivelyDeletesSecrets() {
	client := secretsyncer.DefaultVaultClient{s.vaultClient}
	client.Write(
		"/concourse/main/concourse/foo",
		map[string]interface{}{"value": "bar"},
	)
	client.Write(
		"/concourse/main/foo",
		map[string]interface{}{"value": "bar"},
	)
	client.Write(
		"/concourse/shared/foo",
		map[string]interface{}{"value": "bar"},
	)
	sink := &secretsyncer.VaultSink{client}

	sink.Clear()

	paths, _ := client.List("concourse/")
	s.Equal([]string{"concourse/"}, paths)
}

func randomSecretName() string {
	return strconv.Itoa(rand.Int())
}

func writeFixture(content string) string {
	tmpfile, err := ioutil.TempFile("", "")
	if err != nil {
		panic(fmt.Errorf("creating temp file for test fixture: %s", err))
	}
	defer tmpfile.Close()
	_, err = tmpfile.Write([]byte(content))
	if err != nil {
		panic(fmt.Errorf("writing temp file for test fixture: %s", err))
	}
	return tmpfile.Name()

}

func (s *SecretSyncerIntegrationSuite) HasSecret(path string, expected map[string]interface{}) {
	vaultSecret, err := s.vaultClient.Logical().Read(path)
	if err != nil {
		s.T().Fatalf("reading vault secret '%s': %s", path, err)
	}
	if vaultSecret == nil {
		s.T().Fatalf("reading vault secret '%s': not found", path)
	}
	s.Equal(expected, vaultSecret.Data)
}
