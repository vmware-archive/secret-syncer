package secretsyncer_test

import (
	"testing"

	"github.com/jamieklassen/secret-syncer/secretsyncer"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestSuite(t *testing.T) {
	suite.Run(t, &SyncerSuite{
		Assertions: require.New(t),
	})
}

type SyncerSuite struct {
	suite.Suite
	*require.Assertions
}

type FakeSource struct {
	credentials []secretsyncer.Credential
}

type TestSink struct{}

func (s *SyncerSuite) TestWritesSimplePipelineSecretsFromSourceToSink() {
	// FakeSource
	source := FakeSource{credentials: []secretsyncer.Credential{
		{
			Location: secretsyncer.PipelinePath{
				Team:     "team_name",
				Pipeline: "pipeline_name",
				Secret:   "secret_name",
			},
			Value: secretsyncer.SimpleValue("credential"),
		},
	}}
	// TestSink
	sink := TestSink{}
	syncer := secretsyncer.Syncer{Source: source, Sink: sink}

	syncer.Sync()

	actual, err := sink.Read(secretsyncer.PipelinePath{
		Team:     "team_name",
		Pipeline: "pipeline_name",
		Secret:   "secret_name",
	})
	s.NoError(err)
	s.Equal(secretsyncer.SimpleValue("credential"), actual)
}
