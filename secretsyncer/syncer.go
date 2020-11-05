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
	PipelinePath(PipelinePath) string
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
		Sink:   &VaultSink{vaultClient{client}},
	}, nil
}

func (s Syncer) Sync() error {
	data, _ := s.Source.Read()
	for _, credential := range data {
		var path string
		switch l := credential.Location.(type) {
		case PipelinePath:
			path = s.Sink.PipelinePath(l)
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

// TODO implement team paths and shared paths

type SimpleValue string
type CompoundValue map[string]interface{}
