package secretsyncer

import (
	"fmt"
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
	var rawData map[string]interface{}
	yaml.Unmarshal(bs.Bytes, &rawData)
	data := []Credential{}
	for k, v := range rawData {
		val, err := parseValue(v)
		if err != nil {
			return nil, err
		}
		data = append(data, Credential{
			Location: parseLocation(k),
			Value:    val,
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

func parseValue(value interface{}) (interface{}, error) {
	switch v := value.(type) {
	case string:
		return SimpleValue(v), nil
	case map[interface{}]interface{}:
		return parseCompound(v)
	default:
		return nil, fmt.Errorf("secret values of type '%T' are not allowed", v)
	}
	return nil, nil
}

func parseCompound(yamlMap map[interface{}]interface{}) (CompoundValue, error) {
	cv := CompoundValue{}
	for key, value := range yamlMap {
		stringKey, ok := key.(string)
		if !ok {
			return nil, fmt.Errorf("secret keys of type '%T' are not allowed", key)
		}
		parsedValue, err := parseValue(value)
		if err != nil {
			return nil, err
		}
		cv[stringKey] = parsedValue
	}
	return cv, nil
}
