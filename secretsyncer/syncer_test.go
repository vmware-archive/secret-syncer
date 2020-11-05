package secretsyncer_test

import (
	"errors"
	"fmt"

	"github.com/concourse/secret-syncer/secretsyncer"
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
	Creds map[string]interface{}
}

func (ts *TestSink) WriteSimple(path string, val secretsyncer.SimpleValue) error {
	if ts.Creds == nil {
		ts.Creds = map[string]interface{}{}
	}
	ts.Creds[path] = val
	return nil
}
func (ts *TestSink) WriteCompound(path string, val secretsyncer.CompoundValue) error {
	if ts.Creds == nil {
		ts.Creds = map[string]interface{}{}
	}
	ts.Creds[path] = val
	return nil
}
func (ts *TestSink) Clear() error {
	ts.Creds = map[string]interface{}{}
	return nil
}
func (ts *TestSink) PipelinePath(pp secretsyncer.PipelinePath) string {
	return pp.Team + "/" + pp.Pipeline + "/" + pp.Secret
}
func (ts *TestSink) TeamPath(tp secretsyncer.TeamPath) string {
	return tp.Team + "/" + tp.Secret
}
func (ts *TestSink) SharedPath(sp secretsyncer.SharedPath) string {
	return sp.Secret
}
func (ts *TestSink) Read(path interface{}) interface{} {
	switch p := path.(type) {
	case secretsyncer.TeamPath:
		return ts.Creds[ts.TeamPath(p)]
	case secretsyncer.PipelinePath:
		return ts.Creds[ts.PipelinePath(p)]
	case secretsyncer.SharedPath:
		return ts.Creds[ts.SharedPath(p)]
	}
	return nil
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
func (es ErroringSink) Clear() error {
	return es
}
func (es ErroringSink) PipelinePath(pp secretsyncer.PipelinePath) string {
	return ""
}
func (es ErroringSink) TeamPath(tp secretsyncer.TeamPath) string {
	return ""
}
func (es ErroringSink) SharedPath(sp secretsyncer.SharedPath) string {
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

	s.EqualError(err, fmt.Sprintf("clearing sink: %s", sinkError.Error()))
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

func (s *SyncerSuite) TestWritesSimpleTeamSecretsFromSourceToSink() {
	source := DummySource{credentials: []secretsyncer.Credential{
		{
			Location: secretsyncer.TeamPath{
				Team:   "team_name",
				Secret: "secret_name",
			},
			Value: secretsyncer.SimpleValue("credential"),
		},
	}}
	sink := &TestSink{}

	secretsyncer.Syncer{Source: source, Sink: sink}.Sync()

	s.Equal(
		secretsyncer.SimpleValue("credential"),
		sink.Read(secretsyncer.TeamPath{
			Team:   "team_name",
			Secret: "secret_name",
		}),
	)
}

func (s *SyncerSuite) TestWritesSimpleSharedSecretsFromSourceToSink() {
	source := DummySource{credentials: []secretsyncer.Credential{
		{
			Location: secretsyncer.SharedPath{
				Secret: "secret_name",
			},
			Value: secretsyncer.SimpleValue("credential"),
		},
	}}
	sink := &TestSink{}

	secretsyncer.Syncer{Source: source, Sink: sink}.Sync()

	s.Equal(
		secretsyncer.SimpleValue("credential"),
		sink.Read(secretsyncer.SharedPath{
			Secret: "secret_name",
		}),
	)
}

func (s *SyncerSuite) TestClearsSecretsBeforeWriting() {
	source := DummySource{credentials: []secretsyncer.Credential{}}
	sink := &TestSink{}
	sink.WriteCompound("team/secret", secretsyncer.CompoundValue{
		"username": "user",
		"password": "pass",
	})

	secretsyncer.Syncer{Source: source, Sink: sink}.Sync()

	s.Equal(
		nil,
		sink.Read(secretsyncer.TeamPath{
			Team:   "team",
			Secret: "secret",
		}),
	)
}
