package secretsyncer

import (
	"errors"
	"fmt"
	"strings"

	"gopkg.in/yaml.v2"
)

// yaml
// shared:
//   name1: simple_secret
//   name2: {complex: multi, field: secret}
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
		segments := strings.Split(k, "/")
		switch len(segments) {
		case 1:
			if k != "shared" {
				return nil, errors.New("top-level key must be a location or 'shared'")
			}
			sharedCreds, err := parseSharedCreds(v)
			if err != nil {
				return nil, err
			}
			data = append(data, sharedCreds...)
		case 2:
			val, err := parseValue(v)
			if err != nil {
				return nil, err
			}
			data = append(data, Credential{
				Location: TeamPath{
					Team:   segments[0],
					Secret: segments[1],
				},
				Value: val,
			})
		case 3:
			val, err := parseValue(v)
			if err != nil {
				return nil, err
			}
			data = append(data, Credential{
				Location: PipelinePath{
					Team:     segments[0],
					Pipeline: segments[1],
					Secret:   segments[2],
				},
				Value: val,
			})
		default:
			return nil, errors.New("invalid location format: too many forward slashes")
		}
	}
	return data, nil
}

func parseSharedCreds(value interface{}) ([]Credential, error) {
	sharedCreds, ok := value.(map[interface{}]interface{})
	if !ok {
		return nil, fmt.Errorf("shared creds of type '%T' are not allowed", value)
	}
	creds := []Credential{}
	for k, cred := range sharedCreds {
		secretName, ok := k.(string)
		if !ok {
			return nil, fmt.Errorf("secret keys of type '%T' are not allowed", k)
		}
		value, err := parseValue(cred)
		if err != nil {
			return nil, err
		}
		creds = append(creds, Credential{
			Location: SharedPath{Secret: secretName},
			Value:    value,
		})
	}
	return creds, nil
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
