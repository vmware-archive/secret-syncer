package secretsyncer

import (
	"fmt"
	"io/ioutil"

	vaultapi "github.com/hashicorp/vault/api"
)

type Syncer struct {
	Source Source
	Sink   Sink
}

type Source interface {
	Read() (Data, error)
}
type Sink interface {
	WriteSimple(string, SimpleValue) error
	WriteCompound(string, CompoundValue) error
	Clear() error
	// TODO determining path templates is a slightly different
	// responsibility than writing secrets. split out a different interface.
	PipelinePath(PipelinePath) string
	TeamPath(TeamPath) string
	SharedPath(SharedPath) string
}

func FileSyncer(secretsFile string) (Syncer, error) {
	fileBytes, err := ioutil.ReadFile(secretsFile)
	if err != nil {
		return Syncer{}, err
	}
	client, err := vaultapi.NewClient(nil)
	if err != nil {
		return Syncer{}, fmt.Errorf("creating vault client: %s", err)
	}
	return Syncer{
		Source: BytesSource{fileBytes},
		Sink:   &VaultSink{DefaultVaultClient{client}},
	}, nil
}

func (s Syncer) Sync() error {
	data, err := s.Source.Read()
	if err != nil {
		return fmt.Errorf("reading secrets: %s", err)
	}
	err = s.Sink.Clear()
	if err != nil {
		return fmt.Errorf("clearing sink: %s", err)
	}
	for _, credential := range data {
		var path string
		switch l := credential.Location.(type) {
		case PipelinePath:
			path = s.Sink.PipelinePath(l)
		case TeamPath:
			path = s.Sink.TeamPath(l)
		case SharedPath:
			path = s.Sink.SharedPath(l)
		}
		switch v := credential.Value.(type) {
		case SimpleValue:
			err := s.Sink.WriteSimple(path, v)
			if err != nil {
				return fmt.Errorf("writing simple secret: %s", err)
			}
		case CompoundValue:
			err := s.Sink.WriteCompound(path, v)
			if err != nil {
				return fmt.Errorf("writing compound secret: %s", err)
			}
		}
	}
	return nil
}

// a sample of what a secret store contains:
// []Credential{
// 	{
// 		Location: TeamPath{team:"main", secret:"secret1"},
// 		Value:    SimpleValue("value"),
// 	},
// 	{
// 		Location: PipelinePath{team:"main",pipeline:"pipeline",secret:"secret2"},
// 		Value:    CompoundValue{"foo":"bar","baz":"qux"},
// 	}
// }
type Data = []Credential
type Credential struct {
	Location interface{}
	Value    interface{}
}

type PipelinePath struct {
	Team     string
	Pipeline string
	Secret   string
}
type TeamPath struct {
	Team   string
	Secret string
}
type SharedPath struct {
	Secret string
}

type SimpleValue string
type CompoundValue map[string]interface{}
