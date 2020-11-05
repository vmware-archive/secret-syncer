package secretsyncer_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestSuite(t *testing.T) {
	suite.Run(t, &SyncerSuite{
		Assertions: require.New(t),
	})
	suite.Run(t, &SourceSuite{
		Assertions: require.New(t),
	})
	suite.Run(t, &VaultSinkSuite{
		Assertions: require.New(t),
	})
}
