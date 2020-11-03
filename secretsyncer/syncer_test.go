package secretsyncer_test

import (
	"github.com/jamieklassen/secret-syncer/secretsyncer"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type SyncerSuite struct {
	suite.Suite
	*require.Assertions
}

type FakeSource struct {
	credentials []secretsyncer.Credential
}

func (fs FakeSource) Read() (secretsyncer.Data, error) {
	return fs.credentials, nil
}

type TestSink struct {
	creds map[string]interface{}
}

func (ts *TestSink) WriteSimple(path string, val secretsyncer.SimpleValue) error {
	if ts.creds == nil {
		ts.creds = map[string]interface{}{}
	}
	ts.creds[path] = val
	return nil
}
func (ts *TestSink) PipelinePath(pp secretsyncer.PipelinePath) string {
	return pp.Team + "/" + pp.Pipeline + "/" + pp.Secret
}
func (ts *TestSink) Read(pp secretsyncer.PipelinePath) interface{} {
	return ts.creds[ts.PipelinePath(pp)]
}

func (s *SyncerSuite) TestWritesSimplePipelineSecretsFromSourceToSink() {
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
	sink := &TestSink{}
	syncer := secretsyncer.Syncer{Source: source, Sink: sink}

	syncer.Sync()

	s.Equal(
		secretsyncer.SimpleValue("credential"),
		sink.Read(secretsyncer.PipelinePath{
			Team:     "team_name",
			Pipeline: "pipeline_name",
			Secret:   "secret_name",
		}),
	)
}
