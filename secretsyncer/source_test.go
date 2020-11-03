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
