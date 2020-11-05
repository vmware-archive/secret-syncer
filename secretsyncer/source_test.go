package secretsyncer_test

import (
	"github.com/jamieklassen/secret-syncer/secretsyncer"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type SourceSuite struct {
	suite.Suite
	*require.Assertions
}

func (s *SourceSuite) TestReadsSimplePipelineSecret() {
	fileBytes := []byte(`team_name/pipeline_name/secret_name: credential`)
	actual, _ := secretsyncer.BytesSource{fileBytes}.Read()
	s.Equal(
		[]secretsyncer.Credential{
			{
				Location: secretsyncer.PipelinePath{
					Team:     "team_name",
					Pipeline: "pipeline_name",
					Secret:   "secret_name",
				},
				Value: secretsyncer.SimpleValue("credential"),
			},
		},
		actual,
	)
}

func (s *SourceSuite) TestFailsOnWrongValueType() {
	fileBytes := []byte(`team_name/pipeline_name/secret_name: []`)
	_, err := secretsyncer.BytesSource{fileBytes}.Read()
	s.EqualError(err, "secret values of type '[]interface {}' are not allowed")
}

func (s *SourceSuite) TestReadsCompoundPipelineSecret() {
	fileBytes := []byte(`team_name/pipeline_name/secret_name:
  username: user
  password: pass
`)
	actual, _ := secretsyncer.BytesSource{fileBytes}.Read()
	s.Equal(
		[]secretsyncer.Credential{
			{
				Location: secretsyncer.PipelinePath{
					Team:     "team_name",
					Pipeline: "pipeline_name",
					Secret:   "secret_name",
				},
				Value: secretsyncer.CompoundValue{
					"username": secretsyncer.SimpleValue("user"),
					"password": secretsyncer.SimpleValue("pass"),
				},
			},
		},
		actual,
	)
}

func (s *SourceSuite) TestFailsOnCompoundSecretWithNumberKeys() {
	fileBytes := []byte(`team_name/pipeline_name/secret_name: {1: foo}`)
	_, err := secretsyncer.BytesSource{fileBytes}.Read()
	s.EqualError(err, "secret keys of type 'int' are not allowed")
}

func (s *SourceSuite) TestFailsOnCompoundSecretWithWrongValueType() {
	fileBytes := []byte(`team_name/pipeline_name/secret_name: {foo: []}`)
	_, err := secretsyncer.BytesSource{fileBytes}.Read()
	s.EqualError(err, "secret values of type '[]interface {}' are not allowed")
}
