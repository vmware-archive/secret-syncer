package secretsyncer

import (
	"strings"

	"gopkg.in/yaml.v2"
)

// yaml
// shared:
// - name1: simple_secret
// - name2: {complex: multi, field: secret}
// main/secret1: value
// main/pipeline/secret2: {foo: bar, baz: qux}
// main/pipeline/secret2: {foo: bar, baz: {deeper: nesting}}
type BytesSource struct {
	Bytes []byte
}

func (bs BytesSource) Read() (Data, error) {
	var value map[string]interface{}
	yaml.Unmarshal(bs.Bytes, &value)
	data := []Credential{}
	for k, v := range value {
		data = append(data, Credential{
			Location: parseLocation(k),
			Value:    parseValue(v),
		})
	}
	return data, nil
}

func parseLocation(key string) interface{} {
	segments := strings.Split(key, "/")
	return PipelinePath{
		Team:     segments[0],
		Pipeline: segments[1],
		Secret:   segments[2],
	}
}

// TODO error on non simple/compound stuff
func parseValue(value interface{}) interface{} {
	switch v := value.(type) {
	case string:
		return SimpleValue(v)
	}
	return nil
}
