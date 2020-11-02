package main

import "io"

// Value = Simple String | Compound (Dict String Value)
type SimpleValue string
type CompoundValue map[string]interface{}

type Data = []Credential
type Convergeable interface {
	Converge(Data) error
}
type SecretWriter struct {
	SecretSink
}

func (SecretWriter) Clear() error {
	return nil
}
func (SecretWriter) Write(Data) error {
	return nil
}

type SecretSink interface {
	WriteSimple(string, SimpleValue) error
	// WriteCompound(string, CompoundValue) error
	Path(string, string) string
}

type NaiveConvergeable struct {
	SecretWriter
}

func (n NaiveConvergeable) Converge(data Data) error {
	err := n.Clear()
	if err != nil {
		return err
	}
	return n.Write(data)
}

type SecretSource interface {
	Read() (Data, error)
}
type FileSecrets struct {
	io.Reader
}

// yaml
// shared:
// - name1: simple_secret
// - name2: {complex: multi, field: secret}
// main/secret1: value
// main/pipeline/secret2: {foo: bar, baz: qux}
// main/pipeline/secret2: {foo: bar, baz: {deeper: nesting}}

type TeamPath struct {
	team   string
	secret string
}
type SharedPath struct{}

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

func main() {
	// input = getInput()
	// convergeable = getConvergeable()
	// data, _ = input.Read()
	// convergeable.Converge(data)
}

// file -> THIS -> vault
// GCS -> file -> THIS -> k8s -> vault
