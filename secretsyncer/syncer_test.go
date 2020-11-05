package secretsyncer_test

import (
	"errors"
	"fmt"

	"github.com/jamieklassen/secret-syncer/secretsyncer"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type SyncerSuite struct {
	suite.Suite
	*require.Assertions
}

type DummySource struct {
	credentials []secretsyncer.Credential
}

func (fs DummySource) Read() (secretsyncer.Data, error) {
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
func (ts *TestSink) WriteCompound(path string, val secretsyncer.CompoundValue) error {
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
	source := DummySource{credentials: []secretsyncer.Credential{
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

	secretsyncer.Syncer{Source: source, Sink: sink}.Sync()

	s.Equal(
		secretsyncer.SimpleValue("credential"),
		sink.Read(secretsyncer.PipelinePath{
			Team:     "team_name",
			Pipeline: "pipeline_name",
			Secret:   "secret_name",
		}),
	)
}

type ErroringSink struct {
	error
}

func (es ErroringSink) WriteSimple(path string, val secretsyncer.SimpleValue) error {
	return es
}
func (es ErroringSink) WriteCompound(path string, val secretsyncer.CompoundValue) error {
	return es
}
func (es ErroringSink) PipelinePath(pp secretsyncer.PipelinePath) string {
	return ""
}

func (s *SyncerSuite) TestFailsOnSecretSinkError() {
	source := DummySource{credentials: []secretsyncer.Credential{
		{
			Location: secretsyncer.PipelinePath{
				Team:     "team_name",
				Pipeline: "pipeline_name",
				Secret:   "secret_name",
			},
			Value: secretsyncer.SimpleValue("credential"),
		},
	}}
	sinkError := errors.New(
		"writing secret value 'credential' to vault path '/concourse/team_name/pipeline_name/secret_name': EOF",
	)
	sink := ErroringSink{sinkError}

	err := secretsyncer.Syncer{Source: source, Sink: sink}.Sync()

	s.EqualError(err, fmt.Sprintf("writing simple secret: %s", sinkError.Error()))
}

func (s *SyncerSuite) TestWritesCompoundPipelineSecretsFromSourceToSink() {
	source := DummySource{credentials: []secretsyncer.Credential{
		{
			Location: secretsyncer.PipelinePath{
				Team:     "team_name",
				Pipeline: "pipeline_name",
				Secret:   "secret_name",
			},
			Value: secretsyncer.CompoundValue{
				"username": "user",
				"password": "pass",
			},
		},
	}}
	sink := &TestSink{}

	secretsyncer.Syncer{Source: source, Sink: sink}.Sync()

	s.Equal(
		secretsyncer.CompoundValue{
			"username": "user",
			"password": "pass",
		},
		sink.Read(secretsyncer.PipelinePath{
			Team:     "team_name",
			Pipeline: "pipeline_name",
			Secret:   "secret_name",
		}),
	)
}
