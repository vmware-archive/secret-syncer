package secretsyncer

import (
	"fmt"
	"io/ioutil"
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
	// WriteCompound(string, CompoundValue) error
	PipelinePath(PipelinePath) string
}

// TODO what if reading the file fails?
func FileSyncer(secretsFile string) Syncer {
	fileBytes, _ := ioutil.ReadFile(secretsFile)
	return Syncer{
		Source: BytesSource{fileBytes},
		Sink:   VaultSink{},
	}
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

// TODO implement compound values
